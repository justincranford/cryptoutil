// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_template

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// serviceRootTemplateFilename is the canonical template filename for the root service entrypoint.
const serviceRootTemplateFilename = "__SERVICE__.go"

const (
	templateParseWithFlagSet      = "ParseWithFlagSet"
	templateNewFromConfig         = "NewFromConfig"
	templateIdentityProductName   = "IdentityProductName"
	templateEntryFuncTemplateName = "Template"
)

// checkServiceRootTemplate validates root __SERVICE__.go content against the canonical template.
func checkServiceRootTemplate(rootDir, psDir string, ps cryptoutilFitnessRegistry.ProductService) []string {
	templatePath := filepath.Join(
		rootDir,
		"api",
		"cryptosuite-registry",
		"templates",
		"internal",
		"apps",
		cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID,
		serviceRootTemplateFilename,
	)

	templateContent, templateReadErr := os.ReadFile(templatePath)
	if templateReadErr != nil {
		return []string{fmt.Sprintf("%s: failed to read SERVICE.go template at %s: %v", ps.PSID, templatePath, templateReadErr)}
	}

	templateValues, valuesErr := serviceRootTemplateValues(ps)
	if valuesErr != nil {
		return []string{fmt.Sprintf("%s: %v", ps.PSID, valuesErr)}
	}

	expectedContent := string(templateContent)
	expectedContent = applyTemplateValues(expectedContent, templateValues)

	actualPath := filepath.Join(psDir, ps.Service+".go")

	actualContent, actualReadErr := os.ReadFile(actualPath)
	if actualReadErr != nil {
		return []string{fmt.Sprintf("%s: failed to read root service file %s: %v", ps.PSID, actualPath, actualReadErr)}
	}

	normalizedExpected := normalizeTemplateContent(expectedContent)
	normalizedActual := normalizeTemplateContent(string(actualContent))

	if normalizedExpected != normalizedActual {
		lineNumber, expectedLine, actualLine := firstMismatchLine(normalizedExpected, normalizedActual)

		return []string{fmt.Sprintf(
			"%s: root service file %s does not match canonical template %s (line %d; expected=%q actual=%q)",
			ps.PSID,
			actualPath,
			templatePath,
			lineNumber,
			expectedLine,
			actualLine,
		)}
	}

	return nil
}

// firstMismatchLine returns the first differing 1-based line and both line values.
func firstMismatchLine(expectedContent, actualContent string) (int, string, string) {
	expectedLines := strings.Split(expectedContent, "\n")
	actualLines := strings.Split(actualContent, "\n")

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for index := 0; index < maxLines; index++ {
		expectedLine := ""
		actualLine := ""

		switch {
		case index < len(expectedLines):
			expectedLine = expectedLines[index]
		case index < len(actualLines):
			actualLine = actualLines[index]
		}

		if index < len(actualLines) && index < len(expectedLines) {
			actualLine = actualLines[index]
		}

		if expectedLine != actualLine {
			return index + 1, expectedLine, actualLine
		}
	}

	return 0, "", ""
}

