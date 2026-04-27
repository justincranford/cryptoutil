// Copyright (c) 2025 Justin Cranford

// Package apps_product_template verifies that every product directory under internal/apps/{PRODUCT}/
// conforms to the canonical MANIFEST.yaml template. The template is read at runtime from
// api/cryptosuite-registry/templates/internal/apps/__PRODUCT__/MANIFEST.yaml.
//
// Placeholder substitution: __PRODUCT__ → product ID (e.g. "sm").
//
// See ENG-HANDBOOK.md Section 9.11.1 for the fitness sub-linter catalog.
package apps_product_template

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

// productManifest mirrors the YAML structure of the product MANIFEST.yaml template.
type productManifest struct {
	RequiredRootFiles []string `yaml:"required_root_files"`
}

// manifestRelPath is the path to the product MANIFEST.yaml relative to rootDir.
const manifestRelPath = "api/cryptosuite-registry/templates/internal/apps/" + cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct + "/MANIFEST.yaml"

// Check validates product template conformance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates product template conformance under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	manifestPath := filepath.Join(rootDir, filepath.FromSlash(manifestRelPath))

	data, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		return fmt.Errorf("failed to read product MANIFEST.yaml at %s: %w", manifestPath, readErr)
	}

	var manifest productManifest
	if unmarshalErr := yaml.Unmarshal(data, &manifest); unmarshalErr != nil {
		return fmt.Errorf("failed to parse product MANIFEST.yaml at %s: %w", manifestPath, unmarshalErr)
	}

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	var violations []string

	for _, product := range cryptoutilFitnessRegistry.AllProducts() {
		productDir := filepath.Join(appsDir, product.ID)

		for _, tmplFile := range manifest.RequiredRootFiles {
			expanded := strings.ReplaceAll(tmplFile, cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct, product.ID)

			if _, err := os.Stat(filepath.Join(productDir, expanded)); os.IsNotExist(err) {
				violations = append(violations, fmt.Sprintf("%s: missing required root file: %s", product.ID, expanded))
			}
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("apps product template violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("apps-product-template: all products pass template validation")

	return nil
}
