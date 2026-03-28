// Copyright (c) 2025 Justin Cranford

// Package secret_content validates that all non-unseal secret files across
// deployments/ have correct content format per ARCHITECTURE.md Section 13.3:
//
//   - hash-pepper-v3.secret: {PREFIX}-hash-pepper-v3-{base64url-43}
//   - browser-username.secret: {PREFIX}-browser-user (service only)
//   - browser-password.secret: {PREFIX}-browser-pass-{base64url-43} (service only)
//   - service-username.secret: {PREFIX}-service-user (service only)
//   - service-password.secret: {PREFIX}-service-pass-{base64url-43} (service only)
//   - postgres-database.secret: {PREFIX_US}_database
//   - postgres-username.secret: {PREFIX_US}_database_user
//   - postgres-password.secret: {PREFIX_US}_database_pass-{base64url-43}
//   - postgres-url.secret: composed from above three postgres values
//   - .secret.never markers at product/suite tiers contain required text
//
// Infrastructure deployments (shared-postgres, shared-telemetry) are skipped.
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

// base64URLPattern matches exactly 43 base64url characters (no padding).
var base64URLPattern = `[A-Za-z0-9_-]{43}`

// infraPrefix identifies infrastructure deployment directories to skip.
var infraPrefix = "shared-"

// tierKind classifies a deployment directory.
type tierKind int

const (
	tierService tierKind = iota
	tierProduct
	tierSuite
)

// neverMarkerProduct is the required content for .secret.never files at product tier.
var neverMarkerProduct = "MUST NEVER be used at product level. Use service-specific secrets."

// neverMarkerSuite is the required content for .secret.never files at suite tier.
var neverMarkerSuite = "MUST NEVER be used at suite level. Use service-specific secrets."

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

		v := validateDeploymentSecrets(secretsDir, deploymentName, info)
		violations = append(violations, v...)
	}

	return violations, nil
}

// deploymentInfo holds resolved tier classification and prefix for a deployment.
type deploymentInfo struct {
	tier     tierKind
	prefix   string // Hyphenated prefix (e.g., "jose-ja").
	prefixUS string // Underscored prefix for PostgreSQL identifiers (e.g., "jose_ja").
}

// buildTierMap returns a mapping of deployment directory names to their tier
// classification and prefix resolution. Uses the entity registry as the source
// of truth for PSIDs, product IDs, and suite IDs.
func buildTierMap() map[string]deploymentInfo {
	tierMap := make(map[string]deploymentInfo)

	for _, ps := range cryptoutilRegistry.AllProductServices() {
		tierMap[ps.PSID] = deploymentInfo{
			tier:     tierService,
			prefix:   ps.PSID,
			prefixUS: strings.ReplaceAll(ps.PSID, "-", "_"),
		}
	}

	for _, prod := range cryptoutilRegistry.AllProducts() {
		tierMap[prod.ID] = deploymentInfo{
			tier:     tierProduct,
			prefix:   prod.ID,
			prefixUS: strings.ReplaceAll(prod.ID, "-", "_"),
		}
	}

	for _, suite := range cryptoutilRegistry.AllSuites() {
		tierMap[suite.ID] = deploymentInfo{
			tier:     tierSuite,
			prefix:   suite.ID,
			prefixUS: strings.ReplaceAll(suite.ID, "-", "_"),
		}
	}

	return tierMap
}

