// Copyright (c) 2025 Justin Cranford

// Package cmd_product_template verifies that every cmd/{PRODUCT}/main.go entry point:
//  1. Exists at cmd/{PRODUCT}/main.go.
//  2. Contains "package main".
//  3. Imports "cryptoutil/internal/apps/{PRODUCT}".
//  4. Uses os.Args[1:] (passes tail of args to the product function).
//
// All 5 products are verified from the canonical registry, so new products are automatically
// covered without manual changes.
//
// See ENG-HANDBOOK.md Section 9.11.1 for the fitness sub-linter catalog.
package cmd_product_template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// Check validates product cmd entry-point structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates product cmd entry-point structure under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var violations []string

	for _, product := range cryptoutilFitnessRegistry.AllProducts() {
		mainPath := filepath.Join(rootDir, "cmd", product.ID, "main.go")

		errs := checkCmdMainFile(mainPath, product.ID, "cryptoutil/internal/apps/"+product.ID, true)
		violations = append(violations, errs...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("cmd product template violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("cmd-product-template: all product cmd entry points pass template validation")

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
