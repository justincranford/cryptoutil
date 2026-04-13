// Copyright (c) 2025 Justin Cranford

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

	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// templatesRelPath is the path to the canonical templates directory relative to the project root.
const templatesRelPath = "api/cryptosuite-registry/templates"

// Shared constants for canonical template parameter values.
const (
	buildDate              = "2026-02-17T00:00:00Z"
	goVersion              = "1.26.1"
	alpineVersion          = "latest"
	cgoEnabled             = "0"
	containerUID           = "65532"
	containerGID           = "65532"
	githubRepoURL          = "https://github.com/justincranford/cryptoutil"
	authors                = "Justin Cranford"
	healthcheckInterval    = "30s"
	healthcheckTimeout     = "10s"
	healthcheckStartPeriod = "30s"
	healthcheckRetries     = "3"
)

// Expansion placeholder keys detected in template paths.
const (
	expansionKeyPSID    = "__PS_ID__"
	expansionKeyProduct = "__PRODUCT__"
	expansionKeySuite   = "__SUITE__"
)

// LoadTemplatesDir walks the canonical templates directory and returns a map of
// template-relative path → raw file content. Skips .gitkeep files.
func LoadTemplatesDir(projectRoot string) (map[string]string, error) {
	templatesDir := filepath.Join(projectRoot, templatesRelPath)

	if _, err := os.Stat(templatesDir); err != nil {
		return nil, fmt.Errorf("templates directory not found: %w", err)
	}

	templates := make(map[string]string)

	err := filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Skip .gitkeep files.
		if d.Name() == ".gitkeep" {
			return nil
		}

		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return fmt.Errorf("compute relative path for %s: %w", path, err)
		}

		// Normalize to forward slashes for cross-platform consistency.
		relPath = filepath.ToSlash(relPath)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", relPath, err)
		}

		templates[relPath] = string(content)

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
func BuildExpectedFS(templates map[string]string) (map[string]string, error) {
	expected := make(map[string]string)

	for tmplPath, tmplContent := range templates {
		switch {
		case strings.Contains(tmplPath, expansionKeyPSID):
			if err := expandPSIDTemplate(tmplPath, tmplContent, expected); err != nil {
				return nil, err
			}
		case strings.Contains(tmplPath, expansionKeyProduct):
			if err := expandProductTemplate(tmplPath, tmplContent, expected); err != nil {
				return nil, err
			}
		case strings.Contains(tmplPath, expansionKeySuite):
			if err := expandSuiteTemplate(tmplPath, tmplContent, expected); err != nil {
				return nil, err
			}
		default:
			// Static template: no path expansion, content-only substitution.
			actualPath := tmplPath
			content := substituteParams(tmplContent, buildStaticParams())
			expected[actualPath] = content
		}
	}

	return expected, nil
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
// Secrets files with BASE64_CHAR43 use length-based matching.
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
	case strings.Contains(expected, base64Char43Placeholder):
		return compareBase64Placeholder(expected, actual)
	default:
		return compareExact(
			normalizeCommentAlignment(expected),
			normalizeCommentAlignment(actual),
		)
	}
}

// expandPSIDTemplate expands a __PS_ID__ template for all 10 PS-IDs.
func expandPSIDTemplate(tmplPath, tmplContent string, expected map[string]string) error {
	for _, ps := range cryptoutilRegistry.AllProductServices() {
		params := buildParams(ps.PSID)
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)
		expected[actualPath] = content
	}

	return nil
}

// expandProductTemplate expands a __PRODUCT__ template for all 5 products.
func expandProductTemplate(tmplPath, tmplContent string, expected map[string]string) error {
	for _, product := range cryptoutilRegistry.AllProducts() {
		params := buildProductParams(product.ID)
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)
		expected[actualPath] = content
	}

	return nil
}

// expandSuiteTemplate expands a __SUITE__ template for the suite.
func expandSuiteTemplate(tmplPath, tmplContent string, expected map[string]string) error {
	for _, suite := range cryptoutilRegistry.AllSuites() {
		params := buildSuiteParams(suite.ID)
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)
		expected[actualPath] = content
	}

	return nil
}

