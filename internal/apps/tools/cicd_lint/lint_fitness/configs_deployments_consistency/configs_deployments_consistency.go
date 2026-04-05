// Copyright (c) 2025 Justin Cranford

// Package configs_deployments_consistency validates that every deployments/{PS-ID}/
// directory has a matching configs/{PS-ID}/ directory. PS-IDs are validated against
// the entity registry.
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

// Check validates configs/deployments consistency from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.Stat, os.ReadDir)
}

// CheckInDir validates configs/deployments consistency under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, statFn func(string) (os.FileInfo, error), readDirFn func(string) ([]os.DirEntry, error)) error {
	logger.Log("Checking configs/deployments consistency...")

	violations, err := FindViolationsInDir(rootDir, statFn, readDirFn)
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
// a matching configs/{PS-ID}/ directory.
func FindViolationsInDir(rootDir string, statFn func(string) (os.FileInfo, error), readDirFn func(string) ([]os.DirEntry, error)) ([]string, error) {
	deploymentsDir := filepath.Join(rootDir, "deployments")

	if _, err := statFn(deploymentsDir); err != nil {
		return nil, fmt.Errorf("deployments/ directory not found: %w", err)
	}

	entries, err := readDirFn(deploymentsDir)
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

		if _, ok := psMap[psID]; !ok {
			continue // Not a registered PS-ID; other linters handle unknown dirs.
		}

		configsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir, psID)

		if _, err := statFn(configsDir); err != nil {
			violations = append(violations, fmt.Sprintf("deployments/%s/ exists but configs/%s/ is missing", psID, psID))
		}
	}

	return violations, nil
}

// buildPSIDMap returns a set of all registered PS-IDs.
func buildPSIDMap() map[string]bool {
	result := make(map[string]bool)

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		result[ps.PSID] = true
	}

	return result
}

// FormatViolations formats violations as a newline-separated string.
func FormatViolations(violations []string) string {
	return strings.Join(violations, "\n")
}
