// Copyright (c) 2025 Justin Cranford

// Package entity_registry_completeness validates that every product-service in the
// canonical entity registry has the required structural components on disk:
//   - deployments/{PS-ID}/ directory
//   - configs/{PRODUCT}/{SERVICE}/ directory
//   - internal/shared/magic/{MagicFile} file
//
// This check prevents structural drift: if a product-service is added to the registry
// but its deployment, config, or magic constants file is missing, the check fails.
package entity_registry_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// Check validates entity registry completeness from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates entity registry completeness under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking entity registry completeness...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkProductService(rootDir, ps)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("entity registry completeness violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("entity-registry-completeness: all 10 product-services have required components")

	return nil
}

// checkProductService verifies the three required components exist for a product-service.
func checkProductService(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	var violations []string

	// 1. deployments/{PS-ID}/ directory.
	deploymentsDir := filepath.Join(rootDir, "deployments", ps.PSID)
	if _, err := os.Stat(deploymentsDir); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing deployments/%s/ directory", ps.PSID, ps.PSID))
	}

	// 2. configs/{PRODUCT}/{SERVICE}/ directory.
	configsDir := filepath.Join(rootDir, "configs", ps.Product, ps.Service)
	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing configs/%s/%s/ directory", ps.PSID, ps.Product, ps.Service))
	}

	// 3. internal/shared/magic/{MagicFile} file.
	magicFile := filepath.Join(rootDir, "internal", "shared", "magic", ps.MagicFile)
	if _, err := os.Stat(magicFile); os.IsNotExist(err) {
		violations = append(violations, fmt.Sprintf("%s: missing internal/shared/magic/%s", ps.PSID, ps.MagicFile))
	}

	return violations
}
