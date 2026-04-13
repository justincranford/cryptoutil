// Copyright (c) 2025 Justin Cranford

// Package deployment_dir_completeness validates that every product-service in the
// canonical entity registry has all required files under deployments/{PS-ID}/:
//   - Dockerfile
//   - compose.yml
//   - secrets/ directory
//   - config/ directory with 10 config files (5 framework + 5 domain):
//   - {PS-ID}-app-framework-common.yml
//   - {PS-ID}-app-framework-sqlite-1.yml
//   - {PS-ID}-app-framework-sqlite-2.yml
//   - {PS-ID}-app-framework-postgresql-1.yml
//   - {PS-ID}-app-framework-postgresql-2.yml
//   - {PS-ID}-app-domain-common.yml
//   - {PS-ID}-app-domain-sqlite-1.yml
//   - {PS-ID}-app-domain-sqlite-2.yml
//   - {PS-ID}-app-domain-postgresql-1.yml
//   - {PS-ID}-app-domain-postgresql-2.yml
//
// This check prevents deployment drift: if a service is added to the registry
// but its deployment files are incomplete, the check fails.
package deployment_dir_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// Check validates deployment directory completeness from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates deployment directory completeness under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking deployment directory completeness...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkDeploymentDir(rootDir, ps.PSID)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("deployment directory completeness violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("deployment-dir-completeness: all 10 product-services have required deployment files")

	return nil
}

// checkDeploymentDir verifies all required files exist for a deployment directory.
func checkDeploymentDir(rootDir, psID string) []string {
	var violations []string

	deployDir := filepath.Join(rootDir, "deployments", psID)

	// 1. Dockerfile
	if _, err := os.Stat(filepath.Join(deployDir, "Dockerfile")); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing deployments/%s/Dockerfile", psID, psID))
	}

	// 2. compose.yml
	if _, err := os.Stat(filepath.Join(deployDir, "compose.yml")); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing deployments/%s/compose.yml", psID, psID))
	}

	// 3. secrets/ directory
	if _, err := os.Stat(filepath.Join(deployDir, "secrets")); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing deployments/%s/secrets/ directory", psID, psID))
	}

	// 4. config/ directory
	configDir := filepath.Join(deployDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing deployments/%s/config/ directory", psID, psID))

		return violations
	}

	// 5. 10 required config files under config/ (5 framework + 5 domain)
	configFiles := []string{
		psID + "-app-framework-common.yml",
		psID + "-app-framework-sqlite-1.yml",
		psID + "-app-framework-sqlite-2.yml",
		psID + "-app-framework-postgresql-1.yml",
		psID + "-app-framework-postgresql-2.yml",
		psID + "-app-domain-common.yml",
		psID + "-app-domain-sqlite-1.yml",
		psID + "-app-domain-sqlite-2.yml",
		psID + "-app-domain-postgresql-1.yml",
		psID + "-app-domain-postgresql-2.yml",
	}

	for _, cfgFile := range configFiles {
		if _, err := os.Stat(filepath.Join(configDir, cfgFile)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing deployments/%s/config/%s", psID, psID, cfgFile))
		}
	}

	return violations
}
