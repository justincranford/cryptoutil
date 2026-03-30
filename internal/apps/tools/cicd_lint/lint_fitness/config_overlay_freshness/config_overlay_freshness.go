// Copyright (c) 2025 Justin Cranford

// Package config_overlay_freshness validates deployment config overlay files against
// canonical templates. Each PS-ID under deployments/{ps-id}/config/ must have
// the correct YAML keys and value patterns for each of its 4 variant files:
//
//   - {ps-id}-app-sqlite-1.yml    -> database-url must be present, value must match ^sqlite://
//   - {ps-id}-app-sqlite-2.yml    -> database-url must be present, value must match ^sqlite://
//   - {ps-id}-app-postgresql-1.yml -> database-url must be absent
//   - {ps-id}-app-postgresql-2.yml -> database-url must be absent
//
// The rules are loaded from the embedded config-overlay-templates.yaml.
// See ARCHITECTURE.md Section 9.11 for config overlay naming and content conventions.
package config_overlay_freshness

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

//go:embed config-overlay-templates.yaml
var overlayTemplatesYAML []byte

// overlayTemplates holds the parsed config-overlay-templates.yaml content.
type overlayTemplates struct {
	Variants []variantTemplate `yaml:"variants"`
}

// variantTemplate describes required/forbidden keys and value patterns for one variant.
type variantTemplate struct {
	Variant          string            `yaml:"variant"`
	Description      string            `yaml:"description"`
	RequiredKeys     []string          `yaml:"required_keys"`
	ForbiddenKeys    []string          `yaml:"forbidden_keys"`
	RequiredPatterns []requiredPattern `yaml:"required_patterns"`
}

// requiredPattern maps a YAML key to a regex pattern that its value must match.
type requiredPattern struct {
	Key         string `yaml:"key"`
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
}

// variantSuffixes maps variant name to the corresponding deployment config file suffix.
var variantSuffixes = map[string]string{
	lintFitnessRegistry.ComposeVariantSQLite1:   lintFitnessRegistry.DeploymentConfigSuffixSQLite1,
	lintFitnessRegistry.ComposeVariantSQLite2:   lintFitnessRegistry.DeploymentConfigSuffixSQLite2,
	lintFitnessRegistry.ComposeVariantPostgres1: lintFitnessRegistry.DeploymentConfigSuffixPostgresql1,
	lintFitnessRegistry.ComposeVariantPostgres2: lintFitnessRegistry.DeploymentConfigSuffixPostgresql2,
}

// Injectable OS functions for test seam injection.
var (
	overlayReadFileFn = os.ReadFile
)

// loadOverlayTemplates parses the embedded config-overlay-templates.yaml.
func loadOverlayTemplates() (*overlayTemplates, error) {
	var tmpl overlayTemplates
	if err := yaml.Unmarshal(overlayTemplatesYAML, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse config-overlay-templates.yaml: %w", err)
	}

	return &tmpl, nil
}

// Check validates deployment config overlay files from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates deployment config overlay files under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking deployment config overlay freshness...")

	tmpl, err := loadOverlayTemplates()
	if err != nil {
		return err
	}

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkPSIDOverlays(rootDir, ps.PSID, tmpl)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("config overlay freshness violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("config-overlay-freshness: all deployment config overlays conform to templates")

	return nil
}

// checkPSIDOverlays validates all 4 variant overlay files for a single PS-ID.
// If the deployments/{ps-id}/config/ directory does not exist, the PS-ID is skipped.
func checkPSIDOverlays(rootDir, psID string, tmpl *overlayTemplates) []string {
	var violations []string

	configDir := filepath.Join(rootDir, "deployments", psID, "config")

	// Skip PS-IDs that have no deployments config directory in this workspace root.
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return nil
	}

	for _, vt := range tmpl.Variants {
		suffix, ok := variantSuffixes[vt.Variant]
		if !ok {
			violations = append(violations, fmt.Sprintf("%s: unknown variant %q in template", psID, vt.Variant))

			continue
		}

		filename := psID + suffix
		configPath := filepath.Join(configDir, filename)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s: missing overlay file: %s", psID, filename))

			continue
		}

		v := checkOverlayFile(configPath, psID, filename, &vt)
		violations = append(violations, v...)
	}

	return violations
}

// checkOverlayFile validates a single overlay file against its variant template.
func checkOverlayFile(configPath, psID, filename string, vt *variantTemplate) []string {
	data, err := overlayReadFileFn(configPath) //nolint:gosec // configPath from controlled directory walk
	if err != nil {
		return []string{fmt.Sprintf("%s: %s: cannot read file: %s", psID, filename, err)}
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return []string{fmt.Sprintf("%s: %s: YAML parse error: %s", psID, filename, err)}
	}

	if config == nil {
		config = map[string]any{}
	}

	var violations []string

	// Check required keys are present.
	for _, key := range vt.RequiredKeys {
		if _, ok := config[key]; !ok {
			violations = append(violations, fmt.Sprintf("%s: %s: missing required key %q", psID, filename, key))
		}
	}

	// Check forbidden keys are absent.
	for _, key := range vt.ForbiddenKeys {
		if _, ok := config[key]; ok {
			violations = append(violations, fmt.Sprintf("%s: %s: forbidden key %q must not be present", psID, filename, key))
		}
	}

	// Check required value patterns.
	for _, rp := range vt.RequiredPatterns {
		val, ok := config[rp.Key]
		if !ok {
			// Missing required key is already reported above; skip pattern check.
			continue
		}

		strVal, isStr := val.(string)
		if !isStr {
			violations = append(violations, fmt.Sprintf("%s: %s: key %q must be a string, got %T", psID, filename, rp.Key, val))

			continue
		}

		matched, regexpErr := regexp.MatchString(rp.Pattern, strVal)
		if regexpErr != nil {
			violations = append(violations, fmt.Sprintf("%s: %s: invalid pattern %q for key %q: %s", psID, filename, rp.Pattern, rp.Key, regexpErr))

			continue
		}

		if !matched {
			violations = append(violations, fmt.Sprintf("%s: %s: key %q value %q does not match pattern %q (%s)", psID, filename, rp.Key, strVal, rp.Pattern, rp.Description))
		}
	}

	return violations
}
