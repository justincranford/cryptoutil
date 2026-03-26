// Copyright (c) 2025 Justin Cranford

// Package standalone_config_otlp_names validates that the otlp-service value in
// each deployment config file matches the canonical naming convention.
//
// For each product-service in the allowlist (sm-im, sm-kms), each required
// config file under deployments/{PS-ID}/config/ must have an otlp-service value
// following the pattern:
//   - {PS-ID}-app-sqlite-1.yml       -> {PS-ID}-sqlite-1
//   - {PS-ID}-app-sqlite-2.yml       -> {PS-ID}-sqlite-2
//   - {PS-ID}-app-postgresql-1.yml   -> {PS-ID}-postgres-1
//   - {PS-ID}-app-postgresql-2.yml   -> {PS-ID}-postgres-2
//
// Only sm-im and sm-kms are in the allowlist.
// The check is registry-driven: it uses the canonical PS registry to determine
// which product-services to validate, rather than scanning the filesystem.
package standalone_config_otlp_names

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// configRule maps a standalone config file suffix to its expected otlp-service suffix.
// Filename is constructed as: {PS-ID} + filenameSuffix.
// Full expected otlp-service value: {PS-ID} + expectedOTLPSuffix.
type configRule struct {
	filenameSuffix     string
	expectedOTLPSuffix string
}

// otlpConfigRules lists the required config file suffixes and their expected otlp-service suffix.
var otlpConfigRules = []configRule{
	{filenameSuffix: "-app-sqlite-1.yml", expectedOTLPSuffix: "-sqlite-1"},
	{filenameSuffix: "-app-sqlite-2.yml", expectedOTLPSuffix: "-sqlite-2"},
	{filenameSuffix: "-app-postgresql-1.yml", expectedOTLPSuffix: "-postgres-1"},
	{filenameSuffix: "-app-postgresql-2.yml", expectedOTLPSuffix: "-postgres-2"},
}

// configAllowlist is the set of PS IDs whose standalone configs are validated.
var configAllowlist = map[string]bool{
	cryptoutilSharedMagic.OTLPServiceSMIM:  true,
	cryptoutilSharedMagic.OTLPServiceSMKMS: true,
}

// Check validates standalone config OTLP service names from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates standalone config OTLP service names under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking standalone config OTLP service names...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		if !configAllowlist[ps.PSID] {
			continue
		}

		v := checkOTLPNames(rootDir, ps)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("standalone config OTLP name violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("standalone-config-otlp-names: all allowlist product-services have correct OTLP service names")

	return nil
}

// checkOTLPNames validates the otlp-service values in each required config file for ps.
func checkOTLPNames(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	configDir := filepath.Join(rootDir, "deployments", ps.PSID, "config")

	var violations []string

	for _, rule := range otlpConfigRules {
		filename := ps.PSID + rule.filenameSuffix
		configPath := filepath.Join(configDir, filename)

		fileViolations := checkOTLPServiceValue(configPath, ps.PSID, rule.expectedOTLPSuffix, rootDir)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// checkOTLPServiceValue parses a config YAML and validates the otlp-service value.
func checkOTLPServiceValue(configPath, psID, expectedSuffix, rootDir string) []string {
	data, err := os.ReadFile(configPath) //nolint:gosec // configPath from controlled registry-driven path
	if err != nil {
		if os.IsNotExist(err) {
			// File absence is a standalone-config-presence violation, not an OTLP names violation.
			return nil
		}

		rel, _ := filepath.Rel(rootDir, configPath)

		return []string{fmt.Sprintf("%s: cannot read config file: %v", rel, err)}
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		rel, _ := filepath.Rel(rootDir, configPath)

		return []string{fmt.Sprintf("%s: YAML parse error: %v", rel, err)}
	}

	otlpServiceRaw, ok := config["otlp-service"]
	if !ok {
		rel, _ := filepath.Rel(rootDir, configPath)

		return []string{fmt.Sprintf("%s: missing required otlp-service key", rel)}
	}

	otlpService, ok := otlpServiceRaw.(string)
	if !ok {
		rel, _ := filepath.Rel(rootDir, configPath)

		return []string{fmt.Sprintf("%s: otlp-service value is not a string", rel)}
	}

	expected := psID + expectedSuffix

	if otlpService != expected {
		rel, _ := filepath.Rel(rootDir, configPath)

		return []string{fmt.Sprintf("%s: otlp-service: got %q, want %q", rel, otlpService, expected)}
	}

	return nil
}
