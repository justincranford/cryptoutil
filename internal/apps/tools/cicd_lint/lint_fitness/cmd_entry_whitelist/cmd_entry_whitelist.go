// Copyright (c) 2025 Justin Cranford

// Package cmd_entry_whitelist enforces that the cmd/ directory contains only the
// 18 allowed entry points: 1 suite, 5 products, 10 product-services, and 2
// infrastructure tools. Any extra or unknown cmd/ directory is a violation
// (ARCHITECTURE.md Section 4.4.7 CLI Patterns).
package cmd_entry_whitelist

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// infraToolCmds lists the 2 allowed infrastructure tool cmd entries.
// These are not in the entity registry since they are tools, not services.
var infraToolCmds = []string{
	cryptoutilSharedMagic.CICDCmdDirCicdLint,
	cryptoutilSharedMagic.CICDCmdDirWorkflow,
}

// Check runs the cmd-entry-whitelist check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks that cmd/ under rootDir contains only the 18 allowed entries.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking cmd/ entry whitelist...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check cmd/ entry whitelist: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  unauthorized cmd/ entry: %s\n", v)
		}

		return fmt.Errorf("cmd-entry-whitelist: found %d unauthorized cmd/ entr%s", len(violations), pluralIes(len(violations)))
	}

	logger.Log("cmd-entry-whitelist: all cmd/ entries are authorized")

	return nil
}

// FindViolationsInDir scans cmd/ under rootDir and returns any unauthorized directory names.
func FindViolationsInDir(rootDir string) ([]string, error) {
	allowed := buildAllowedCmdEntries()

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

		if !allowed[name] {
			violations = append(violations, name)
		}
	}

	sort.Strings(violations)

	return violations, nil
}

// buildAllowedCmdEntries builds the set of allowed cmd/ directory names.
// Includes: suite ID, all product IDs, all PS-IDs, and infra tool names.
func buildAllowedCmdEntries() map[string]bool {
	allowed := make(map[string]bool)

	for _, suite := range lintFitnessRegistry.AllSuites() {
		allowed[suite.ID] = true
	}

	for _, product := range lintFitnessRegistry.AllProducts() {
		allowed[product.ID] = true
	}

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		allowed[ps.PSID] = true
	}

	for _, tool := range infraToolCmds {
		allowed[tool] = true
	}

	return allowed
}

// pluralIes returns "y" for count==1, "ies" otherwise (for "entry"/"entries").
func pluralIes(count int) string {
	if count == 1 {
		return "y"
	}

	return "ies"
}

// AllowedCount returns the total number of allowed cmd/ entries.
func AllowedCount() int {
	return len(buildAllowedCmdEntries())
}

// AllowedEntries returns a sorted slice of all allowed cmd/ entry names.
func AllowedEntries() []string {
	allowed := buildAllowedCmdEntries()
	result := make([]string, 0, len(allowed))

	for name := range allowed {
		result = append(result, name)
	}

	sort.Strings(result)

	return result
}

// AllowedEntrySet returns the set of allowed cmd/ entry names for testing.
func AllowedEntrySet() map[string]bool {
	return buildAllowedCmdEntries()
}
