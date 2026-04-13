// Copyright (c) 2025 Justin Cranford

// Package secrets_compliance validates that all deployments/ secrets directories
// contain the expected set of secret files per ENG-HANDBOOK.md Section 13.3
// and the Decision 14 template structure.
//
// Rules:
//  1. All 14 expected .secret files exist in each PS-ID secrets directory.
//  2. Secret values use correct PS-ID prefix format (enforced by secret-content linter).
//  3. BASE64_CHAR43 positions contain values >= 43 characters (enforced by secret-content linter).
//  4. No extra unexpected .secret files allowed.
//  5. Product/suite levels use .secret.never markers where required (enforced by secret-content linter).
//
// This linter focuses on structural completeness (file presence), complementing
// secret-content and unseal-secret-content linters that check file values.
package secrets_compliance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// expectedPSIDSecrets lists all 15 expected .secret files in a PS-ID secrets directory.
var expectedPSIDSecrets = []string{
	"unseal-1of5.secret",
	"unseal-2of5.secret",
	"unseal-3of5.secret",
	"unseal-4of5.secret",
	"unseal-5of5.secret",
	"hash-pepper-v3.secret",
	"postgres-url.secret",
	"postgres-username.secret",
	"postgres-password.secret",
	"postgres-database.secret",
	"browser-username.secret",
	"browser-password.secret",
	"service-username.secret",
	"service-password.secret",
	"issuing-ca-key.secret",
}

// expectedProductSecrets lists all 15 expected secret/marker files in a product-level secrets directory.
// browser/service/issuing-ca use .secret.never marker extension at product/suite level.
var expectedProductSecrets = []string{
	"unseal-1of5.secret",
	"unseal-2of5.secret",
	"unseal-3of5.secret",
	"unseal-4of5.secret",
	"unseal-5of5.secret",
	"hash-pepper-v3.secret",
	"postgres-url.secret",
	"postgres-username.secret",
	"postgres-password.secret",
	"postgres-database.secret",
	"browser-username.secret.never",
	"browser-password.secret.never",
	"service-username.secret.never",
	"service-password.secret.never",
	"issuing-ca-key.secret.never",
}

// expectedSuiteSecrets lists all 15 expected secret/marker files in a suite-level secrets directory.
// Mirrors product secrets — same .secret.never pattern.
var expectedSuiteSecrets = []string{
	"unseal-1of5.secret",
	"unseal-2of5.secret",
	"unseal-3of5.secret",
	"unseal-4of5.secret",
	"unseal-5of5.secret",
	"hash-pepper-v3.secret",
	"postgres-url.secret",
	"postgres-username.secret",
	"postgres-password.secret",
	"postgres-database.secret",
	"browser-username.secret.never",
	"browser-password.secret.never",
	"service-username.secret.never",
	"service-password.secret.never",
	"issuing-ca-key.secret.never",
}

// secretsComplianceFn is the function signature for seam injection in tests.
type secretsComplianceFn func(rootDir string) ([]string, error)

// Check validates secrets directory compliance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkInDir(logger, ".", defaultComplianceFn)
}

// checkInDir validates secrets directory compliance under rootDir.
// Uses seam injection for testability.
func checkInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, fn secretsComplianceFn) error {
	logger.Log("Checking secrets directory compliance...")

	violations, err := fn(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check secrets compliance: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("secrets-compliance violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("secrets-compliance: all deployment secrets directories have expected files")

	return nil
}

// defaultComplianceFn runs the standard structural compliance check.
func defaultComplianceFn(rootDir string) ([]string, error) {
	return findViolations(rootDir)
}

// findViolations checks all deployment secrets directories for structural compliance.
func findViolations(rootDir string) ([]string, error) {
	var violations []string

	// Check PS-ID level secrets directories.
	for _, ps := range cryptoutilRegistry.AllProductServices() {
		secretsDir := filepath.Join(rootDir, "deployments", ps.PSID, "secrets")
		v := checkSecretsDir(secretsDir, ps.PSID, expectedPSIDSecrets)
		violations = append(violations, v...)
	}

	// Check product level secrets directories.
	for _, prod := range cryptoutilRegistry.AllProducts() {
		secretsDir := filepath.Join(rootDir, "deployments", prod.ID, "secrets")
		v := checkSecretsDir(secretsDir, prod.ID, expectedProductSecrets)
		violations = append(violations, v...)
	}

	// Check suite level secrets directories.
	for _, suite := range cryptoutilRegistry.AllSuites() {
		secretsDir := filepath.Join(rootDir, "deployments", suite.ID, "secrets")
		v := checkSecretsDir(secretsDir, suite.ID, expectedSuiteSecrets)
		violations = append(violations, v...)
	}

	return violations, nil
}

// checkSecretsDir validates that a specific secrets directory has all expected files
// and no unexpected .secret files.
func checkSecretsDir(secretsDir, deploymentName string, expectedFiles []string) []string {
	var violations []string

	// Check directory exists.
	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s/secrets: directory does not exist", deploymentName)}
	}

	// Check all expected files are present.
	for _, expected := range expectedFiles {
		filePath := filepath.Join(secretsDir, expected)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s/secrets/%s: missing expected secret file", deploymentName, expected))
		}
	}

	// Check for unexpected .secret files (extra files not in expected list).
	entries, err := os.ReadDir(secretsDir)
	if err != nil {
		return append(violations, fmt.Sprintf("%s/secrets: failed to read directory: %v", deploymentName, err))
	}

	expectedSet := make(map[string]bool, len(expectedFiles))
	for _, f := range expectedFiles {
		expectedSet[f] = true
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only flag .secret and .secret.never files — README and other non-secret files are allowed.
		if !strings.HasSuffix(name, ".secret") && !strings.HasSuffix(name, ".secret.never") {
			continue
		}

		if !expectedSet[name] {
			violations = append(violations, fmt.Sprintf("%s/secrets/%s: unexpected secret file (not in expected list)", deploymentName, name))
		}
	}

	return violations
}
