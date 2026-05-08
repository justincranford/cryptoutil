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

// templateBaseDir is the canonical templates directory for PS-ID application files.
const templateBaseDir = "api/cryptosuite-registry/templates/internal/apps/__PS_ID__"

// templateFileSpec maps one canonical template file to one concrete PS-ID file path.
type templateFileSpec struct {
	templatePath string
	actualPath   string
	displayName  string
}

var canonicalTemplateFiles = []templateFileSpec{
	{templatePath: "__SERVICE__.go.tmpl", actualPath: "__SERVICE__.go", displayName: "service command wiring"},
	{templatePath: "__SERVICE___test.go.tmpl", actualPath: "__SERVICE___test.go", displayName: "CLI tests"},
	{templatePath: "server/__SERVICE___port_conflict_test.go.tmpl", actualPath: "server/__SERVICE___port_conflict_test.go", displayName: "port conflict tests"},
	{templatePath: "client/client.go.tmpl", actualPath: "client/client.go", displayName: "typed client"},
}

const (
	templateParseWithFlagSet      = "ParseWithFlagSet"
	templateNewFromConfig         = "NewFromConfig"
	templateIdentityProductName   = "IdentityProductName"
	templateEntryFuncTemplateName = "Template"
	templateServerImportAlias     = "cryptoutilAppsServiceServer"
)

// checkRootTemplates validates canonical template conformance for root and client files.
func checkRootTemplates(rootDir, psDir string, ps cryptoutilFitnessRegistry.ProductService) []string {
	templateValues, valuesErr := serviceRootTemplateValues(ps)
	if valuesErr != nil {
		return []string{fmt.Sprintf("%s: %v", ps.PSID, valuesErr)}
	}

	var violations []string

	for _, spec := range canonicalTemplateFiles {
		templatePath := filepath.Join(rootDir, filepath.FromSlash(templateBaseDir), filepath.FromSlash(spec.templatePath))
		violations = append(violations, checkSingleTemplateFile(templatePath, psDir, spec, ps, templateValues)...)
	}

	return violations
}

func checkSingleTemplateFile(
	templatePath string,
	psDir string,
	spec templateFileSpec,
	ps cryptoutilFitnessRegistry.ProductService,
	templateValues map[string]string,
) []string {
	templateContent, templateReadErr := os.ReadFile(templatePath)
	if templateReadErr != nil {
		return []string{fmt.Sprintf("%s: failed to read %s template at %s: %v", ps.PSID, spec.displayName, templatePath, templateReadErr)}
	}

	expectedContent := applyTemplateValues(string(templateContent), templateValues)
	actualRelPath := applyTemplateValues(spec.actualPath, templateValues)
	actualPath := filepath.Join(psDir, filepath.FromSlash(actualRelPath))

	actualContent, actualReadErr := os.ReadFile(actualPath)
	if actualReadErr != nil {
		return []string{fmt.Sprintf("%s: failed to read %s file %s: %v", ps.PSID, spec.displayName, actualPath, actualReadErr)}
	}

	normalizedExpected := normalizeTemplateContent(expectedContent)
	normalizedActual := normalizeTemplateContent(string(actualContent))

	if normalizedExpected == normalizedActual {
		return nil
	}

	lineNumber, expectedLine, actualLine := firstMismatchLine(normalizedExpected, normalizedActual)

	return []string{fmt.Sprintf(
		"%s: %s file %s does not match canonical template %s (line %d; expected=%q actual=%q)",
		ps.PSID,
		spec.displayName,
		actualPath,
		templatePath,
		lineNumber,
		expectedLine,
		actualLine,
	)}
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
	joined = strings.TrimLeft(joined, "\n")

	return strings.TrimRight(joined, "\n")
}

// serviceRootTemplateValues returns placeholder substitutions for __SERVICE__.go template expansion.
func serviceRootTemplateValues(ps cryptoutilFitnessRegistry.ProductService) (map[string]string, error) {
	serviceConfigAlias := "cryptoutilAppsServiceServerConfig"
	serviceConfigImportPath := ""
	serviceServerImportPath := ""
	configBeforeServer := false

	values := map[string]string{
		cryptoutilSharedMagic.CICDTemplateExpansionKeyService: ps.Service,
		cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID:    ps.PSID,
	}

	switch ps.PSID {
	case cryptoutilSharedMagic.OTLPServiceSMIM:
		serviceConfigImportPath = "cryptoutil/internal/apps/sm-im/server/config"
		serviceServerImportPath = "cryptoutil/internal/apps/sm-im/server"
		values["__ENTRY_FUNC__"] = "Im"
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "IMDisplayName"
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
		configBeforeServer = true
		values["__ENTRY_FUNC__"] = "Kms"
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "KMSDisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "JoseJADisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "PKICADisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "AuthzDisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "IDPDisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "RPDisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "RSDisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "SPADisplayName"
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
		values["__SERVICE_DISPLAY_NAME_CONST__"] = "TemplateDisplayName"
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
	values["__SERVER_ALIAS__"] = templateServerImportAlias

	if configBeforeServer {
		// sm-kms: framework config (alphabetically before apps/) comes before server
		values["__FIRST_APP_IMPORT_ALIAS__"] = serviceConfigAlias
		values["__FIRST_APP_IMPORT_PATH__"] = serviceConfigImportPath
		values["__SECOND_APP_IMPORT_ALIAS__"] = templateServerImportAlias
		values["__SECOND_APP_IMPORT_PATH__"] = serviceServerImportPath
	} else {
		// all other services: server comes before server/config alphabetically
		values["__FIRST_APP_IMPORT_ALIAS__"] = templateServerImportAlias
		values["__FIRST_APP_IMPORT_PATH__"] = serviceServerImportPath
		values["__SECOND_APP_IMPORT_ALIAS__"] = serviceConfigAlias
		values["__SECOND_APP_IMPORT_PATH__"] = serviceConfigImportPath
	}

	return values, nil
}