// validateDeploymentSecrets validates all non-unseal secret files in a deployment.
func validateDeploymentSecrets(secretsDir, deploymentName string, info deploymentInfo) []string {
	var violations []string

	// Validate standard .secret files present at all tiers.
	violations = append(violations, validateSecret(secretsDir, deploymentName, "hash-pepper-v3.secret",
		fmt.Sprintf("^%s-hash-pepper-v3-%s$", regexp.QuoteMeta(info.prefix), base64URLPattern))...)
	violations = append(violations, validateSecret(secretsDir, deploymentName, "postgres-database.secret",
		fmt.Sprintf("^%s_database$", regexp.QuoteMeta(info.prefixUS)))...)
	violations = append(violations, validateSecret(secretsDir, deploymentName, "postgres-username.secret",
		fmt.Sprintf("^%s_database_user$", regexp.QuoteMeta(info.prefixUS)))...)
	violations = append(violations, validateSecret(secretsDir, deploymentName, "postgres-password.secret",
		fmt.Sprintf("^%s_database_pass-%s$", regexp.QuoteMeta(info.prefixUS), base64URLPattern))...)
	violations = append(violations, validatePostgresURL(secretsDir, deploymentName, info)...)

	// Browser/service credentials: .secret at service tier, .secret.never at product/suite tiers.
	switch info.tier {
	case tierService:
		violations = append(violations, validateSecret(secretsDir, deploymentName, "browser-username.secret",
			fmt.Sprintf("^%s-browser-user$", regexp.QuoteMeta(info.prefix)))...)
		violations = append(violations, validateSecret(secretsDir, deploymentName, "browser-password.secret",
			fmt.Sprintf("^%s-browser-pass-%s$", regexp.QuoteMeta(info.prefix), base64URLPattern))...)
		violations = append(violations, validateSecret(secretsDir, deploymentName, "service-username.secret",
			fmt.Sprintf("^%s-service-user$", regexp.QuoteMeta(info.prefix)))...)
		violations = append(violations, validateSecret(secretsDir, deploymentName, "service-password.secret",
			fmt.Sprintf("^%s-service-pass-%s$", regexp.QuoteMeta(info.prefix), base64URLPattern))...)
	case tierProduct:
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "browser-username.secret.never", neverMarkerProduct)...)
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "browser-password.secret.never", neverMarkerProduct)...)
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "service-username.secret.never", neverMarkerProduct)...)
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "service-password.secret.never", neverMarkerProduct)...)
	case tierSuite:
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "browser-username.secret.never", neverMarkerSuite)...)
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "browser-password.secret.never", neverMarkerSuite)...)
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "service-username.secret.never", neverMarkerSuite)...)
		violations = append(violations, validateNeverMarker(secretsDir, deploymentName, "service-password.secret.never", neverMarkerSuite)...)
	}

	return violations
}

// validateSecret reads a secret file and validates its content against a regex pattern.
func validateSecret(secretsDir, deploymentName, filename, pattern string) []string {
	filePath := filepath.Join(secretsDir, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Missing files are validated by deployment-dir-completeness linter.
		}

		return []string{fmt.Sprintf("%s: failed to read %s: %v", deploymentName, filename, err)}
	}

	line := strings.TrimSpace(string(content))
	if line == "" {
		return []string{fmt.Sprintf("%s/%s: secret file is empty", deploymentName, filename)}
	}

	re := regexp.MustCompile(pattern)
	if !re.MatchString(line) {
		return []string{fmt.Sprintf("%s/%s: content does not match expected pattern %s: %q", deploymentName, filename, pattern, line)}
	}

	return nil
}

// validateNeverMarker reads a .secret.never file and validates its marker content.
func validateNeverMarker(secretsDir, deploymentName, filename, expectedContent string) []string {
	filePath := filepath.Join(secretsDir, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Missing files are validated by deployment-dir-completeness linter.
		}

		return []string{fmt.Sprintf("%s: failed to read %s: %v", deploymentName, filename, err)}
	}

	line := strings.TrimSpace(string(content))
	if line != expectedContent {
		return []string{fmt.Sprintf("%s/%s: expected %q, got %q", deploymentName, filename, expectedContent, line)}
	}

	return nil
}

// validatePostgresURL validates the postgres-url.secret content structure.
func validatePostgresURL(secretsDir, deploymentName string, info deploymentInfo) []string {
	filePath := filepath.Join(secretsDir, "postgres-url.secret")

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return []string{fmt.Sprintf("%s: failed to read postgres-url.secret: %v", deploymentName, err)}
	}

	line := strings.TrimSpace(string(content))
	if line == "" {
		return []string{fmt.Sprintf("%s/postgres-url.secret: secret file is empty", deploymentName)}
	}

	pattern := fmt.Sprintf(
		`^postgres://%s_database_user:%s_database_pass-%s@%s-postgres:5432/%s_database\?sslmode=disable$`,
		regexp.QuoteMeta(info.prefixUS),
		regexp.QuoteMeta(info.prefixUS),
		base64URLPattern,
		regexp.QuoteMeta(info.prefix),
		regexp.QuoteMeta(info.prefixUS),
	)

	re := regexp.MustCompile(pattern)
	if !re.MatchString(line) {
		return []string{fmt.Sprintf("%s/postgres-url.secret: content does not match expected URL pattern: %q", deploymentName, line)}
	}

	return nil
}
