// Copyright (c) 2025 Justin Cranford

// Package standalone_config_presence validates that each product-service in the
// standalone config allowlist has all required config files present under
// deployments/{PS-ID}/config/.
//
// Required files per allowlist PS (PS-ID-prefixed):
//   - {PS-ID}-app-framework-common.yml
//   - {PS-ID}-app-framework-sqlite-1.yml
//   - {PS-ID}-app-framework-postgresql-1.yml
//   - {PS-ID}-app-framework-postgresql-2.yml
//
// Only sm-im and sm-kms are in the allowlist.
// Other product-services do not use the standardized standalone config layout.
package standalone_config_presence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// configFileSuffixes lists the suffixes appended to the PS-ID to form the required config filenames.
var configFileSuffixes = []string{
	"-app-framework-common.yml",
	"-app-framework-sqlite-1.yml",
	"-app-framework-postgresql-1.yml",
	"-app-framework-postgresql-2.yml",
}

// configAllowlist is the set of PS IDs that must have the required config files.
var configAllowlist = map[string]bool{
	cryptoutilSharedMagic.OTLPServiceSMIM:  true,
	cryptoutilSharedMagic.OTLPServiceSMKMS: true,
}

// Check validates standalone config file presence from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates standalone config file presence under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking standalone config file presence...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		if !configAllowlist[ps.PSID] {
			continue
		}

		v := checkConfigPresence(rootDir, ps)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("standalone config presence violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("standalone-config-presence: all allowlist product-services have required config files")

	return nil
}

// checkConfigPresence verifies that each required config file exists for ps.
func checkConfigPresence(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	configDir := filepath.Join(rootDir, "deployments", ps.PSID, "config")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: deployments/%s/config/ directory does not exist", ps.PSID, ps.PSID)}
	}

	var violations []string

	for _, suffix := range configFileSuffixes {
		filename := ps.PSID + suffix

		configPath := filepath.Join(configDir, filename)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: deployments/%s/config/%s: file does not exist", ps.PSID, ps.PSID, filename))
		}
	}

	return violations
}
