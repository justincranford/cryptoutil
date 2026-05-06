// Copyright (c) 2025-2026 Justin Cranford.
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
	RequiredRootFiles             []string `yaml:"required_root_files"`
	RequiredDirs                  []string `yaml:"required_dirs"`
	RequiredServerFiles           []string `yaml:"required_server_files"`
	RequiredServerDirs            []string `yaml:"required_server_dirs"`
	RequiredServerConfigFiles     []string `yaml:"required_server_config_files"`
	RequiredServerRepositoryFiles []string `yaml:"required_server_repository_files"`
	RequiredServerRepositoryDirs  []string `yaml:"required_server_repository_dirs"`
	RequiredE2EFiles              []string `yaml:"required_e2e_files"`
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

// psIDExclusions bundles all exclusion maps to avoid a long parameter list.
// Each field maps a template filename (pre-substitution) to a set of PS-IDs exempt from that check.
type psIDExclusions struct {
	enforceRootTemplates bool
	rootFiles            map[string]map[string]bool
	serverFiles          map[string]map[string]bool
	serverDirs           map[string]map[string]bool
	configFiles          map[string]map[string]bool
	repoFiles            map[string]map[string]bool
	repoDirs             map[string]map[string]bool
	e2eFiles             map[string]map[string]bool
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
	// Lifecycle tests are still non-canonical for these services.
	"__SERVICE___lifecycle_test.go": {
		cryptoutilSharedMagic.OTLPServiceJoseJA:           true,
		cryptoutilSharedMagic.OTLPServicePKICA:            true,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate: true,
	},
	// Port-conflict tests live at PS-ID root, not server/ yet; all 10 excluded.
	"__SERVICE___port_conflict_test.go": allTenPSIDs,
}

// knownServerDirExclusions maps required server/ subdirectory names to PS-IDs excluded from that check.
// Remove entries as each PS-ID creates the missing subdirectory.
var knownServerDirExclusions = map[string]map[string]bool{
	// apis/: sm-kms uses businesslogic/handler layout (V20); pki-ca uses cmd/config/middleware (V20).
	// identity-* migrated and are no longer excluded.
	"apis": {
		cryptoutilSharedMagic.OTLPServiceSMKMS: true,
		cryptoutilSharedMagic.OTLPServicePKICA: true,
	},
	// model/: sm-kms and pki-ca pending V20.
	"model": {
		cryptoutilSharedMagic.OTLPServiceSMKMS: true,
		cryptoutilSharedMagic.OTLPServicePKICA: true,
	},
	// repository/: pki-ca pending V20 migration (sm-kms already has repository/).
	"repository": {
		cryptoutilSharedMagic.OTLPServicePKICA: true,
	},
}

// knownServerConfigFileExclusions maps required server/config/ filenames to exempt PS-IDs.
// Remove entries as each PS-ID adds the missing config file.
var knownServerConfigFileExclusions = map[string]map[string]bool{
	// sm-kms has no server/config/ directory yet (pending V20 migration).
	"config.go": {
		cryptoutilSharedMagic.OTLPServiceSMKMS: true,
	},
	"config_test.go": {
		cryptoutilSharedMagic.OTLPServiceSMKMS: true,
	},
	// config_test_helper.go: sm-kms (no server/config/), pki-ca (pending V20 migration).
	"config_test_helper.go": {
		cryptoutilSharedMagic.OTLPServiceSMKMS: true,
		cryptoutilSharedMagic.OTLPServicePKICA: true,
	},
}

// knownServerRepositoryFileExclusions maps required server/repository/ filenames to exempt PS-IDs.
// Remove entries as each PS-ID creates its server/repository/ directory with required files.
var knownServerRepositoryFileExclusions = map[string]map[string]bool{
	// pki-ca has no server/repository/ directory yet (pending V20 migration).
	"migrations.go": {
		cryptoutilSharedMagic.OTLPServicePKICA: true,
	},
}

// knownServerRepositoryDirExclusions maps required server/repository/ subdirectory names to exempt PS-IDs.
// Remove entries as each PS-ID creates its server/repository/migrations/ directory.
var knownServerRepositoryDirExclusions = map[string]map[string]bool{
	// pki-ca has no server/repository/ directory yet (pending V20 migration).
	"migrations": {
		cryptoutilSharedMagic.OTLPServicePKICA: true,
	},
}

// knownE2EFileExclusions maps required e2e/ filenames to exempt PS-IDs.
// checkE2EFiles only fires when e2e/ dir exists; services without e2e/ are skipped automatically.
// Remove entries as each PS-ID adopts the canonical e2e file naming convention.
var knownE2EFileExclusions = map[string]map[string]bool{
	// All identity services now include testmain_e2e_test.go.
	"testmain_e2e_test.go": {},
	// Services listed here still use non-canonical smoke test filenames.
	"__SERVICE___e2e_test.go": {
		cryptoutilSharedMagic.OTLPServiceSMKMS:            true,
		cryptoutilSharedMagic.OTLPServiceSMIM:             true,
		cryptoutilSharedMagic.OTLPServiceJoseJA:           true,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate: true,
	},
}

