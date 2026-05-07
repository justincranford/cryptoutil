// Copyright (c) 2025-2026 Justin Cranford.
// Package template_drift verifies that all deployment artifacts match their
// canonical templates after placeholder substitution. This catches structural drift
// between services' Dockerfiles, compose files, config overlays, and secrets.
// ENG-HANDBOOK.md Section 9.11.1 Fitness Sub-Linter Catalog.
package template_drift

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// walkDirFn is the function signature for walking a directory tree.
// Production uses filepath.WalkDir; tests may inject alternatives.
type walkDirFn func(root string, fn fs.WalkDirFunc) error

// LoadTemplatesDir walks the canonical templates directory and returns a map of
// template-relative path â†’ raw file content. Skips .gitkeep files.
func LoadTemplatesDir(projectRoot string) (map[string]string, error) {
	return loadTemplatesDirFn(projectRoot, filepath.WalkDir)
}

// loadTemplatesDirFn is the seam-injectable version of LoadTemplatesDir.
func loadTemplatesDirFn(projectRoot string, walkFn walkDirFn) (map[string]string, error) {
	templatesDir := filepath.Join(projectRoot, cryptoutilSharedMagic.CICDTemplatesRelPath)

	if _, err := os.Stat(templatesDir); err != nil {
		return nil, fmt.Errorf("templates directory not found: %w", err)
	}

	templates := make(map[string]string)

	err := walkFn(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil // Skip directories.
		} else if d.Name() == ".gitkeep" {
			return nil // Skip marker-only placeholder files.
		}

		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return fmt.Errorf("compute relative path for %s: %w", path, err)
		}

		// Normalize to forward slashes for cross-platform consistency.
		relPath = filepath.ToSlash(relPath)

		// Skip structural meta-files and cmd/internal Go source templates that are
		// validated by dedicated fitness linters rather than template-drift exact match.
		if isStructuralMetaFile(relPath) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", relPath, err)
		}

		// Strip the //go:build ignore header from Go source templates.
		// The actual project files do not carry this tag â€” it is only present in the
		// template copy to prevent the compiler from picking up placeholder-bearing files.
		templates[relPath] = stripBuildIgnoreTag(string(content))

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk templates directory: %w", err)
	}

	return templates, nil
}

// BuildExpectedFS expands all templates into an expected filesystem map.
// The returned map has actual-relative paths (relative to project root) as keys
// and expected file content as values.
func BuildExpectedFS(templates map[string]string) map[string]string {
	expected := make(map[string]string)

	for tmplPath, tmplContent := range templates {
		switch {
		case strings.Contains(tmplPath, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID):
			expandPSIDTemplate(tmplPath, tmplContent, expected)
		case strings.Contains(tmplPath, cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct):
			expandProductTemplate(tmplPath, tmplContent, expected)
		case strings.Contains(tmplPath, cryptoutilSharedMagic.CICDTemplateExpansionKeySuite):
			expandSuiteTemplate(tmplPath, tmplContent, expected)
		default:
			// Static template: no path expansion, content-only substitution.
			actualPath := tmplPath
			content := substituteParams(tmplContent, buildStaticParams())
			expected[actualPath] = content
		}
	}

	return expected
}

// CompareExpectedFS compares the expected filesystem against actual files on disk.
// Returns an aggregated error listing all mismatches; nil if everything matches.
func CompareExpectedFS(expected map[string]string, projectRoot string) error {
	var errs []string

	for relPath, expectedContent := range expected {
		actualPath := filepath.Join(projectRoot, filepath.FromSlash(relPath))

		actual, err := os.ReadFile(actualPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", relPath, err))

			continue
		}

		// Choose comparison strategy based on path.
		diff := chooseComparison(relPath, expectedContent, string(actual))
		if diff != "" {
			errs = append(errs, fmt.Sprintf("%s: content drift:\n%s", relPath, diff))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("template-compliance violations:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

// chooseComparison selects the comparison strategy for each file path.
// pki-ca compose uses superset (allows domain-specific extra volume mounts).
// pki-ca framework common config uses prefix (allows domain-specific CRL additions).
// Standalone configs use prefix (allows domain-specific additions after framework settings).
// Secrets files with __BASE64_CHAR43__ use length-based matching.
// All other files use exact comparison.
func chooseComparison(relPath, expected, actual string) string {
	normalized := filepath.ToSlash(relPath)

	switch {
	case strings.Contains(normalized, "deployments/pki-ca/compose.yml"):
		return compareSupersetOrdered(
			normalizeCommentAlignment(expected),
			normalizeCommentAlignment(actual),
		)
	case strings.Contains(normalized, "deployments/pki-ca/config/pki-ca-app-framework-common.yml"):
		return comparePrefix(expected, actual)
	case strings.HasPrefix(normalized, "configs/") && strings.HasSuffix(normalized, "-framework.yml"):
		return comparePrefix(expected, actual)
	case strings.Contains(expected, cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder):
		return compareBase64Placeholder(expected, actual)
	default:
		return compareExact(
			normalizeCommentAlignment(expected),
			normalizeCommentAlignment(actual),
		)
	}
}

// expandPSIDTemplate expands a __PS_ID__ template for all 10 PS-IDs.
func expandPSIDTemplate(tmplPath, tmplContent string, expected map[string]string) {
	for _, ps := range cryptoutilRegistry.AllProductServices() {
		params := buildParams(ps.PSID)
		addGoSourceParams(params, ps)
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)

		// Skip templates that still contain unresolved __PLACEHOLDER__ tokens after substitution.
		// This gracefully handles templates (e.g. __SERVICE__.go) that require additional
		// params not yet wired in â€” they should not produce false comparison failures.
		if hasUnresolvedPlaceholders(content) {
			continue
		}

		expected[actualPath] = content
	}
}

// expandProductTemplate expands a __PRODUCT__ template for all 5 products.
func expandProductTemplate(tmplPath, tmplContent string, expected map[string]string) {
	for _, product := range cryptoutilRegistry.AllProducts() {
		params := buildProductParams(product.ID)
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)
		expected[actualPath] = content
	}
}

// expandSuiteTemplate expands a __SUITE__ template for the suite.
func expandSuiteTemplate(tmplPath, tmplContent string, expected map[string]string) {
	for _, suite := range cryptoutilRegistry.AllSuites() {
		params := buildSuiteParams(suite.ID)
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)
		expected[actualPath] = content
	}
}

// substituteParams replaces all __KEY__ placeholders in s with their values from params.
func substituteParams(s string, params map[string]string) string {
	result := s
	for placeholder, value := range params {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
