// Copyright (c) 2025 Justin Cranford

// Package apps_ps_id_template verifies that every PS-ID directory under internal/apps/{PS-ID}/
// conforms to the canonical MANIFEST.yaml template. The template is read at runtime from
// api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml.
//
// Placeholders are substituted per PS-ID before each check:
//   - __SERVICE__ → service component of the PS-ID (e.g. "kms" for "sm-kms")
//   - __PS_ID__   → full PS-ID (e.g. "sm-kms")
//
// See ENG-HANDBOOK.md Section 9.11.1 for the fitness sub-linter catalog.
package apps_ps_id_template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// psIDManifest mirrors the YAML structure of the PS-ID MANIFEST.yaml template.
type psIDManifest struct {
	RequiredRootFiles   []string `yaml:"required_root_files"`
	RequiredDirs        []string `yaml:"required_dirs"`
	RequiredServerFiles []string `yaml:"required_server_files"`
}

// allTenPSIDs is a convenience set containing all 10 known PS-IDs used in exclusion maps.
// Update when PS-IDs are added or removed from the registry.
var allTenPSIDs = map[string]bool{
	cryptoutilSharedMagic.OTLPServiceSMKMS:            true,
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

// knownRootFileExclusions maps template filenames (pre-substitution) to sets of PS-IDs
// that are exempt from that specific root-file check.
// Remove entries as each PS-ID migrates to the canonical structure.
var knownRootFileExclusions = map[string]map[string]bool{
	// sm-im has im_cli_commands_test.go + im_cli_url_test.go instead of im_cli_test.go.
	"__SERVICE___cli_test.go": {
		cryptoutilSharedMagic.OTLPServiceSMIM: true,
	},
}

// knownServerFileExclusions maps template filenames (pre-substitution) to sets of PS-IDs
// that are exempt from that specific server-file check.
// Remove entries as each PS-ID migrates swagger/test files to server/.
var knownServerFileExclusions = map[string]map[string]bool{
	// swagger files live at PS-ID root, not server/ yet; all 10 excluded during migration.
	"swagger.go":      allTenPSIDs,
	"swagger_test.go": allTenPSIDs,
	// sm-kms: testmain_test.go not yet present in server/ (pending migration).
	"testmain_test.go": {
		cryptoutilSharedMagic.OTLPServiceSMKMS: true,
	},
	// lifecycle and port-conflict tests live at PS-ID root, not server/ yet; all 10 excluded.
	"__SERVICE___lifecycle_test.go":     allTenPSIDs,
	"__SERVICE___port_conflict_test.go": allTenPSIDs,
}

// manifestRelPath is the path to the PS-ID MANIFEST.yaml relative to rootDir.
const manifestRelPath = "api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml"

// Check validates PS-ID template conformance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID template conformance under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, knownRootFileExclusions, knownServerFileExclusions)
}

// checkInDirWithExclusions implements the validation logic with configurable exclusion sets.
// rootExcl and serverExcl map pre-substitution template filenames to sets of PS-IDs to skip.
// This seam allows tests to inject empty exclusion sets and exercise all code paths.
func checkInDirWithExclusions(
	logger *cryptoutilCmdCicdCommon.Logger,
	rootDir string,
	rootExcl map[string]map[string]bool,
	serverExcl map[string]map[string]bool,
) error {
	manifestPath := filepath.Join(rootDir, filepath.FromSlash(manifestRelPath))

	data, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		return fmt.Errorf("failed to read PS-ID MANIFEST.yaml at %s: %w", manifestPath, readErr)
	}

	var manifest psIDManifest
	if unmarshalErr := yaml.Unmarshal(data, &manifest); unmarshalErr != nil {
		return fmt.Errorf("failed to parse PS-ID MANIFEST.yaml at %s: %w", manifestPath, unmarshalErr)
	}

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var violations []string

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(appsDir, ps.PSID)
		errs := checkPSIDFiles(psDir, ps, manifest, rootExcl, serverExcl)
		violations = append(violations, errs...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("apps PS-ID template violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("apps-ps-id-template: all non-excluded PS-IDs pass template validation")

	return nil
}

// checkPSIDFiles checks one PS-ID directory against the manifest, applying exclusions.
func checkPSIDFiles(
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	rootExcl map[string]map[string]bool,
	serverExcl map[string]map[string]bool,
) []string {
	var violations []string

	// Check required root files.
	for _, tmplFile := range manifest.RequiredRootFiles {
		if excl, ok := rootExcl[tmplFile]; ok && excl[ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(psDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required root file: %s", ps.PSID, expanded))
		}
	}

	// Check required directories.
	for _, dir := range manifest.RequiredDirs {
		if _, err := os.Stat(filepath.Join(psDir, dir)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required directory: %s", ps.PSID, dir))
		}
	}

	// Check required server files.
	serverDir := filepath.Join(psDir, "server")

	for _, tmplFile := range manifest.RequiredServerFiles {
		if excl, ok := serverExcl[tmplFile]; ok && excl[ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(serverDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required server file: %s", ps.PSID, expanded))
		}
	}

	return violations
}
