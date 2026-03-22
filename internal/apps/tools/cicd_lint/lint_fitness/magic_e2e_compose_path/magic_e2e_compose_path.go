// Copyright (c) 2025 Justin Cranford

// Package magic_e2e_compose_path validates that every *E2EComposeFile magic constant
// in internal/shared/magic/ resolves to an existing compose.yml on disk.
//
// For each magic file, the checker:
//  1. Parses the Go source for constants with suffix "E2EComposeFile".
//  2. Resolves the relative path from the e2e test directory
//     (rootDir/internal/apps/{InternalAppsDir}e2e/<relative-path>).
//  3. Verifies the resolved file exists.
//
// Magic files that contain no *E2EComposeFile constant are skipped.
// When multiple product-services share the same magic file (e.g. identity),
// only one scan is performed per magic file.
package magic_e2e_compose_path

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// composeFileRe matches Go string constant assignments whose name ends in E2EComposeFile.
// Group 1: constant name  Group 2: string value.
var composeFileRe = regexp.MustCompile(`(\w+E2EComposeFile)\s*=\s*"([^"]+)"`)

// Check validates E2E compose file path constants from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates E2E compose file path constants under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking magic E2E compose file path constants...")

	var violations []string

	// Deduplicate by MagicFile so shared magic files (e.g. identity) are scanned once.
	seen := make(map[string]bool)

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		if seen[ps.MagicFile] {
			continue
		}

		seen[ps.MagicFile] = true

		v := checkMagicFile(rootDir, ps)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("magic E2E compose path violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("magic-e2e-compose-path: all E2EComposeFile constants resolve to existing files")

	return nil
}

// checkMagicFile reads the magic file for ps and verifies any E2EComposeFile constants.
// If the file contains no E2EComposeFile constant the PS is skipped.
func checkMagicFile(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	magicPath := filepath.Join(rootDir, "internal", "shared", "magic", ps.MagicFile)

	src, err := os.ReadFile(magicPath)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read %s: %v", ps.PSID, ps.MagicFile, err)}
	}

	matches := composeFileRe.FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		// No E2EComposeFile constant in this magic file — skip.
		return nil
	}

	var violations []string

	// E2E base directory: rootDir/internal/apps/{InternalAppsDir}e2e/
	e2eDir := filepath.Join(rootDir, "internal", "apps", filepath.FromSlash(ps.InternalAppsDir), "e2e")

	for _, m := range matches {
		constName := m[1]
		relPath := m[2]

		// Resolve: e2eDir / relPath (supports ../ traversal naturally via Clean)
		resolved := filepath.Clean(filepath.Join(e2eDir, filepath.FromSlash(relPath)))

		if _, statErr := os.Stat(resolved); os.IsNotExist(statErr) {
			// Make the path in the violation relative to rootDir for readability.
			display := relPath
			if rel, relErr := filepath.Rel(rootDir, resolved); relErr == nil {
				display = rel
			}

			violations = append(violations, fmt.Sprintf("%s: %s: %s = %q resolves to non-existent path %s", ps.PSID, ps.MagicFile, constName, relPath, display))
		}
	}

	return violations
}
