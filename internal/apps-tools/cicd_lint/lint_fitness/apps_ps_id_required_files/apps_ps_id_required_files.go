// Copyright (c) 2025-2026 Justin Cranford.
// Package apps_ps_id_required_files verifies that every PS-ID directory under
// internal/apps/{PS-ID}/ contains the required entry file ({SERVICE}.go) and
// usage file ({SERVICE}_usage.go). Uses the canonical registry so new PS-IDs
// are automatically covered without manual code changes.
package apps_ps_id_required_files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// Check validates PS-ID required files from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID required files under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var errors []string

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serviceDir := filepath.Join(appsDir, ps.PSID)

		if errs := checkPSIDRequiredFiles(serviceDir, ps.PSID, ps.Service); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("apps PS-ID required files violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("apps-ps-id-required-files: all PS-IDs pass required files validation")

	return nil
}

// checkPSIDRequiredFiles verifies {SERVICE}.go and {SERVICE}_usage.go exist in the PS-ID dir.
func checkPSIDRequiredFiles(serviceDir, psid, service string) []string {
	var errors []string

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: PS-ID directory missing: internal/apps/%s/", serviceDir, psid)}
	}

	entryFile := filepath.Join(serviceDir, service+".go")
	if _, err := os.Stat(entryFile); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/%s.go", serviceDir, psid, service))
	}

	usageFile := filepath.Join(serviceDir, service+"_usage.go")
	if _, err := os.Stat(usageFile); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/%s_usage.go", serviceDir, psid, service))
	}

	return errors
}
