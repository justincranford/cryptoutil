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

// walkDirFn is the function signature for walking a directory tree.
// Production uses filepath.WalkDir; tests may inject alternatives.
type walkDirFn func(root string, fn fs.WalkDirFunc) error

// LoadTemplatesDir walks the canonical templates directory and returns a map of
// template-relative path → raw file content. Skips .gitkeep files.
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
		actualPath := substituteParams(tmplPath, params)
		content := substituteParams(tmplContent, params)
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

// buildParams constructs the full parameter map for PS-ID template instantiation.
func buildParams(psID string) map[string]string {
	basePort := cryptoutilRegistry.PublicPort(psID)

	return map[string]string{
		cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID: psID,
		"__PS_ID_UPPER__":      strings.ToUpper(psID),
		"__PS_ID_UNDERSCORE__": strings.ReplaceAll(psID, "-", "_"),
		cryptoutilSharedMagic.CICDTemplateExpansionKeySuite: cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		"__IMAGE_TAG__":                cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
		"__BUILD_DATE__":               cryptoutilSharedMagic.CICDTemplateBuildDate,
		"__GO_VERSION__":               cryptoutilSharedMagic.CICDTemplateGoVersion,
		"__ALPINE_VERSION__":           cryptoutilSharedMagic.CICDTemplateAlpineVersion,
		"__CGO_ENABLED__":              cryptoutilSharedMagic.CICDTemplateCGOEnabled,
		"__CONTAINER_UID__":            cryptoutilSharedMagic.CICDTemplateContainerUID,
		"__CONTAINER_GID__":            cryptoutilSharedMagic.CICDTemplateContainerGID,
		"__GITHUB_REPOSITORY_URL__":    cryptoutilSharedMagic.CICDTemplateGitHubRepoURL,
		"__AUTHORS__":                  cryptoutilSharedMagic.CICDTemplateAuthors,
		"__HEALTHCHECK_INTERVAL__":     cryptoutilSharedMagic.CICDTemplateHealthcheckInterval,
		"__HEALTHCHECK_TIMEOUT__":      cryptoutilSharedMagic.CICDTemplateHealthcheckTimeout,
		"__HEALTHCHECK_START_PERIOD__": cryptoutilSharedMagic.CICDTemplateHealthcheckStartPeriod,
		"__HEALTHCHECK_RETRIES__":      cryptoutilSharedMagic.CICDTemplateHealthcheckRetries,
		"__PRODUCT_DISPLAY_NAME__":     cryptoutilRegistry.ProductDisplayName(psID),
		"__PS_DISPLAY_NAME__":          cryptoutilRegistry.ServiceDisplayName(psID),
		"__PS_PUBLIC_PORT_BASE__":      fmt.Sprintf("%d", basePort),
		"__PS_PUBLIC_PORT_END__":       fmt.Sprintf("%d", cryptoutilRegistry.PortRangeEnd(psID)),
		"__PS_PUBLIC_PORT_SQLITE_1__":  fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetSQLite1),
		"__PS_PUBLIC_PORT_SQLITE_2__":  fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetSQLite2),
		"__PS_PUBLIC_PORT_PG_1__":      fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetPostgres1),
		"__PS_PUBLIC_PORT_PG_2__":      fmt.Sprintf("%d", basePort+cryptoutilRegistry.ComposeVariantOffsetPostgres2),
	}
}

// buildInstanceParams extends base params with per-instance values.
func buildInstanceParams(psID string, instanceNum int, port int) map[string]string {
	params := buildParams(psID)
	params["__INSTANCE_NUM__"] = fmt.Sprintf("%d", instanceNum)
	params["__PS_PUBLIC_PORT__"] = fmt.Sprintf("%d", port)

	return params
}

// buildProductParams constructs the parameter map for product-level template instantiation.
func buildProductParams(productID string) map[string]string {
	psIDs := cryptoutilRegistry.PSIDsForProduct(productID)
	initPSID := cryptoutilRegistry.ProductInitPSID(productID)
	displayName := cryptoutilRegistry.ProductDisplayNameByID(productID)

	params := map[string]string{
		cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct: productID,
		"__PRODUCT_UPPER__":                                 strings.ToUpper(productID),
		"__PRODUCT_DISPLAY_NAME__":                          displayName,
		"__PRODUCT_INIT_PS_ID__":                            initPSID,
		"__PRODUCT_PS_ID_LIST_DISPLAY__":                    buildProductPSIDListDisplay(productID, psIDs),
		"__PRODUCT_INCLUDE_LIST__":                          buildProductIncludeList(psIDs),
		"__PRODUCT_SERVICE_OVERRIDES__":                     buildProductServiceOverrides(productID, psIDs),
		cryptoutilSharedMagic.CICDTemplateExpansionKeySuite: cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		"__IMAGE_TAG__":                                     cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
	}

	return params
}

