// Copyright (c) 2025 Justin Cranford

// Package unseal_secret_content validates that all unseal secret files across
// deployments/ have correct content format:
//
//   - Pattern: {deployment-name}-unseal-key-{N}-of-5-{64-hex-chars}
//   - The prefix must match the deployment directory name
//   - N must match the filename (unseal-{N}of5.secret → shard N)
//   - Hex must be exactly 64 lowercase hex characters
//   - All hex values must be unique across all shards within a deployment
//   - Generic placeholder values (e.g. dev-unseal-key-N-of-5) are rejected
package unseal_secret_content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// unsealPattern matches: {prefix}-unseal-key-{N}-of-5-{64-hex-chars}.
var unsealPattern = regexp.MustCompile(`^(.+)-unseal-key-(\d+)-of-5-([0-9a-f]{64})$`)

// genericPrefix is the banned placeholder prefix.
var genericPrefix = cryptoutilSharedMagic.UnsealGenericBannedPrefix

// Check validates unseal secret content from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates unseal secret content under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking unseal secret content...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check unseal secret content: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("unseal secret content violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("unseal-secret-content: all unseal secrets have valid content")

	return nil
}

// FindViolationsInDir scans deployments/ under rootDir for unseal secret content violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
	deploymentsDir := filepath.Join(rootDir, "deployments")

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployments/ directory: %w", err)
	}

	prefixMap := buildDeploymentPrefixMap()

	var violations []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		deploymentName := entry.Name()
		secretsDir := filepath.Join(deploymentsDir, deploymentName, "secrets")

		if _, statErr := os.Stat(secretsDir); os.IsNotExist(statErr) {
			continue
		}

		// Resolve the expected unseal prefix for this deployment.
		expectedPrefix := deploymentName
		if mapped, exists := prefixMap[deploymentName]; exists {
			expectedPrefix = mapped
		}

		v := validateDeploymentUnsealSecrets(secretsDir, deploymentName, expectedPrefix)
		violations = append(violations, v...)
	}

	return violations, nil
}

// buildDeploymentPrefixMap returns a mapping from deployment directory names to
// their expected unseal content prefix. Most deployments use the directory name
// as the prefix. Suite deployments (e.g. "cryptoutil-suite") use the suite ID
// (e.g. "cryptoutil") as the prefix instead.
func buildDeploymentPrefixMap() map[string]string {
	prefixMap := make(map[string]string)

	for _, suite := range cryptoutilRegistry.AllSuites() {
		deploymentDir := suite.ID + "-suite"
		prefixMap[deploymentDir] = suite.ID
	}

	return prefixMap
}

// validateDeploymentUnsealSecrets checks all unseal-*.secret files in a deployment's secrets/ dir.
func validateDeploymentUnsealSecrets(secretsDir, deploymentName, expectedPrefix string) []string {
	var violations []string

	hexValues := make(map[string]string)

	for shardNum := 1; shardNum <= cryptoutilSharedMagic.UnsealTotalShards; shardNum++ {
		filename := fmt.Sprintf("unseal-%dof5.secret", shardNum)
		filePath := filepath.Join(secretsDir, filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			violations = append(violations, fmt.Sprintf("%s: failed to read %s: %v", deploymentName, filename, err))

			continue
		}

		line := strings.TrimSpace(string(content))
		if line == "" {
			violations = append(violations, fmt.Sprintf("%s/%s: unseal secret file is empty", deploymentName, filename))

			continue
		}

		v := validateUnsealContent(line, deploymentName, expectedPrefix, filename, shardNum, hexValues)
		violations = append(violations, v...)
	}

	return violations
}

// validateUnsealContent validates a single unseal secret value.
func validateUnsealContent(content, deploymentName, expectedPrefix, filename string, expectedShard int, hexValues map[string]string) []string {
	var violations []string

	matches := unsealPattern.FindStringSubmatch(content)
	if matches == nil {
		violations = append(violations, fmt.Sprintf("%s/%s: content does not match pattern {prefix}-unseal-key-{N}-of-5-{64hex}: %q", deploymentName, filename, content))

		return violations
	}

	prefix := matches[1]
	shardStr := matches[2]
	hexValue := matches[3]

	// Validate prefix matches expected prefix (may differ from deployment name for suites).
	if prefix != expectedPrefix {
		violations = append(violations, fmt.Sprintf("%s/%s: prefix %q does not match expected prefix %q", deploymentName, filename, prefix, expectedPrefix))
	}

	// Validate shard number matches filename.
	expectedShardStr := fmt.Sprintf("%d", expectedShard)
	if shardStr != expectedShardStr {
		violations = append(violations, fmt.Sprintf("%s/%s: shard number %s does not match expected %s from filename", deploymentName, filename, shardStr, expectedShardStr))
	}

	// Validate no generic placeholder prefix.
	if prefix == genericPrefix {
		violations = append(violations, fmt.Sprintf("%s/%s: uses generic %q prefix (must use deployment-specific prefix)", deploymentName, filename, genericPrefix))
	}

	// Validate hex uniqueness within deployment.
	if existingFile, exists := hexValues[hexValue]; exists {
		violations = append(violations, fmt.Sprintf("%s/%s: duplicate hex value (same as %s)", deploymentName, filename, existingFile))
	} else {
		hexValues[hexValue] = filename
	}

	return violations
}
