// Copyright (c) 2025 Justin Cranford

// Package secret_content validates that all non-unseal secret files across
// deployments/ have correct content format per ENG-HANDBOOK.md Section 13.3.
//
// Rules are defined in secret-schemas.yaml in this directory. The linter
// loads the schema at runtime, expands {PREFIX}/{PREFIX_US}/{B64URL43}
// placeholders per deployment, and validates each secret file.
//
// Infrastructure deployments (shared-*) are skipped.
// Unseal secrets are validated by the unseal-secret-content linter.
package secret_content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// infraPrefix identifies infrastructure deployment directories to skip.
const infraPrefix = "shared-"

// tierName constants align with the tier strings used in secret-schemas.yaml.
const (
	tierNameService = "service"
	tierNameProduct = "product"
	tierNameSuite   = "suite"
)

// NeverMarkerProduct is the required content for .secret.never files at product tier.
// Exported for use in tests and the secret-schemas.yaml schema.
const NeverMarkerProduct = "MUST NEVER be used at product level. Use service-specific secrets."

// NeverMarkerSuite is the required content for .secret.never files at suite tier.
// Exported for use in tests and the secret-schemas.yaml schema.
const NeverMarkerSuite = "MUST NEVER be used at suite level. Use service-specific secrets."

// tierKind classifies a deployment directory.
type tierKind int

const (
	tierService tierKind = iota
	tierProduct
	tierSuite
)

// deploymentInfo holds resolved tier classification and prefix for a deployment.
type deploymentInfo struct {
	tier     tierKind
	tierName string // "service", "product", or "suite" — matches schema tier values
	prefix   string // Hyphenated prefix (e.g., "jose-ja").
	prefixUS string // Underscored prefix for PostgreSQL identifiers (e.g., "jose_ja").
}

// Check validates non-unseal secret content from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates non-unseal secret content under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking non-unseal secret content...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check secret content: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("secret content violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("secret-content: all non-unseal secrets have valid content")

	return nil
}

// FindViolationsInDir scans deployments/ under rootDir for secret content violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
	schemas, err := LoadSecretSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to load secret schemas: %w", err)
	}

	deploymentsDir := filepath.Join(rootDir, "deployments")

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployments/ directory: %w", err)
	}

	tierMap := buildTierMap()

	var violations []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		deploymentName := entry.Name()

		// Skip infrastructure deployments.
		if strings.HasPrefix(deploymentName, infraPrefix) {
			continue
		}

		secretsDir := filepath.Join(deploymentsDir, deploymentName, "secrets")
		if _, statErr := os.Stat(secretsDir); os.IsNotExist(statErr) {
			continue
		}

		info, exists := tierMap[deploymentName]
		if !exists {
			continue
		}

		v := validateDeploymentSecrets(secretsDir, deploymentName, info, schemas)
		violations = append(violations, v...)
	}

	return violations, nil
}

// buildTierMap returns a mapping of deployment directory names to their tier
// classification and prefix resolution. Uses the entity registry as the source
// of truth for PSIDs, product IDs, and suite IDs.
func buildTierMap() map[string]deploymentInfo {
	tierMap := make(map[string]deploymentInfo)

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		tierMap[ps.PSID] = deploymentInfo{
			tier:     tierService,
			tierName: tierNameService,
			prefix:   ps.PSID,
			prefixUS: strings.ReplaceAll(ps.PSID, "-", "_"),
		}
	}

	for _, prod := range cryptoutilRegistry.AllProducts() {
		tierMap[prod.ID] = deploymentInfo{
			tier:     tierProduct,
			tierName: tierNameProduct,
			prefix:   prod.ID,
			prefixUS: strings.ReplaceAll(prod.ID, "-", "_"),
		}
	}

	for _, suite := range cryptoutilRegistry.AllSuites() {
		tierMap[suite.ID] = deploymentInfo{
			tier:     tierSuite,
			tierName: tierNameSuite,
			prefix:   suite.ID,
			prefixUS: strings.ReplaceAll(suite.ID, "-", "_"),
		}
	}

	return tierMap
}

// validateDeploymentSecrets validates all non-unseal secret files in a deployment
// using the schema-driven rules from secret-schemas.yaml.
func validateDeploymentSecrets(secretsDir, deploymentName string, info deploymentInfo, schemas SecretSchemas) []string {
	var violations []string

	for _, rule := range schemas.ForTier(info.tierName) {
		filePath := filepath.Join(secretsDir, rule.Filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Missing files validated by deployment-dir-completeness linter.
			}

			violations = append(violations, fmt.Sprintf("%s: failed to read %s: %v", deploymentName, rule.Filename, err))

			continue
		}

		line := strings.TrimSpace(string(content))

		if rule.NeverContent != "" {
			if line != rule.NeverContent {
				violations = append(violations, fmt.Sprintf("%s/%s: expected %q, got %q",
					deploymentName, rule.Filename, rule.NeverContent, line))
			}

			continue
		}

		// Regular .secret file: validate against expanded regex.
		if line == "" {
			violations = append(violations, fmt.Sprintf("%s/%s: secret file is empty",
				deploymentName, rule.Filename))

			continue
		}

		expandedPattern := ExpandPattern(rule.ValuePattern, regexp.QuoteMeta(info.prefix), regexp.QuoteMeta(info.prefixUS))

		re := regexp.MustCompile(expandedPattern)
		if !re.MatchString(line) {
			violations = append(violations, fmt.Sprintf("%s/%s: content does not match expected pattern %s: %q",
				deploymentName, rule.Filename, expandedPattern, line))
		}
	}

	return violations
}