// buildSuiteParams constructs the parameter map for suite-level template instantiation.
func buildSuiteParams(sID string) map[string]string {
	initPSID := cryptoutilRegistry.SuiteInitPSID(sID)
	displayName := cryptoutilRegistry.SuiteDisplayName(sID)
	products := cryptoutilRegistry.AllProducts()

	return map[string]string{
		cryptoutilSharedMagic.CICDTemplateExpansionKeySuite: sID,
		"__SUITE_UPPER__":              strings.ToUpper(sID),
		"__SUITE_DISPLAY_NAME__":       displayName,
		"__SUITE_INIT_PS_ID__":         initPSID,
		"__SUITE_INCLUDE_LIST__":       buildSuiteIncludeList(products),
		"__SUITE_SERVICE_OVERRIDES__":  buildSuiteServiceOverrides(),
		"__IMAGE_TAG__":                cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
		"__BUILD_DATE__":               cryptoutilSharedMagic.CICDTemplateBuildDate,
		"__GO_VERSION__":               cryptoutilSharedMagic.CICDTemplateGoVersion,
		"__ALPINE_VERSION__":           cryptoutilSharedMagic.CICDTemplateAlpineVersion,
		"__CGO_ENABLED__":              cryptoutilSharedMagic.CICDTemplateCGOEnabled,
		"__CONTAINER_UID__":            cryptoutilSharedMagic.CICDTemplateContainerUID,
		"__CONTAINER_GID__":            cryptoutilSharedMagic.CICDTemplateContainerGID,
		"__GITHUB_REPOSITORY_URL__":    cryptoutilSharedMagic.CICDTemplateGitHubRepoURL,
		"__AUTHORS__":                  cryptoutilSharedMagic.CICDTemplateAuthors,
		"__HEALTHCHECK_INTERVAL__":     cryptoutilSharedMagic.CICDTemplateHealthcheckInterval,
		"__HEALTHCHECK_TIMEOUT__":      cryptoutilSharedMagic.CICDTemplateHealthcheckTimeout,
		"__HEALTHCHECK_START_PERIOD__": cryptoutilSharedMagic.CICDTemplateHealthcheckStartPeriod,
		"__HEALTHCHECK_RETRIES__":      cryptoutilSharedMagic.CICDTemplateHealthcheckRetries,
	}
}

// buildStaticParams constructs the parameter map for static (non-expanded) templates.
// Static templates like shared-telemetry still need __SUITE__ substitution in content.
func buildStaticParams() map[string]string {
	return map[string]string{
		cryptoutilSharedMagic.CICDTemplateExpansionKeySuite: cryptoutilSharedMagic.DefaultOTLPServiceDefault,
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
		fmt.Fprintf(&sb, "  - path: ../%s/compose.yml\n", psID)
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
		{cryptoutilSharedMagic.CICDTemplateVariantSQLite1, cryptoutilRegistry.ComposeVariantOffsetSQLite1},
		{cryptoutilSharedMagic.CICDTemplateVariantSQLite2, cryptoutilRegistry.ComposeVariantOffsetSQLite2},
		{cryptoutilSharedMagic.CICDTemplateVariantPostgres1, cryptoutilRegistry.ComposeVariantOffsetPostgres1},
		{cryptoutilSharedMagic.CICDTemplateVariantPostgres2, cryptoutilRegistry.ComposeVariantOffsetPostgres2},
	}

	for _, psID := range psIDs {
		basePort := cryptoutilRegistry.PublicPort(psID)
		productPort := basePort + cryptoutilRegistry.PortTierOffsetProduct

		for _, v := range variants {
			port := productPort + v.offset
			fmt.Fprintf(&sb, "  %s-app-%s:\n", psID, v.suffix)
			sb.WriteString("    ports: !override\n")
			fmt.Fprintf(&sb, "      - \"%d:%d\"\n", port, cryptoutilSharedMagic.DockerContainerPublicHTTPSPort)
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
		fmt.Fprintf(&sb, "  - path: ../%s/compose.yml\n", p.ID)
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

			fmt.Fprintf(&sb, "  # %s: PS-PUBLIC %d-%d -> SUITE %d-%d\n",
				strings.ToUpper(ps.PSID),
				basePort, endPort,
				basePort+cryptoutilRegistry.PortTierOffsetSuite,
				endPort+cryptoutilRegistry.PortTierOffsetSuite)

			currentProduct = product
		} else {
			basePort := cryptoutilRegistry.PublicPort(ps.PSID)
			endPort := basePort + cryptoutilRegistry.ComposeVariantOffsetPostgres2

			fmt.Fprintf(&sb, "\n  # %s: PS-PUBLIC %d-%d -> SUITE %d-%d\n",
				strings.ToUpper(ps.PSID),
				basePort, endPort,
				basePort+cryptoutilRegistry.PortTierOffsetSuite,
				endPort+cryptoutilRegistry.PortTierOffsetSuite)
		}

		basePort := cryptoutilRegistry.PublicPort(ps.PSID)
		suitePort := basePort + cryptoutilRegistry.PortTierOffsetSuite

		variants := []struct {
			suffix string
			offset int
		}{
			{cryptoutilSharedMagic.CICDTemplateVariantSQLite1, cryptoutilRegistry.ComposeVariantOffsetSQLite1},
			{cryptoutilSharedMagic.CICDTemplateVariantSQLite2, cryptoutilRegistry.ComposeVariantOffsetSQLite2},
			{cryptoutilSharedMagic.CICDTemplateVariantPostgres1, cryptoutilRegistry.ComposeVariantOffsetPostgres1},
			{cryptoutilSharedMagic.CICDTemplateVariantPostgres2, cryptoutilRegistry.ComposeVariantOffsetPostgres2},
		}

		for _, v := range variants {
			port := suitePort + v.offset
			fmt.Fprintf(&sb, "  %s-app-%s: {ports: !override [\"%d:%d\"]}\n", ps.PSID, v.suffix, port, cryptoutilSharedMagic.DockerContainerPublicHTTPSPort)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}