// manifestRelPath is the path to the PS-ID MANIFEST.yaml relative to rootDir.
const manifestRelPath = "api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml"

// Check validates PS-ID template conformance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates PS-ID template conformance under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	excl := psIDExclusions{
		enforceRootTemplates: true,
		rootFiles:            knownRootFileExclusions,
		serverFiles:          knownServerFileExclusions,
		serverDirs:           knownServerDirExclusions,
		configFiles:          knownServerConfigFileExclusions,
		repoFiles:            knownServerRepositoryFileExclusions,
		repoDirs:             knownServerRepositoryDirExclusions,
		e2eFiles:             knownE2EFileExclusions,
	}

	return checkInDirWithExclusions(logger, rootDir, excl)
}

// checkInDirWithExclusions implements the validation logic with configurable exclusion sets.
// The excl struct maps pre-substitution template filenames to sets of PS-IDs to skip.
// This seam allows tests to inject empty exclusion sets and exercise all code paths.
func checkInDirWithExclusions(
	logger *cryptoutilCmdCicdCommon.Logger,
	rootDir string,
	excl psIDExclusions,
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
		errs := checkPSIDFiles(rootDir, psDir, ps, manifest, excl)
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
	rootDir string,
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	excl psIDExclusions,
) []string {
	var violations []string

	// Check required root files.
	for _, tmplFile := range manifest.RequiredRootFiles {
		if excl, ok := excl.rootFiles[tmplFile]; ok && excl[ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(psDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required root file: %s", ps.PSID, expanded))
		}
	}

	if excl.enforceRootTemplates {
		violations = append(violations, checkRootTemplates(rootDir, psDir, ps)...)
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
		if excl, ok := excl.serverFiles[tmplFile]; ok && excl[ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(serverDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required server file: %s", ps.PSID, expanded))
		}
	}

	violations = append(violations, checkServerDirs(psDir, ps, manifest, excl)...)
	violations = append(violations, checkServerConfigFiles(psDir, ps, manifest, excl)...)
	violations = append(violations, checkServerRepositoryFiles(psDir, ps, manifest, excl)...)
	violations = append(violations, checkServerRepositoryDirs(psDir, ps, manifest, excl)...)
	violations = append(violations, checkE2EFiles(psDir, ps, manifest, excl)...)

	return violations
}

// checkServerDirs verifies that each entry in RequiredServerDirs exists under server/.
func checkServerDirs(
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	excl psIDExclusions,
) []string {
	var violations []string

	serverDir := filepath.Join(psDir, "server")

	for _, dir := range manifest.RequiredServerDirs {
		if excl.serverDirs[dir][ps.PSID] {
			continue
		}

		if _, err := os.Stat(filepath.Join(serverDir, dir)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required server subdirectory: %s", ps.PSID, dir))
		}
	}

	return violations
}

// checkServerConfigFiles verifies that each entry in RequiredServerConfigFiles exists under server/config/.
func checkServerConfigFiles(
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	excl psIDExclusions,
) []string {
	var violations []string

	configDir := filepath.Join(psDir, "server", "config")

	for _, tmplFile := range manifest.RequiredServerConfigFiles {
		if excl.configFiles[tmplFile][ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(configDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required server config file: %s", ps.PSID, expanded))
		}
	}

	return violations
}

// checkServerRepositoryFiles verifies that each entry in RequiredServerRepositoryFiles exists under server/repository/.
func checkServerRepositoryFiles(
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	excl psIDExclusions,
) []string {
	var violations []string

	repoDir := filepath.Join(psDir, "server", "repository")

	for _, tmplFile := range manifest.RequiredServerRepositoryFiles {
		if excl.repoFiles[tmplFile][ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(repoDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required server repository file: %s", ps.PSID, expanded))
		}
	}

	return violations
}

// checkServerRepositoryDirs verifies that each entry in RequiredServerRepositoryDirs exists under server/repository/.
func checkServerRepositoryDirs(
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	excl psIDExclusions,
) []string {
	var violations []string

	repoDir := filepath.Join(psDir, "server", "repository")

	for _, dir := range manifest.RequiredServerRepositoryDirs {
		if excl.repoDirs[dir][ps.PSID] {
			continue
		}

		if _, err := os.Stat(filepath.Join(repoDir, dir)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required server repository subdirectory: %s", ps.PSID, dir))
		}
	}

	return violations
}

// checkE2EFiles verifies that each entry in RequiredE2EFiles exists under e2e/.
// This check is skipped entirely when the e2e/ directory does not exist.
func checkE2EFiles(
	psDir string,
	ps cryptoutilFitnessRegistry.ProductService,
	manifest psIDManifest,
	excl psIDExclusions,
) []string {
	e2eDir := filepath.Join(psDir, "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		return nil
	}

	var violations []string

	for _, tmplFile := range manifest.RequiredE2EFiles {
		if excl.e2eFiles[tmplFile][ps.PSID] {
			continue
		}

		expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyService, ps.Service)
		expanded = strings.ReplaceAll(expanded, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, ps.PSID)

		if _, err := os.Stat(filepath.Join(e2eDir, expanded)); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing required e2e file: %s", ps.PSID, expanded))
		}
	}

	return violations
}
