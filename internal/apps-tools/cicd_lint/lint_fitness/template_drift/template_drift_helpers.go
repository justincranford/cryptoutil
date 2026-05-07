// Copyright (c) 2025-2026 Justin Cranford.
// Package template_drift — helper utilities for placeholder detection and Go template stripping.
package template_drift

import (
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// unresolvedPlaceholderRE matches any remaining __UPPER_SNAKE_CASE__ token.
var unresolvedPlaceholderRE = regexp.MustCompile(`__[A-Z][A-Z0-9_]*__`)

// hasUnresolvedPlaceholders reports whether s still contains any unresolved __PLACEHOLDER__
// tokens after parameter substitution. The __BASE64_CHAR43__ sentinel is excluded — it is a
// legitimate comparison placeholder kept in expected content for base64 field matching.
// Templates with remaining unresolved tokens are skipped to avoid false comparison failures.
func hasUnresolvedPlaceholders(s string) bool {
	// Temporarily remove the base64 sentinel before checking for unresolved placeholders.
	cleaned := strings.ReplaceAll(s, cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder, "")

	return unresolvedPlaceholderRE.MatchString(cleaned)
}

// isStructuralMetaFile reports whether relPath should be excluded from content comparison.
// Two categories are excluded:
//  1. Pure meta-files (MANIFEST.yaml, README.md) that guide Phase 4 linters with no
//     corresponding actual project file.
//  2. Go source templates in cmd/ and internal/, which are validated by dedicated
//     fitness linters and are not template-drift exact-match targets.
func isStructuralMetaFile(relPath string) bool {
	base := filepath.ToSlash(filepath.Base(relPath))

	if base == "MANIFEST.yaml" || base == "README.md" {
		return true
	}

	// Exclude all Go source templates from cmd/ and internal/ from template-drift
	// exact-match checks.
	if strings.HasPrefix(relPath, "cmd/") || strings.HasPrefix(relPath, "internal/") {
		return true
	}

	return false
}

// stripBuildIgnoreTag removes the //go:build ignore header line (and any immediately
// following blank line) from Go source template content.
// Canonical templates carry //go:build ignore to prevent the compiler from picking up
// placeholder-bearing files. The actual project files do NOT have this tag.
func stripBuildIgnoreTag(content string) string {
	const buildIgnoreLine = "//go:build ignore\n"

	after, found := strings.CutPrefix(content, buildIgnoreLine)
	if !found {
		return content
	}

	// Also strip a single blank line that typically follows the build tag.
	after, _ = strings.CutPrefix(after, "\n")

	return after
}
