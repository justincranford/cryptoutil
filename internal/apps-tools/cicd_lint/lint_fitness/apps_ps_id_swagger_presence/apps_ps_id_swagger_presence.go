// Copyright (c) 2025 Justin Cranford

// Package apps_ps_id_swagger_presence verifies that every PS-ID has swagger.go and
// swagger_test.go files inside its server/ subdirectory.  All PS-IDs are currently listed in
// knownExclusions because swagger files live at the service root rather than server/; they will be
// removed from the exclusion list as each PS-ID is migrated during the conformance phase.
package apps_ps_id_swagger_presence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

// knownExclusions lists PS-IDs that are temporarily exempt from server swagger checks.
var knownExclusions = map[string]bool{}

// Check validates PS-ID swagger file presence from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID swagger file presence under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, knownExclusions)
}

// checkInDirWithExclusions implements the validation logic with a configurable exclusion set.
// This seam allows tests to inject an empty exclusion set and exercise all code paths.
func checkInDirWithExclusions(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, exclusions map[string]bool) error {
	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var errors []string

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		if exclusions[ps.PSID] {
			continue
		}

		serviceDir := filepath.Join(appsDir, ps.PSID)

		if errs := checkPSIDSwaggerFiles(serviceDir, ps.PSID); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("apps PS-ID swagger presence violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("apps-ps-id-swagger-presence: all non-excluded PS-IDs pass swagger file validation")

	return nil
}

// checkPSIDSwaggerFiles verifies server/swagger.go and server/swagger_test.go exist.
func checkPSIDSwaggerFiles(serviceDir, psid string) []string {
	var errors []string

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: PS-ID directory missing: internal/apps/%s/", serviceDir, psid)}
	}

	for _, filename := range []string{"swagger.go", "swagger_test.go"} {
		target := filepath.Join(serviceDir, "server", filename)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/server/%s", serviceDir, psid, filename))
		}
	}

	return errors
}
