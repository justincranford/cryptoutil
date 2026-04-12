// Copyright (c) 2025 Justin Cranford

// Package template_drift verifies that all PS-ID deployment artifacts match their
// canonical templates after placeholder substitution. This catches structural drift
// between services' Dockerfiles, compose files, and config overlays.
// ENG-HANDBOOK.md Section 9.11.1 Fitness Sub-Linter Catalog.
package template_drift

import (
	"embed"
	"fmt"
	"regexp"
	"strings"

	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

//go:embed templates/*
var templatesFS embed.FS

// Shared constants for canonical template parameter values.
const (
	suiteID                = cryptoutilSharedMagic.DefaultOTLPServiceDefault
	imageTag               = cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault
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
	pkiCAPSID              = cryptoutilSharedMagic.OTLPServicePKICA
)

// buildParams constructs the full parameter map for template instantiation.
func buildParams(psID string) map[string]string {
	basePort := cryptoutilRegistry.PublicPort(psID)

	return map[string]string{
		"__PS_ID__":                     psID,
		"__PS_ID_UPPER__":               strings.ToUpper(psID),
		"__SUITE__":                     suiteID,
		"__IMAGE_TAG__":                 imageTag,
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

// instantiateFn is the function signature for template instantiation.
// Production code uses the default instantiate; tests can inject alternatives.
type instantiateFn func(templateName string, params map[string]string) (string, error)

// instantiate loads a template file and replaces all placeholders with values.
func instantiate(templateName string, params map[string]string) (string, error) {
	content, err := templatesFS.ReadFile("templates/" + templateName)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", templateName, err)
	}

	result := string(content)
	for placeholder, value := range params {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// normalizeLineEndings converts CRLF to LF for consistent comparison.
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// multiSpaceRegexp matches two or more consecutive spaces.
var multiSpaceRegexp = regexp.MustCompile(`  +`)

// normalizeCommentAlignment collapses runs of spaces into a single space
// in compose header comment lines. This handles column-aligned comments
// where padding varies based on PS-ID length.
func normalizeCommentAlignment(s string) string {
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			lines[i] = multiSpaceRegexp.ReplaceAllString(line, " ")
		}
	}

	return strings.Join(lines, "\n")
}

// compareExact returns a diff description if expected and actual differ, or empty string if equal.
func compareExact(expected, actual string) string {
	expected = strings.TrimRight(normalizeLineEndings(expected), "\n")
	actual = strings.TrimRight(normalizeLineEndings(actual), "\n")

	if expected == actual {
		return ""
	}

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var diffs []string

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := range maxLines {
		switch {
		case i >= len(expectedLines):
			diffs = append(diffs, fmt.Sprintf("  line %d: unexpected extra line: %q", i+1, actualLines[i]))
		case i >= len(actualLines):
			diffs = append(diffs, fmt.Sprintf("  line %d: missing expected line: %q", i+1, expectedLines[i]))
		case expectedLines[i] != actualLines[i]:
			diffs = append(diffs, fmt.Sprintf("  line %d:\n    want: %q\n    got:  %q", i+1, expectedLines[i], actualLines[i]))
		}
	}

	return strings.Join(diffs, "\n")
}

// comparePrefix checks that actual starts with expected content.
// Extra lines after the expected prefix are allowed (domain-specific additions).
func comparePrefix(expected, actual string) string {
	expected = strings.TrimRight(normalizeLineEndings(expected), "\n")
	actual = strings.TrimRight(normalizeLineEndings(actual), "\n")

	if strings.HasPrefix(actual, expected) {
		return ""
	}

	// Fall back to line-by-line comparison of the prefix portion.
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	if len(actualLines) < len(expectedLines) {
		return fmt.Sprintf("  actual file has %d lines, expected at least %d", len(actualLines), len(expectedLines))
	}

	var diffs []string

	for i, expectedLine := range expectedLines {
		if actualLines[i] != expectedLine {
			diffs = append(diffs, fmt.Sprintf("  line %d:\n    want: %q\n    got:  %q", i+1, expectedLine, actualLines[i]))
		}
	}

	return strings.Join(diffs, "\n")
}

// compareSupersetOrdered checks that actual contains all expected lines in order,
// allowing extra lines interspersed (e.g., pki-ca extra volume mounts).
func compareSupersetOrdered(expected, actual string) string {
	expected = strings.TrimRight(normalizeLineEndings(expected), "\n")
	actual = strings.TrimRight(normalizeLineEndings(actual), "\n")

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	actualIdx := 0

	for _, expectedLine := range expectedLines {
		found := false

		for actualIdx < len(actualLines) {
			if actualLines[actualIdx] == expectedLine {
				found = true
				actualIdx++

				break
			}

			actualIdx++
		}

		if !found {
			return fmt.Sprintf("  missing expected line: %q", expectedLine)
		}
	}

	return ""
}
