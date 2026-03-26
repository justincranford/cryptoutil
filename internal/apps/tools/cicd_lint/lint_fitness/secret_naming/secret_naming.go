// Copyright (c) 2025 Justin Cranford

// Package secret_naming validates that all secret files under deployments/
// follow naming conventions:
//
//   - All filenames use hyphens (no underscores)
//   - All files have .secret or .secret.never extension
//   - File names are lowercase
package secret_naming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// Check validates secret naming from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates secret naming under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking secret file naming conventions...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check secret naming: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("secret naming violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("secret-naming: all secret files follow naming conventions")

	return nil
}

// FindViolationsInDir scans deployments/ under rootDir for secret naming violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
	deploymentsDir := filepath.Join(rootDir, "deployments")

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployments/ directory: %w", err)
	}

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

		v := validateSecretsDir(secretsDir, deploymentName)
		violations = append(violations, v...)
	}

	return violations, nil
}

// validateSecretsDir checks all files in a deployment's secrets/ directory.
func validateSecretsDir(secretsDir, deploymentName string) []string {
	var violations []string

	entries, err := os.ReadDir(secretsDir)
	if err != nil {
		violations = append(violations, fmt.Sprintf("%s: failed to read secrets/ directory: %v", deploymentName, err))

		return violations
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Check for underscores in filename.
		if strings.Contains(filename, "_") {
			violations = append(violations, fmt.Sprintf("%s/secrets/%s: filename contains underscore (use hyphens)", deploymentName, filename))
		}

		// Check for correct extension.
		if !strings.HasSuffix(filename, ".secret") && !strings.HasSuffix(filename, ".secret.never") {
			violations = append(violations, fmt.Sprintf("%s/secrets/%s: missing .secret or .secret.never extension", deploymentName, filename))
		}

		// Check for lowercase.
		if filename != strings.ToLower(filename) {
			violations = append(violations, fmt.Sprintf("%s/secrets/%s: filename must be lowercase", deploymentName, filename))
		}
	}

	return violations
}
