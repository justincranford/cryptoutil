// Copyright (c) 2025-2026 Justin Cranford.
// Package apps_ps_id_server_package verifies that every PS-ID directory contains the required
// server package files: server/server.go for all 10 PS-IDs, and server/public_server.go for
// all PS-IDs except those in knownExclusionsPublicServer (currently sm-kms and skeleton-template).
package apps_ps_id_server_package

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// knownExclusionsPublicServer lists PS-IDs that currently lack server/public_server.go.
// Exclusions are removed as each PS-ID is brought into full conformance.
var knownExclusionsPublicServer = map[string]bool{
	cryptoutilSharedMagic.OTLPServiceSMKMS:            true, // Legacy service; no public server split yet.
	cryptoutilSharedMagic.OTLPServiceSkeletonTemplate: true, // Template service; no public server split yet.
}

// Check validates PS-ID server package files from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID server package files under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var errors []string

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serviceDir := filepath.Join(appsDir, ps.PSID)

		if errs := checkPSIDServerFiles(serviceDir, ps.PSID); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("apps PS-ID server package violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("apps-ps-id-server-package: all PS-IDs pass server package validation")

	return nil
}

// checkPSIDServerFiles verifies server/server.go (all) and server/public_server.go (most) exist.
func checkPSIDServerFiles(serviceDir, psid string) []string {
	var errors []string

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: PS-ID directory missing: internal/apps/%s/", serviceDir, psid)}
	}

	serverGo := filepath.Join(serviceDir, "server", "server.go")
	if _, err := os.Stat(serverGo); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/server/server.go", serviceDir, psid))
	}

	if !knownExclusionsPublicServer[psid] {
		publicServerGo := filepath.Join(serviceDir, "server", "public_server.go")
		if _, err := os.Stat(publicServerGo); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("%s: missing required file internal/apps/%s/server/public_server.go", serviceDir, psid))
		}
	}

	return errors
}
