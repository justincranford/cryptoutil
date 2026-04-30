// Copyright (c) 2025-2026 Justin Cranford.
// Package cmd_suite_template verifies that cmd/{SUITE}/main.go entry points:
//  1. Exist at cmd/{SUITE}/main.go.
//  2. Contain "package main".
//  3. Import "cryptoutil/internal/apps/{SUITE}".
//  4. Use os.Args directly (NOT os.Args[1:]) — the suite passes the full args including
//     the binary name, and routing is done by the suite dispatcher.
//
// See ENG-HANDBOOK.md Section 9.11.1 for the fitness sub-linter catalog.
package cmd_suite_template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// Check validates suite cmd entry-point structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates suite cmd entry-point structure under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var violations []string

	for _, suite := range cryptoutilFitnessRegistry.AllSuites() {
		mainPath := filepath.Join(rootDir, "cmd", suite.ID, "main.go")

		// requireArgsSlice=false → suite must use os.Args, not os.Args[1:].
		errs := checkCmdMainFile(mainPath, suite.ID, "cryptoutil/internal/apps/"+suite.ID, false)
		violations = append(violations, errs...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("cmd suite template violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("cmd-suite-template: all suite cmd entry points pass template validation")

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