// substituteParams replaces all __KEY__ placeholders in s with their values from params.
func substituteParams(s string, params map[string]string) string {
	result := s
	for placeholder, value := range params {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// buildParams constructs the full parameter map for PS-ID template instantiation.
func buildParams(psID string) map[string]string {
	basePort := cryptoutilRegistry.PublicPort(psID)

	return map[string]string{
		"__PS_ID__":                     psID,
		"__PS_ID_UPPER__":               strings.ToUpper(psID),
		"__PS_ID_UNDERSCORE__":          strings.ReplaceAll(psID, "-", "_"),
		"__SUITE__":                     cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		"__IMAGE_TAG__":                 cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
		"__BUILD_DATE__":                buildDate,
		"__GO_VERSION__":                goVersion,
		"__ALPINE_VERSION__":            alpineVersion,
		"__CGO_ENABLED__":               cgoEnabled,
		"__CONTAINER_UID__":             containerUID,
		"__CONTAINER_GID__":             containerGID,
		"__GITHUB_REPOSITORY_URL__":     githubRepoURL,
		"__AUTHORS__":                   authors,
		"__HEALTHCHECK_INTERVAL__":      healthcheckInterval,
		"__HEALTHCHECK_TIMEOUT__":       healthcheckTimeout,
		"__HEALTHCHECK_START_PERIOD__":  healthcheckStartPeriod,
		"__HEALTHCHECK_RETRIES__":       healthcheckRetries,
		"__PRODUCT_DISPLAY_NAME__":      cryptoutilRegistry.ProductDisplayName(psID),
		"__SERVICE_DISPLAY_NAME__":      cryptoutilRegistry.ServiceDisplayName(psID),
		"__SERVICE_APP_PORT_BASE__":     fmt.Sprintf("%d", basePort),
		"__SERVICE_APP_PORT_END__":      fmt.Sprintf("%d", cryptoutilRegistry.PortRangeEnd(psID)),
		"__SERVICE_APP_PORT_SQLITE_1__": fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetSQLite1),
		"__SERVICE_APP_PORT_SQLITE_2__": fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetSQLite2),
		"__SERVICE_APP_PORT_PG_1__":     fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetPostgres1),
		"__SERVICE_APP_PORT_PG_2__":     fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetPostgres2),
	}
}

// buildInstanceParams extends base params with per-instance values.
func buildInstanceParams(psID string, instanceNum int, port int) map[string]string {
	params := buildParams(psID)
	params["__INSTANCE_NUM__"] = fmt.Sprintf("%d", instanceNum)
	params["__SERVICE_APP_PORT__"] = fmt.Sprintf("%d", port)

	return params
}

// buildProductParams constructs the parameter map for product-level template instantiation.
func buildProductParams(productID string) map[string]string {
	psIDs := cryptoutilRegistry.PSIDsForProduct(productID)
	initPSID := cryptoutilRegistry.ProductInitPSID(productID)
	displayName := cryptoutilRegistry.ProductDisplayNameByID(productID)

	params := map[string]string{
		"__PRODUCT__":                    productID,
		"__PRODUCT_UPPER__":              strings.ToUpper(productID),
		"__PRODUCT_DISPLAY_NAME__":       displayName,
		"__PRODUCT_INIT_PS_ID__":         initPSID,
		"__PRODUCT_PS_ID_LIST_DISPLAY__": buildProductPSIDListDisplay(productID, psIDs),
		"__PRODUCT_INCLUDE_LIST__":       buildProductIncludeList(psIDs),
		"__PRODUCT_SERVICE_OVERRIDES__":  buildProductServiceOverrides(productID, psIDs),
		"__SUITE__":                      cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		"__IMAGE_TAG__":                  cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
	}

	return params
}

// buildSuiteParams constructs the parameter map for suite-level template instantiation.
func buildSuiteParams(sID string) map[string]string {
	initPSID := cryptoutilRegistry.SuiteInitPSID(sID)
	displayName := cryptoutilRegistry.SuiteDisplayName(sID)
	products := cryptoutilRegistry.AllProducts()

	params := map[string]string{
		"__SUITE__":                   sID,
		"__SUITE_UPPER__":             strings.ToUpper(sID),
		"__SUITE_DISPLAY_NAME__":      displayName,
		"__SUITE_INIT_PS_ID__":        initPSID,
		"__SUITE_INCLUDE_LIST__":      buildSuiteIncludeList(products),
		"__SUITE_SERVICE_OVERRIDES__": buildSuiteServiceOverrides(),
		"__IMAGE_TAG__":               cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
		// Build params for suite Dockerfile.
		"__BUILD_DATE__":               buildDate,
		"__GO_VERSION__":               goVersion,
		"__ALPINE_VERSION__":           alpineVersion,
		"__CGO_ENABLED__":              cgoEnabled,
		"__CONTAINER_UID__":            containerUID,
		"__CONTAINER_GID__":            containerGID,
		"__GITHUB_REPOSITORY_URL__":    githubRepoURL,
		"__AUTHORS__":                  authors,
		"__HEALTHCHECK_INTERVAL__":     healthcheckInterval,
		"__HEALTHCHECK_TIMEOUT__":      healthcheckTimeout,
		"__HEALTHCHECK_START_PERIOD__": healthcheckStartPeriod,
		"__HEALTHCHECK_RETRIES__":      healthcheckRetries,
	}

	return params
}

// buildStaticParams constructs the parameter map for static (non-expanded) templates.
// Static templates like shared-telemetry still need __SUITE__ substitution in content.
func buildStaticParams() map[string]string {
	return map[string]string{
		"__SUITE__":     cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		"__IMAGE_TAG__": cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
	}
}

// buildProductPSIDListDisplay builds the display string for product compose header comment.
// Format for SM (2 services): "sm-kms and sm-im"
// Format for Identity (5 services): "all 5 identity PS-IDs".
func buildProductPSIDListDisplay(productID string, psIDs []string) string {
	serviceCount := len(psIDs)
	if serviceCount == 0 {
		return ""
	}

	// Extract service names (strip product prefix).
	services := make([]string, len(psIDs))
	for i, psID := range psIDs {
		services[i] = strings.TrimPrefix(psID, productID+"-")
	}

	return fmt.Sprintf("%s (%d service%s: %s)",
		strings.ToUpper(productID)+" product",
		serviceCount,
		pluralS(serviceCount),
		strings.Join(services, ", "),
	)
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}

	return "s"
}

