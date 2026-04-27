// Copyright (c) 2025 Justin Cranford

// Package apps_ps_id_test_patterns verifies that every PS-ID's server/ directory contains the
// expected test infrastructure files: testmain_test.go, a *_lifecycle_test.go file, and a
// *_port_conflict_test.go file.  PS-IDs that have not yet been brought into conformance are
// listed in the per-check exclusion maps.
package apps_ps_id_test_patterns

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// knownExclusionsTestMain lists PS-IDs that lack server/testmain_test.go.
var knownExclusionsTestMain = map[string]bool{}

// knownExclusionsLifecycle lists PS-IDs that lack server/*_lifecycle_test.go.
var knownExclusionsLifecycle = map[string]bool{
	cryptoutilSharedMagic.OTLPServiceSMIM:             true,
	cryptoutilSharedMagic.OTLPServiceJoseJA:           true,
	cryptoutilSharedMagic.OTLPServiceIdentityAuthz:    true,
	cryptoutilSharedMagic.OTLPServiceIdentityIDP:      true,
	cryptoutilSharedMagic.OTLPServiceIdentityRS:       true,
	cryptoutilSharedMagic.OTLPServiceIdentityRP:       true,
	cryptoutilSharedMagic.OTLPServiceIdentitySPA:      true,
	cryptoutilSharedMagic.OTLPServiceSkeletonTemplate: true,
}

// knownExclusionsPortConflict lists PS-IDs that lack server/*_port_conflict_test.go.
// All except sm-kms are excluded; no other PS-ID has this file yet.
var knownExclusionsPortConflict = map[string]bool{
	cryptoutilSharedMagic.OTLPServiceSMIM:             true,
	cryptoutilSharedMagic.OTLPServiceJoseJA:           true,
	cryptoutilSharedMagic.OTLPServicePKICA:            true,
	cryptoutilSharedMagic.OTLPServiceIdentityAuthz:    true,
	cryptoutilSharedMagic.OTLPServiceIdentityIDP:      true,
	cryptoutilSharedMagic.OTLPServiceIdentityRS:       true,
	cryptoutilSharedMagic.OTLPServiceIdentityRP:       true,
	cryptoutilSharedMagic.OTLPServiceIdentitySPA:      true,
	cryptoutilSharedMagic.OTLPServiceSkeletonTemplate: true,
}

// Check validates PS-ID test pattern files from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID test pattern files under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, knownExclusionsTestMain, knownExclusionsLifecycle, knownExclusionsPortConflict)
}

// checkInDirWithExclusions implements the validation logic with configurable exclusion sets.
func checkInDirWithExclusions(
	logger *cryptoutilCmdCicdCommon.Logger,
	rootDir string,
	exclTestMain, exclLifecycle, exclPortConflict map[string]bool,
) error {
	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var errors []string

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serviceDir := filepath.Join(appsDir, ps.PSID)

		if errs := checkPSIDTestPatterns(serviceDir, ps.PSID, exclTestMain, exclLifecycle, exclPortConflict); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("apps PS-ID test pattern violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("apps-ps-id-test-patterns: all non-excluded PS-IDs pass test pattern validation")

	return nil
}

// checkPSIDTestPatterns verifies the three test pattern files for a single PS-ID.
func checkPSIDTestPatterns(serviceDir, psid string, exclTestMain, exclLifecycle, exclPortConflict map[string]bool) []string {
	var errors []string

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: PS-ID directory missing: internal/apps/%s/", serviceDir, psid)}
	}

	serverDir := filepath.Join(serviceDir, "server")

	if !exclTestMain[psid] {
		if !fileExists(filepath.Join(serverDir, "testmain_test.go")) {
			errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/server/testmain_test.go", serviceDir, psid))
		}
	}

	if !exclLifecycle[psid] {
		if !globExists(serverDir, "*_lifecycle_test.go") {
			errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/server/*_lifecycle_test.go", serviceDir, psid))
		}
	}

	if !exclPortConflict[psid] {
		if !globExists(serverDir, "*_port_conflict_test.go") {
			errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/server/*_port_conflict_test.go", serviceDir, psid))
		}
	}

	return errors
}

// fileExists returns true if the path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}

// globExists returns true if at least one file matches the pattern in dir.
func globExists(dir, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))

	return err == nil && len(matches) > 0
}
