// Copyright (c) 2025-2026 Justin Cranford.
// Package template_drift — parameter builders for placeholder substitution.
package template_drift

import (
	"fmt"
	"strings"

	cryptoutilRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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

// addGoSourceParams injects Go source-code-specific placeholder values into params.
// These are used by internal/apps/__PS_ID__/*.go templates that reference service-level
// constants such as __SERVICE__, __USAGE_PREFIX__, and __PRODUCT_NAME_CONST__.
func addGoSourceParams(params map[string]string, ps cryptoutilRegistry.ProductService) {
	p := ps.GoTemplateParams
	params[cryptoutilSharedMagic.CICDTemplateExpansionKeyService] = ps.Service
	params["__USAGE_PREFIX__"] = p.UsagePrefix
	params["__PRODUCT_NAME_CONST__"] = p.ProductNameConst
	params["__SERVICE_NAME_CONST__"] = p.ServiceNameConst
	params["__SERVICE_ID_CONST__"] = p.ServiceIDConst
	params["__SERVICE_PORT_CONST__"] = p.ServicePortConst
	params["__SERVICE_DISPLAY_NAME_CONST__"] = p.ServiceDisplayNameConst
}