// applyTemplateValues applies placeholder substitutions in descending key-length order
// to avoid collisions where one key is a substring of another.
func applyTemplateValues(templateContent string, templateValues map[string]string) string {
	keys := make([]string, 0, len(templateValues))
	for key := range templateValues {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	expanded := templateContent
	for _, key := range keys {
		expanded = strings.ReplaceAll(expanded, key, templateValues[key])
	}

	return expanded
}

// normalizeTemplateContent standardizes line endings for deterministic template comparisons.
func normalizeTemplateContent(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")

	lines := strings.Split(normalized, "\n")
	if len(lines) >= 2 && strings.TrimSpace(lines[0]) == "//go:build ignore" && strings.TrimSpace(lines[1]) == "" {
		lines = lines[2:]
	}

	joined := strings.Join(lines, "\n")

	return strings.TrimRight(joined, "\n")
}

// serviceRootTemplateValues returns placeholder substitutions for __SERVICE__.go template expansion.
func serviceRootTemplateValues(ps cryptoutilFitnessRegistry.ProductService) (map[string]string, error) {
	serviceConfigAlias := "cryptoutilAppsServiceServerConfig"
	serviceConfigImportPath := ""
	serviceServerImportPath := ""
	configBeforeTLS := false

	values := map[string]string{
		cryptoutilSharedMagic.CICDTemplateExpansionKeyService: ps.Service,
		cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID:    ps.PSID,
	}

	switch ps.PSID {
	case cryptoutilSharedMagic.OTLPServiceSMIM:
		serviceConfigImportPath = "cryptoutil/internal/apps/sm-im/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/sm-im/server"
		values["__ENTRY_FUNC__"] = "Im"
		values["__SERVICE_ID_CONST__"] = "IMServiceID"
		values["__PRODUCT_NAME_CONST__"] = "IMProductName"
		values["__SERVICE_NAME_CONST__"] = "IMServiceName"
		values["__SERVICE_PORT_CONST__"] = "IMServicePort"
		values["__USAGE_PREFIX__"] = "IM"
		values["__SERVER_SETTINGS_TYPE__"] = "SmIMServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = "NewIMServerFromConfig"
	case cryptoutilSharedMagic.OTLPServiceSMKMS:
		serviceConfigImportPath = "cryptoutil/internal/apps-framework/service/config"
		serviceServerImportPath = "cryptoutil/internal/apps/sm-kms/server"
		serviceConfigAlias = "cryptoutilAppsFrameworkServiceConfig"
		configBeforeTLS = true
		values["__ENTRY_FUNC__"] = "Kms"
		values["__SERVICE_ID_CONST__"] = "KMSServiceID"
		values["__PRODUCT_NAME_CONST__"] = "SMProductName"
		values["__SERVICE_NAME_CONST__"] = "KMSServiceName"
		values["__SERVICE_PORT_CONST__"] = "KMSServicePort"
		values["__USAGE_PREFIX__"] = "KMS"
		values["__SERVER_SETTINGS_TYPE__"] = "ServiceFrameworkServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = "NewKMSServerFromConfig"
	case cryptoutilSharedMagic.OTLPServiceJoseJA:
		serviceConfigImportPath = "cryptoutil/internal/apps/jose-ja/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/jose-ja/server"
		values["__ENTRY_FUNC__"] = "Ja"
		values["__SERVICE_ID_CONST__"] = "JoseJAServiceID"
		values["__PRODUCT_NAME_CONST__"] = "JoseProductName"
		values["__SERVICE_NAME_CONST__"] = "JoseJAServiceName"
		values["__SERVICE_PORT_CONST__"] = "JoseJAServicePort"
		values["__USAGE_PREFIX__"] = "JA"
		values["__SERVER_SETTINGS_TYPE__"] = "JoseJAServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	case cryptoutilSharedMagic.OTLPServicePKICA:
		serviceConfigImportPath = "cryptoutil/internal/apps/pki-ca/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/pki-ca/server"
		values["__ENTRY_FUNC__"] = "Ca"
		values["__SERVICE_ID_CONST__"] = "PKICAServiceID"
		values["__PRODUCT_NAME_CONST__"] = "PKIProductName"
		values["__SERVICE_NAME_CONST__"] = "PKICAServiceName"
		values["__SERVICE_PORT_CONST__"] = "PKICAServicePort"
		values["__USAGE_PREFIX__"] = "CA"
		values["__SERVER_SETTINGS_TYPE__"] = "CAServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	case cryptoutilSharedMagic.OTLPServiceIdentityAuthz:
		serviceConfigImportPath = "cryptoutil/internal/apps/identity-authz/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/identity-authz/server"
		values["__ENTRY_FUNC__"] = "Authz"
		values["__SERVICE_ID_CONST__"] = "IdentityAuthzServiceID"
		values["__PRODUCT_NAME_CONST__"] = templateIdentityProductName
		values["__SERVICE_NAME_CONST__"] = "AuthzServiceName"
		values["__SERVICE_PORT_CONST__"] = "IdentityAuthzServicePort"
		values["__USAGE_PREFIX__"] = "AUTHZ"
		values["__SERVER_SETTINGS_TYPE__"] = "IdentityAuthzServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	case cryptoutilSharedMagic.OTLPServiceIdentityIDP:
		serviceConfigImportPath = "cryptoutil/internal/apps/identity-idp/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/identity-idp/server"
		values["__ENTRY_FUNC__"] = "Idp"
		values["__SERVICE_ID_CONST__"] = "IdentityIDPServiceID"
		values["__PRODUCT_NAME_CONST__"] = templateIdentityProductName
		values["__SERVICE_NAME_CONST__"] = "IDPServiceName"
		values["__SERVICE_PORT_CONST__"] = "IdentityIDPServicePort"
		values["__USAGE_PREFIX__"] = "IDP"
		values["__SERVER_SETTINGS_TYPE__"] = "IdentityIDPServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	case cryptoutilSharedMagic.OTLPServiceIdentityRP:
		serviceConfigImportPath = "cryptoutil/internal/apps/identity-rp/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/identity-rp/server"
		values["__ENTRY_FUNC__"] = "Rp"
		values["__SERVICE_ID_CONST__"] = "IdentityRPServiceID"
		values["__PRODUCT_NAME_CONST__"] = templateIdentityProductName
		values["__SERVICE_NAME_CONST__"] = "RPServiceName"
		values["__SERVICE_PORT_CONST__"] = "IdentityRPServicePort"
		values["__USAGE_PREFIX__"] = "RP"
		values["__SERVER_SETTINGS_TYPE__"] = "IdentityRPServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	case cryptoutilSharedMagic.OTLPServiceIdentityRS:
		serviceConfigImportPath = "cryptoutil/internal/apps/identity-rs/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/identity-rs/server"
		values["__ENTRY_FUNC__"] = "Rs"
		values["__SERVICE_ID_CONST__"] = "IdentityRSServiceID"
		values["__PRODUCT_NAME_CONST__"] = templateIdentityProductName
		values["__SERVICE_NAME_CONST__"] = "RSServiceName"
		values["__SERVICE_PORT_CONST__"] = "IdentityRSServicePort"
		values["__USAGE_PREFIX__"] = "RS"
		values["__SERVER_SETTINGS_TYPE__"] = "IdentityRSServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	case cryptoutilSharedMagic.OTLPServiceIdentitySPA:
		serviceConfigImportPath = "cryptoutil/internal/apps/identity-spa/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/identity-spa/server"
		values["__ENTRY_FUNC__"] = "Spa"
		values["__SERVICE_ID_CONST__"] = "IdentitySPAServiceID"
		values["__PRODUCT_NAME_CONST__"] = "IdentityProductName"
		values["__SERVICE_NAME_CONST__"] = "SPAServiceName"
		values["__SERVICE_PORT_CONST__"] = "IdentitySPAServicePort"
		values["__USAGE_PREFIX__"] = "SPA"
		values["__SERVER_SETTINGS_TYPE__"] = "IdentitySPAServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = "ParseWithFlagSet"
		values["__NEW_SERVER_CONSTRUCTOR__"] = "NewFromConfig"
	case cryptoutilSharedMagic.OTLPServiceSkeletonTemplate:
		serviceConfigImportPath = "cryptoutil/internal/apps/skeleton-template/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/skeleton-template/server"
		values["__ENTRY_FUNC__"] = templateEntryFuncTemplateName
		values["__SERVICE_ID_CONST__"] = "SkeletonTemplateServiceID"
		values["__PRODUCT_NAME_CONST__"] = "SkeletonProductName"
		values["__SERVICE_NAME_CONST__"] = "SkeletonTemplateServiceName"
		values["__SERVICE_PORT_CONST__"] = "SkeletonTemplateServicePort"
		values["__USAGE_PREFIX__"] = "Template"
		values["__SERVER_SETTINGS_TYPE__"] = "SkeletonTemplateServerSettings"
		values["__PARSE_CONFIG_FUNC__"] = templateParseWithFlagSet
		values["__NEW_SERVER_CONSTRUCTOR__"] = templateNewFromConfig
	default:
		return nil, fmt.Errorf("unsupported PS-ID for SERVICE.go template expansion")
	}

	values["__SERVICE_CONFIG_ALIAS__"] = serviceConfigAlias

	tlsImportAlias := "cryptoutilAppsFrameworkTls"
	tlsImportPath := "cryptoutil/internal/apps-framework/tls"
	serverImportAlias := "cryptoutilAppsServiceServer"
	values["__SERVER_ALIAS__"] = serverImportAlias

	if configBeforeTLS {
		values["__FIRST_APP_IMPORT_ALIAS__"] = serviceConfigAlias
		values["__FIRST_APP_IMPORT_PATH__"] = serviceConfigImportPath
		values["__SECOND_APP_IMPORT_ALIAS__"] = tlsImportAlias
		values["__SECOND_APP_IMPORT_PATH__"] = tlsImportPath
		values["__THIRD_APP_IMPORT_ALIAS__"] = serverImportAlias
		values["__THIRD_APP_IMPORT_PATH__"] = serviceServerImportPath
	} else {
		values["__FIRST_APP_IMPORT_ALIAS__"] = tlsImportAlias
		values["__FIRST_APP_IMPORT_PATH__"] = tlsImportPath
		values["__SECOND_APP_IMPORT_ALIAS__"] = serverImportAlias
		values["__SECOND_APP_IMPORT_PATH__"] = serviceServerImportPath
		values["__THIRD_APP_IMPORT_ALIAS__"] = serviceConfigAlias
		values["__THIRD_APP_IMPORT_PATH__"] = serviceConfigImportPath
	}

	return values, nil
}
