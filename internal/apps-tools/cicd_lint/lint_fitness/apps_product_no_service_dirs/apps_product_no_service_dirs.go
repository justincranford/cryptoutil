// Copyright (c) 2025-2026 Justin Cranford.
// Package apps_product_no_service_dirs verifies that product directories under internal/apps/ do
// not contain subdirectories named after individual services.  Service-named subdirectories are
// duplicate copies of the canonical PS-ID entry point (internal/apps/{PS-ID}/); they must be
// deleted once duplicate content is confirmed.
//
// Violations are cross-referenced against the entity registry: a subdirectory is a violation only
// when its name matches a known service name for the enclosing product.
//
// All five existing service subdirs are listed in knownExclusions until they are deleted during
// the conformance migration phase.
package apps_product_no_service_dirs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// knownExclusions maps "product/service" pairs that are currently present and scheduled for
// deletion.  All five service subdirs were deleted during conformance migration (framework-v17 Phase 5).
var knownExclusions = map[string]bool{}

// Check validates product directories from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates product directories under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, knownExclusions)
}

// checkInDirWithExclusions implements the validation logic with a configurable exclusion set.
func checkInDirWithExclusions(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, exclusions map[string]bool) error {
	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	// Build product → set of service names from registry.
	productServices := buildProductServiceMap()

	var errors []string

	for _, product := range cryptoutilFitnessRegistry.AllProducts() {
		productDir := filepath.Join(appsDir, product.ID)

		if errs := checkProductDir(productDir, product.ID, productServices[product.ID], exclusions); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("apps product no-service-dirs violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("apps-product-no-service-dirs: all product directories pass service-dir validation")

	return nil
}

// buildProductServiceMap returns a map from product ID to its set of service names.
func buildProductServiceMap() map[string]map[string]bool {
	result := make(map[string]map[string]bool)

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		if result[ps.Product] == nil {
			result[ps.Product] = make(map[string]bool)
		}

		result[ps.Product][ps.Service] = true
	}

	return result
}

// checkProductDir scans subdirectories in productDir for service-named directories.
func checkProductDir(productDir, productID string, serviceNames map[string]bool, exclusions map[string]bool) []string {
	if _, err := os.Stat(productDir); os.IsNotExist(err) {
		// Missing product dir is not a violation here; product_structure handles that check.
		return nil
	}

	entries, err := os.ReadDir(productDir)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read product directory: %v", productDir, err)}
	}

	errors := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subdirName := entry.Name()
		if !serviceNames[subdirName] {
			continue // not a service-named subdir — allowed
		}

		key := productID + "/" + subdirName
		if exclusions[key] {
			continue // known exclusion — skip
		}

		errors = append(errors, fmt.Sprintf("%s/%s: service-named subdir found in product directory (delete duplicate)", productDir, subdirName))
	}

	return errors
}