// buildProductIncludeList generates the Docker Compose include entries for a product.
// Format:
//
//	include:
//	  - path: ../sm-kms/compose.yml
//	  - path: ../sm-im/compose.yml
func buildProductIncludeList(psIDs []string) string {
	var sb strings.Builder

	sb.WriteString("include:\n")

	for _, psID := range psIDs {
		sb.WriteString(fmt.Sprintf("  - path: ../%s/compose.yml\n", psID))
	}

	return strings.TrimRight(sb.String(), "\n")
}

// buildProductServiceOverrides generates the port override blocks for a product compose.
// Product format (multi-line ports: !override):
//
//	sm-kms-app-sqlite-1:
//	  ports: !override
//	    - "18000:8080"
func buildProductServiceOverrides(productID string, psIDs []string) string {
	var sb strings.Builder

	variants := []struct {
		suffix string
		offset int
	}{
		{"sqlite-1", cryptoutilRegistry.ComposeVariantOffsetSQLite1},
		{"sqlite-2", cryptoutilRegistry.ComposeVariantOffsetSQLite2},
		{"postgresql-1", cryptoutilRegistry.ComposeVariantOffsetPostgres1},
		{"postgresql-2", cryptoutilRegistry.ComposeVariantOffsetPostgres2},
	}

	for _, psID := range psIDs {
		basePort := cryptoutilRegistry.PublicPort(psID)
		productPort := basePort + cryptoutilRegistry.PortTierOffsetProduct

		for _, v := range variants {
			port := productPort + v.offset
			sb.WriteString(fmt.Sprintf("  %s-app-%s:\n", psID, v.suffix))
			sb.WriteString("    ports: !override\n")
			sb.WriteString(fmt.Sprintf("      - \"%d:8080\"\n", port))
			sb.WriteString("\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// buildSuiteIncludeList generates the Docker Compose include entries for the suite.
func buildSuiteIncludeList(products []cryptoutilRegistry.Product) string {
	var sb strings.Builder

	sb.WriteString("include:\n")

	for _, p := range products {
		sb.WriteString(fmt.Sprintf("  - path: ../%s/compose.yml\n", p.ID))
	}

	return strings.TrimRight(sb.String(), "\n")
}

// buildSuiteServiceOverrides generates the inline port override blocks for a suite compose.
// Suite format (inline): sm-kms-app-sqlite-1: {ports: !override ["28000:8080"]}.
func buildSuiteServiceOverrides() string {
	var sb strings.Builder

	allPS := cryptoutilRegistry.AllProductServices()

	// Group by product for comments.
	currentProduct := ""

	for _, ps := range allPS {
		product := cryptoutilRegistry.ProductForPSID(ps.PSID)

		if product != currentProduct {
			if currentProduct != "" {
				sb.WriteString("\n")
			}

			basePort := cryptoutilRegistry.PublicPort(ps.PSID)
			endPort := basePort + cryptoutilRegistry.ComposeVariantOffsetPostgres2

			sb.WriteString(fmt.Sprintf("  # %s: SERVICE %d-%d -> SUITE %d-%d\n",
				strings.ToUpper(ps.PSID),
				basePort, endPort,
				basePort+cryptoutilRegistry.PortTierOffsetSuite,
				endPort+cryptoutilRegistry.PortTierOffsetSuite,
			))

			currentProduct = product
		} else {
			basePort := cryptoutilRegistry.PublicPort(ps.PSID)
			endPort := basePort + cryptoutilRegistry.ComposeVariantOffsetPostgres2

			sb.WriteString(fmt.Sprintf("\n  # %s: SERVICE %d-%d -> SUITE %d-%d\n",
				strings.ToUpper(ps.PSID),
				basePort, endPort,
				basePort+cryptoutilRegistry.PortTierOffsetSuite,
				endPort+cryptoutilRegistry.PortTierOffsetSuite,
			))
		}

		basePort := cryptoutilRegistry.PublicPort(ps.PSID)
		suitePort := basePort + cryptoutilRegistry.PortTierOffsetSuite

		variants := []struct {
			suffix string
			offset int
		}{
			{"sqlite-1", cryptoutilRegistry.ComposeVariantOffsetSQLite1},
			{"sqlite-2", cryptoutilRegistry.ComposeVariantOffsetSQLite2},
			{"postgresql-1", cryptoutilRegistry.ComposeVariantOffsetPostgres1},
			{"postgresql-2", cryptoutilRegistry.ComposeVariantOffsetPostgres2},
		}

		for _, v := range variants {
			port := suitePort + v.offset
			sb.WriteString(fmt.Sprintf("  %s-app-%s: {ports: !override [\"%d:8080\"]}\n", ps.PSID, v.suffix, port))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}
