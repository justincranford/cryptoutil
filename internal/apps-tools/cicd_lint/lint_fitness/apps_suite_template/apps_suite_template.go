// Copyright (c) 2025-2026 Justin Cranford.
// Package apps_suite_template verifies that the suite directory under internal/apps/{SUITE}/
// conforms to the canonical MANIFEST.yaml template. The template is read at runtime from
// api/cryptosuite-registry/templates/internal/apps/__SUITE__/MANIFEST.yaml.
//
// Placeholder substitution: __SUITE__ → suite ID (currently "cryptoutil").
//
// This linter supersedes apps-suite-required-files, which is retired when this linter is registered.
//
// See ENG-HANDBOOK.md Section 9.11.1 for the fitness sub-linter catalog.
package apps_suite_template

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

// suiteManifest mirrors the YAML structure of the suite MANIFEST.yaml template.
type suiteManifest struct {
	RequiredRootFiles []string `yaml:"required_root_files"`
}

// manifestRelPath is the path to the suite MANIFEST.yaml relative to rootDir.
const manifestRelPath = "api/cryptosuite-registry/templates/internal/apps/" + cryptoutilSharedMagic.CICDTemplateExpansionKeySuite + "/MANIFEST.yaml"

// Check validates suite template conformance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates suite template conformance under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	manifestPath := filepath.Join(rootDir, filepath.FromSlash(manifestRelPath))

	data, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		return fmt.Errorf("failed to read suite MANIFEST.yaml at %s: %w", manifestPath, readErr)
	}

	var manifest suiteManifest
	if unmarshalErr := yaml.Unmarshal(data, &manifest); unmarshalErr != nil {
		return fmt.Errorf("failed to parse suite MANIFEST.yaml at %s: %w", manifestPath, unmarshalErr)
	}

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var violations []string

	for _, suite := range cryptoutilFitnessRegistry.AllSuites() {
		suiteDir := filepath.Join(appsDir, suite.ID)

		for _, tmplFile := range manifest.RequiredRootFiles {
			expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeySuite, suite.ID)

			if _, err := os.Stat(filepath.Join(suiteDir, expanded)); os.IsNotExist(err) {
				violations = append(violations, fmt.Sprintf("%s: missing required root file: %s", suite.ID, expanded))
			}
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("apps suite template violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("apps-suite-template: all suites pass template validation")

	return nil
}
