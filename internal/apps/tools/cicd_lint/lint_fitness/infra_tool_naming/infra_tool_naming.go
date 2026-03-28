// Copyright (c) 2025 Justin Cranford

// Package infra_tool_naming enforces that infrastructure tool cmd/ directories
// follow the cicd- prefix convention and have matching internal/apps/tools/
// counterparts (ARCHITECTURE.md Section 4.4.7 CLI Patterns).
package infra_tool_naming

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check runs the infra-tool-naming check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates that all non-registry cmd/ entries follow the cicd- prefix
// naming convention and have matching internal/apps/tools/ counterpart directories.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking infrastructure tool naming conventions...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check infra-tool naming: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			logger.Log(fmt.Sprintf("  VIOLATION: %s", v))
		}

		return fmt.Errorf("infra-tool-naming: found %d violation(s)", len(violations))
	}

	logger.Log("infra-tool-naming: all infrastructure tool directories follow naming conventions")

	return nil
}

// FindViolationsInDir scans cmd/ entries that are not in the product/service/suite
// registry, validates they use the cicd- prefix, and checks for matching
// internal/apps/tools/ counterparts.
func FindViolationsInDir(rootDir string) ([]string, error) {
	registryEntries := buildRegistrySet()

	cmdDir := filepath.Join(rootDir, "cmd")

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cmd/ directory: %w", err)
	}

	var violations []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip entries that are registered products/services/suites.
		if registryEntries[name] {
			continue
		}

		// This is an infrastructure tool — validate prefix.
		if !strings.HasPrefix(name, cryptoutilSharedMagic.CICDInfraToolCmdPrefix) {
			violations = append(violations, fmt.Sprintf("cmd/%s: infrastructure tool MUST be prefixed with %q", name, cryptoutilSharedMagic.CICDInfraToolCmdPrefix))

			continue
		}

		// Validate matching internal/apps/tools/ counterpart exists.
		internalName := strings.ReplaceAll(name, "-", "_")
		internalDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDInfraToolInternalDir, internalName)

		if _, statErr := os.Stat(internalDir); os.IsNotExist(statErr) {
			violations = append(violations, fmt.Sprintf("cmd/%s: missing counterpart %s/%s", name, cryptoutilSharedMagic.CICDInfraToolInternalDir, internalName))
		}
	}

	sort.Strings(violations)

	return violations, nil
}

// buildRegistrySet returns a set of all known product/service/suite cmd/ entry names.
func buildRegistrySet() map[string]bool {
	known := make(map[string]bool)

	for _, suite := range lintFitnessRegistry.AllSuites() {
		known[suite.ID] = true
	}

	for _, product := range lintFitnessRegistry.AllProducts() {
		known[product.ID] = true
	}

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		known[ps.PSID] = true
	}

	return known
}
