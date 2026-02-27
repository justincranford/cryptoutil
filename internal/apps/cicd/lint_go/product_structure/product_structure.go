// Copyright (c) 2025 Justin Cranford

// Package product_structure validates product directory conventions.
package product_structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// knownProducts lists all product names that must follow structural conventions.
var knownProducts = []string{"identity", "jose", "pki", "skeleton", "sm"}

// Check validates product structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates product directories under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var errors []string

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	for _, product := range knownProducts {
		productDir := filepath.Join(appsDir, product)
		if errs := checkProductDir(productDir, product); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("product structure violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("product-structure: all products pass structural validation")

	return nil
}

// checkProductDir validates a single product directory.
func checkProductDir(productDir, product string) []string {
	var errors []string

	if _, err := os.Stat(productDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: product directory missing", productDir)}
	}

	// Check for PRODUCT.go entry file.
	entryFile := filepath.Join(productDir, product+".go")
	if _, err := os.Stat(entryFile); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("%s: missing entry file %s.go", productDir, product))
	}

	// Check for PRODUCT_test.go test file.
	testFile := filepath.Join(productDir, product+"_test.go")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("%s: missing test file %s_test.go", productDir, product))
	}

	return errors
}
