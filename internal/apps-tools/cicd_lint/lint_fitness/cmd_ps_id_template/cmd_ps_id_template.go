// Copyright (c) 2025 Justin Cranford

// Package cmd_ps_id_template verifies that every cmd/{PS-ID}/main.go entry point:
//  1. Exists at cmd/{PS-ID}/main.go.
//  2. Contains "package main".
//  3. Imports "cryptoutil/internal/apps/{PS-ID}".
//  4. Uses os.Args[1:] (passes tail of args to the service function, excluding the binary name).
//
// All 10 PS-IDs are verified from the canonical registry, so new PS-IDs are automatically
// covered without manual changes.
//
// See ENG-HANDBOOK.md Section 9.11.1 for the fitness sub-linter catalog.
package cmd_ps_id_template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// Check validates PS-ID cmd entry-point structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID cmd entry-point structure under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var violations []string

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		mainPath := filepath.Join(rootDir, "cmd", ps.PSID, "main.go")

		errs := checkCmdMainFile(mainPath, ps.PSID, "cryptoutil/internal/apps/"+ps.PSID, true)
		violations = append(violations, errs...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("cmd PS-ID template violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("cmd-ps-id-template: all PS-ID cmd entry points pass template validation")

	return nil
}

// checkCmdMainFile verifies a cmd main.go file for structural compliance.
// requireArgsSlice: true → must contain "os.Args[1:]"; false → must NOT contain "os.Args[1:]".
func checkCmdMainFile(mainPath, entityID, expectedImport string, requireArgsSlice bool) []string {
	var violations []string

	data, err := os.ReadFile(mainPath)
	if err != nil {
		violations = append(violations, fmt.Sprintf("%s: cmd/%s/main.go missing or unreadable: %v", entityID, entityID, err))

		return violations
	}

	content := string(data)

	hasPackageMain := strings.Contains(content, "package main")
	hasExpectedImport := strings.Contains(content, expectedImport)

	if !hasPackageMain {
		violations = append(violations, fmt.Sprintf("%s: cmd/%s/main.go missing 'package main'", entityID, entityID))
	}

	if !hasExpectedImport {
		violations = append(violations, fmt.Sprintf("%s: cmd/%s/main.go missing import %q", entityID, entityID, expectedImport))
	}

	switch {
	case requireArgsSlice && !strings.Contains(content, "os.Args[1:]"):
		violations = append(violations, fmt.Sprintf("%s: cmd/%s/main.go missing 'os.Args[1:]'", entityID, entityID))
	case !requireArgsSlice && strings.Contains(content, "os.Args[1:]"):
		violations = append(violations, fmt.Sprintf("%s: cmd/%s/main.go must use 'os.Args' not 'os.Args[1:]'", entityID, entityID))
	}

	return violations
}
