// Copyright (c) 2025 Justin Cranford

// Package magic_e2e_container_names validates that E2E container name magic constants
// in internal/shared/magic/ match the expected compose service names.
//
// For each product-service that has E2E container name constants in its magic file,
// this check parses the Go source and verifies:
//   - *E2ESQLiteContainer      = "{ps-id}-app-sqlite-1"
//   - *E2EPostgreSQL1Container = "{ps-id}-app-postgresql-1"
//   - *E2EPostgreSQL2Container = "{ps-id}-app-postgresql-2"
//
// Product-services whose magic files do not contain any E2ESQLiteContainer constant
// (e.g. identity, pki-ca) are skipped — not all services have the 3-tuple pattern.
//
// This prevents magic constants from drifting out of sync with compose service names.
package magic_e2e_container_names

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// containerCheck describes one E2E container name constant to verify.
type containerCheck struct {
	constSuffix    string // suffix of the Go constant name (e.g. "E2ESQLiteContainer")
	expectedSuffix string // expected value suffix after the PS-ID  (e.g. "-app-sqlite-1")
}

// containerChecks is the ordered list of container name constants to verify.
var containerChecks = []containerCheck{
	{constSuffix: "E2ESQLiteContainer", expectedSuffix: "-app-sqlite-1"},
	{constSuffix: "E2EPostgreSQL1Container", expectedSuffix: "-app-postgresql-1"},
	{constSuffix: "E2EPostgreSQL2Container", expectedSuffix: "-app-postgresql-2"},
}

// constValueRe matches Go string constant assignments of the form:
//
//	ConstName = "value"
//
// Group 1: constant name  Group 2: string value.
var constValueRe = regexp.MustCompile(`(\w+)\s*=\s*"([^"]+)"`)

// Check validates E2E container name constants from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates E2E container name constants under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking magic E2E container name constants...")

	var violations []string

	// Deduplicate by MagicFile: multiple PS may share the same magic file (e.g. identity).
	// When a magic file is shared, we only need to scan it once, but we must verify the
	// constant values against the correct PS-IDs present in that file.
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
		return fmt.Errorf("magic E2E container name violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("magic-e2e-container-names: all E2E container name constants match expected compose service names")

	return nil
}

// checkMagicFile reads the magic file for ps and verifies any E2E container name constants.
// If the file contains no E2ESQLiteContainer constant the PS is skipped (not all PS have E2E tests).
func checkMagicFile(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	magicPath := filepath.Join(rootDir, "internal", "shared", "magic", ps.MagicFile)

	src, err := os.ReadFile(magicPath)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read %s: %v", ps.PSID, ps.MagicFile, err)}
	}

	constMap := parseConstants(string(src))

	// If no *E2ESQLiteContainer constant exists in this magic file, skip it.
	hasSQLite := false

	for name := range constMap {
		if strings.HasSuffix(name, "E2ESQLiteContainer") {
			hasSQLite = true

			break
		}
	}

	if !hasSQLite {
		return nil
	}

	var violations []string

	for _, check := range containerChecks {
		v := verifyContainerConst(ps.PSID, ps.MagicFile, constMap, check)
		violations = append(violations, v...)
	}

	return violations
}

// parseConstants parses all Go string constant assignments from src.
// Returns a map of constant name → string value.
func parseConstants(src string) map[string]string {
	matches := constValueRe.FindAllStringSubmatch(src, -1)
	result := make(map[string]string, len(matches))

	for _, m := range matches {
		result[m[1]] = m[2]
	}

	return result
}

// verifyContainerConst finds the constant with the given suffix and verifies its value
// matches psID + check.expectedSuffix.
func verifyContainerConst(psID, magicFile string, constMap map[string]string, check containerCheck) []string {
	// Find the constant by suffix (prefix varies per PS).
	constName := ""
	constValue := ""

	for name, value := range constMap {
		if strings.HasSuffix(name, check.constSuffix) {
			constName = name
			constValue = value

			break
		}
	}

	if constName == "" {
		// Constant not found — this is a violation when E2ESQLiteContainer exists
		// but a sibling constant is missing.
		return []string{fmt.Sprintf("%s: %s: missing constant with suffix %s", psID, magicFile, check.constSuffix)}
	}

	expected := psID + check.expectedSuffix
	if constValue != expected {
		return []string{fmt.Sprintf("%s: %s: %s = %q, want %q", psID, magicFile, constName, constValue, expected)}
	}

	return nil
}
