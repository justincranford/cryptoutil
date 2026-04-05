// Copyright (c) 2025 Justin Cranford

// Package otlp_service_name_pattern validates that service config files use the
// canonical otlp-service naming convention:
//   - Standalone: {ps-id}-sqlite.yml  -> {ps-id}-sqlite-1
//   - Deployment:  deployments/{ps-id}/config/{ps-id}-app-sqlite-1.yml  -> {ps-id}-sqlite-1
//   - Deployment:  deployments/{ps-id}/config/{ps-id}-app-postgresql-1.yml -> {ps-id}-postgres-1
//   - Deployment:  deployments/{ps-id}/config/{ps-id}-app-postgresql-2.yml -> {ps-id}-postgres-2
//
// See ENG-HANDBOOK.md Section 9.11 for naming convention details.
package otlp_service_name_pattern

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
// Filename is constructed as: {ps-id} + filenameSuffix.
// The prefix is computed from the service ID (product-service).
type configRule struct {
	filenameSuffix     string
	expectedOTLPSuffix string
}

// standaloneConfigRules defines the expected otlp-service suffix for each config file suffix.
// Full expected value: {ps-id} + expectedOTLPSuffix.
var standaloneConfigRules = []configRule{
	{filenameSuffix: "-sqlite.yml", expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantSQLite1},
	{filenameSuffix: "-pg-1.yml", expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantPostgres1},
	{filenameSuffix: "-pg-2.yml", expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantPostgres2},
}

// deploymentConfigRules defines the expected otlp-service suffix for deployment config overlay files.
// File: deployments/{psid}/config/{psid}{filenameSuffix} → otlp-service: {psid}-{expectedOTLPSuffix}.
var deploymentConfigRules = []configRule{
	{filenameSuffix: lintFitnessRegistry.DeploymentConfigSuffixSQLite1, expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantSQLite1},
	{filenameSuffix: lintFitnessRegistry.DeploymentConfigSuffixSQLite2, expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantSQLite2},
	{filenameSuffix: lintFitnessRegistry.DeploymentConfigSuffixPostgresql1, expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantPostgres1},
	{filenameSuffix: lintFitnessRegistry.DeploymentConfigSuffixPostgresql2, expectedOTLPSuffix: "-" + lintFitnessRegistry.ComposeVariantPostgres2},
}

// excludedProductDirs lists top-level directories under configs/ that are NOT product directories.
// These are skipped to avoid treating archived or utility directories as product-service configs.
var excludedProductDirs = map[string]bool{
	"orphaned": true,
}

// Check validates otlp-service name patterns in standalone config files.
// Scans configs/{PRODUCT}/{SERVICE}/config-*.yml files.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.ReadDir, os.ReadFile)
}

// CheckInDir validates otlp-service names under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking OTLP service name patterns in standalone config files...")

	configsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir)

	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		return fmt.Errorf("configs/ directory not found at %s", configsDir)
	}

	var violations []string

	// Walk configs/{PRODUCT}/{SERVICE}/ looking for config-*.yml files at depth 2.
	productEntries, err := readDirFn(configsDir)
	if err != nil {
		return fmt.Errorf("failed to read configs dir: %w", err)
	}

	for _, productEntry := range productEntries {
		if !productEntry.IsDir() {
			continue
		}

		if excludedProductDirs[productEntry.Name()] {
			continue
		}

		productDir := filepath.Join(configsDir, productEntry.Name())

		serviceEntries, readErr := readDirFn(productDir)
		if readErr != nil {
			return fmt.Errorf("failed to read product dir %s: %w", productDir, readErr)
		}

		for _, serviceEntry := range serviceEntries {
			if !serviceEntry.IsDir() {
				continue
			}

			serviceDir := filepath.Join(productDir, serviceEntry.Name())
			psID := productEntry.Name() + "-" + serviceEntry.Name()

			v := checkServiceDir(serviceDir, psID, rootDir, readFileFn)
			violations = append(violations, v...)
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("OTLP service name violations:\n%s", strings.Join(violations, "\n"))
	}

	// Also validate deployment config overlay files using the entity registry.
	deploymentViolations := checkDeploymentConfigs(rootDir, readFileFn)
	violations = append(violations, deploymentViolations...)

	if len(violations) > 0 {
		return fmt.Errorf("OTLP service name violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("otlp-service-name-pattern: all standalone config files use canonical names")

	return nil
}

// checkDeploymentConfigs validates otlp-service values in deployment config overlay files
// using the entity registry to iterate product-services by PS-ID.
func checkDeploymentConfigs(rootDir string, readFileFn func(string) ([]byte, error)) []string {
	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkDeploymentConfigDir(rootDir, ps.PSID, readFileFn)
		violations = append(violations, v...)
	}

	return violations
}

// checkDeploymentConfigDir checks all deployment config overlay files for a single PS-ID.
// It looks in deployments/{psid}/config/ for files matching the deployment config rules.
func checkDeploymentConfigDir(rootDir, psID string, readFileFn func(string) ([]byte, error)) []string {
	var violations []string

	configDir := filepath.Join(rootDir, "deployments", psID, "config")

	for _, rule := range deploymentConfigRules {
		filename := psID + rule.filenameSuffix
		configPath := filepath.Join(configDir, filename)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue // File does not exist — not a violation.
		}

		v := checkOTLPServiceValue(configPath, psID, rule.expectedOTLPSuffix, rootDir, readFileFn)
		violations = append(violations, v...)
	}

	return violations
}

// checkServiceDir checks all config-*.yml files in a service directory for correct otlp-service names.
func checkServiceDir(serviceDir, psID, rootDir string, readFileFn func(string) ([]byte, error)) []string {
	var violations []string

	for _, rule := range standaloneConfigRules {
		filename := psID + rule.filenameSuffix
		configPath := filepath.Join(serviceDir, filename)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue // File does not exist — not a violation (file presence checked elsewhere).
		}

		v := checkOTLPServiceValue(configPath, psID, rule.expectedOTLPSuffix, rootDir, readFileFn)
		violations = append(violations, v...)
	}

	return violations
}

// checkOTLPServiceValue parses a config YAML and validates the otlp-service value.
func checkOTLPServiceValue(configPath, psID, expectedSuffix, rootDir string, readFileFn func(string) ([]byte, error)) []string {
	data, err := readFileFn(configPath) //nolint:gosec // configPath from controlled directory walk
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read file: %s", configPath, err)}
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return []string{fmt.Sprintf("%s: YAML parse error: %s", configPath, err)}
	}

	otlpServiceRaw, ok := config["otlp-service"]
	if !ok {
		return nil // No otlp-service key — not a violation (key is optional).
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
