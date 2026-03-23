// Copyright (c) 2025 Justin Cranford

// Package configs_deployments_consistency validates that every deployments/{PS-ID}/
// directory has a matching configs/{PRODUCT}/{SERVICE}/ directory, using the entity
// registry to map PS-ID to PRODUCT/SERVICE.
package configs_deployments_consistency

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Seam variables for test injection.
var (
	configsDeploymentsStatFn    = os.Stat
	configsDeploymentsReadDirFn = os.ReadDir
)

// Check validates configs/deployments consistency from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates configs/deployments consistency under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking configs/deployments consistency...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check configs/deployments consistency: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  configs/deployments inconsistency: %s\n", v)
		}

		return fmt.Errorf("found %d configs/deployments inconsistencies", len(violations))
	}

	logger.Log("configs-deployments-consistency: all deployments have matching configs")

	return nil
}

// FindViolationsInDir scans deployments/ under rootDir and verifies each PS-ID has
// a matching configs/{PRODUCT}/{SERVICE}/ directory.
func FindViolationsInDir(rootDir string) ([]string, error) {
	deploymentsDir := filepath.Join(rootDir, "deployments")

	if _, err := configsDeploymentsStatFn(deploymentsDir); err != nil {
		return nil, fmt.Errorf("deployments/ directory not found: %w", err)
	}

	entries, err := configsDeploymentsReadDirFn(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployments/ directory: %w", err)
	}

	psMap := buildPSIDMap()

	var violations []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		psID := entry.Name()

		ps, ok := psMap[psID]
		if !ok {
			continue // Not a registered PS-ID; other linters handle unknown dirs.
		}

		configsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir, ps.Product, ps.Service)

		if _, err := configsDeploymentsStatFn(configsDir); err != nil {
			violations = append(violations, fmt.Sprintf("deployments/%s/ exists but configs/%s/%s/ is missing", psID, ps.Product, ps.Service))
		}
	}

	return violations, nil
}

// buildPSIDMap returns a map from PS-ID to ProductService.
func buildPSIDMap() map[string]lintFitnessRegistry.ProductService {
	result := make(map[string]lintFitnessRegistry.ProductService)

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		result[ps.PSID] = ps
	}

	return result
}

// FormatViolations formats violations as a newline-separated string.
func FormatViolations(violations []string) string {
	return strings.Join(violations, "\n")
}
