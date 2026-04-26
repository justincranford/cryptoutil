// Copyright (c) 2025 Justin Cranford

package realm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSanitizeSchemaName_CharacterBoundaries kills CONDITIONALS_BOUNDARY mutants
// at tenant.go:267 on the character range checks (r >= 'a' && r <= 'z', etc.).
func TestSanitizeSchemaName_CharacterBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		// Lowercase boundary: 'a' and 'z' should be kept; chars just outside replaced.
		{
			name:   "lowercase a boundary",
			input:  "x" + string(rune('a')) + "x",
			expect: "xax",
		},
		{
			name:   "lowercase z boundary",
			input:  "x" + string(rune('z')) + "x",
			expect: "xzx",
		},
		{
			name:   "char before a replaced",
			input:  "x" + string(rune('a'-1)) + "x",
			expect: "x_x",
		},
		{
			name:   "char after z replaced",
			input:  "x" + string(rune('z'+1)) + "x",
			expect: "x_x",
		},
		// Uppercase boundary: 'A' and 'Z' should be kept (lowercased).
		{
			name:   "uppercase A boundary",
			input:  "x" + string(rune('A')) + "x",
			expect: "xax",
		},
		{
			name:   "uppercase Z boundary",
			input:  "x" + string(rune('Z')) + "x",
			expect: "xzx",
		},
		{
			name:   "char before A replaced",
			input:  "x" + string(rune('A'-1)) + "x",
			expect: "x_x",
		},
		{
			name:   "char after Z replaced",
			input:  "x" + string(rune('Z'+1)) + "x",
			expect: "x_x",
		},
		// Digit boundary: '0' and '9' should be kept.
		{
			name:   "digit 0 boundary",
			input:  "x" + string(rune('0')) + "x",
			expect: "x0x",
		},
		{
			name:   "digit 9 boundary",
			input:  "x" + string(rune('9')) + "x",
			expect: "x9x",
		},
		{
			name:   "char before 0 replaced",
			input:  "x" + string(rune('0'-1)) + "x",
			expect: "x_x",
		},
		{
			name:   "char after 9 replaced",
			input:  "x" + string(rune('9'+1)) + "x",
			expect: "x_x",
		},
		// Underscore kept.
		{
			name:   "underscore kept",
			input:  "x_x",
			expect: "x_x",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := sanitizeSchemaName(tc.input)
			require.Equal(t, tc.expect, result)
		})
	}
}

// TestSanitizeSchemaName_StartsWithDigitBoundary kills CONDITIONALS_BOUNDARY
// at tenant.go:275 on `safe[0] >= '0' && safe[0] <= '9'` and `len(safe) > 0`.
func TestSanitizeSchemaName_StartsWithDigitBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		hasPrefix bool
	}{
		{
			name:      "starts with 0",
			input:     "0abc",
			hasPrefix: true,
		},
		{
			name:      "starts with 9",
			input:     "9abc",
			hasPrefix: true,
		},
		{
			name:      "starts with letter a",
			input:     "abc",
			hasPrefix: false,
		},
		{
			name:      "starts with underscore",
			input:     "_abc",
			hasPrefix: false,
		},
		{
			name:      "empty string does not panic",
			input:     "",
			hasPrefix: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := sanitizeSchemaName(tc.input)
			if tc.hasPrefix {
				require.True(t, strings.HasPrefix(result, "t_"),
					"digit-starting name should get t_ prefix, got: %s", result)
			} else {
				require.False(t, strings.HasPrefix(result, "t_"),
					"non-digit-starting name should NOT get t_ prefix, got: %s", result)
			}
		})
	}
}

// TestSanitizeSchemaName_LengthBoundary kills CONDITIONALS_BOUNDARY
// at tenant.go:281 on `len(safe) > maxSchemaNameLength` where max=63.
func TestSanitizeSchemaName_LengthBoundary(t *testing.T) {
	t.Parallel()

	const maxSchemaNameLength = 63

	tests := []struct {
		name      string
		inputLen  int
		expectLen int
	}{
		{
			name:      "exactly at max length",
			inputLen:  maxSchemaNameLength,
			expectLen: maxSchemaNameLength,
		},
		{
			name:      "one over max length",
			inputLen:  maxSchemaNameLength + 1,
			expectLen: maxSchemaNameLength,
		},
		{
			name:      "one under max length",
			inputLen:  maxSchemaNameLength - 1,
			expectLen: maxSchemaNameLength - 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := strings.Repeat("a", tc.inputLen)
			result := sanitizeSchemaName(input)
			require.Len(t, result, tc.expectLen)
		})
	}
}
