// Copyright (c) 2025 Justin Cranford

// Package configs_naming validates that the configs/ directory structure follows
// the canonical flat hierarchy:
//
//   - configs/{suite}/                  - suite-level configs (e.g. cryptoutil/)
//   - configs/{PS-ID}/                  - service-level configs (e.g. sm-kms/, jose-ja/)
//
// Top-level directories must be a known suite ID or PS-ID from the entity registry.
// Subdirectories within PS-ID dirs (e.g. profiles/, domain/) are allowed.
package configs_naming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check validates configs/ directory structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates configs/ directory structure under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking configs/ naming and structure...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check configs naming: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("configs/ naming violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("configs-naming: configs/ structure is valid")

	return nil
}

// FindViolationsInDir scans configs/ under rootDir and returns all naming violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
	allowedDirs := buildAllowedDirs()

	configsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir)

	topEntries, err := os.ReadDir(configsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read configs/ directory: %w", err)
	}

	var violations []string

	for _, entry := range topEntries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !allowedDirs[name] {
			violations = append(violations, fmt.Sprintf("configs/%s: unknown directory (not a registered suite ID or PS-ID in entity registry)", name))
		}
	}

	return violations, nil
}

// buildAllowedDirs returns a set of allowed top-level directory names under configs/.
// This includes all suite IDs and all PS-IDs from the entity registry.
func buildAllowedDirs() map[string]bool {
	allowed := make(map[string]bool)

	for _, s := range lintFitnessRegistry.AllSuites() {
		allowed[s.ID] = true
	}

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		allowed[ps.PSID] = true
	}

	return allowed
}
