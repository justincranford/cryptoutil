package template_drift

import (
	"fmt"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// base64Char43Placeholder is the placeholder used in secret templates to indicate
// a position where a base64-encoded value of at least 43 characters should appear.
const base64Char43Placeholder = "BASE64_CHAR43"

// base64Char43MinLength is the minimum length for a BASE64_CHAR43 value.
const base64Char43MinLength = cryptoutilSharedMagic.DefaultCodeChallengeLength

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

// compareBase64Placeholder validates that the actual content matches the expected pattern,
// where BASE64_CHAR43 placeholders are replaced with values of at least 43 characters.
// The expected string is split on BASE64_CHAR43 markers. Each segment between markers
// must appear in sequence in the actual string, and the actual content between them
// must be at least 43 characters long.
func compareBase64Placeholder(expected, actual string) string {
	expected = strings.TrimRight(normalizeLineEndings(expected), "\n")
	actual = strings.TrimRight(normalizeLineEndings(actual), "\n")

	parts := strings.Split(expected, base64Char43Placeholder)

	// For each segment boundary: verify the fixed parts appear in order
	// and the gaps between them are ≥ 43 chars.
	remaining := actual

	for i, part := range parts {
		if part == "" && i > 0 {
			// Trailing empty part after final BASE64_CHAR43: validate remaining length.
			if len(remaining) < base64Char43MinLength {
				return fmt.Sprintf("  BASE64_CHAR43 value too short: got %d chars, need ≥ %d",
					len(remaining), base64Char43MinLength)
			}

			continue
		}

		idx := strings.Index(remaining, part)
		if idx < 0 {
			return fmt.Sprintf("  expected fixed segment %q not found in actual", part)
		}

		if i > 0 && idx < base64Char43MinLength {
			return fmt.Sprintf("  BASE64_CHAR43 value too short: got %d chars, need ≥ %d",
				idx, base64Char43MinLength)
		}

		remaining = remaining[idx+len(part):]
	}

	return ""
}
