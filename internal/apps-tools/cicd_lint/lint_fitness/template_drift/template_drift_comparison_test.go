// Copyright (c) 2025-2026 Justin Cranford.
package template_drift

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestNormalizeLineEndings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "lf only", input: testABC, want: testABC},
		{name: "crlf to lf", input: testABCValueD37A6C0B, want: testABC},
		{name: "mixed line endings", input: "a\r\nb\nc\r\n", want: "a\nb\nc\n"},
		{name: cryptoutilSharedMagic.EMPTY, input: "", want: ""},
		{name: "no newlines", input: testAbc, want: testAbc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeLineEndings(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeCommentAlignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "comment multi spaces", input: "# foo  bar", want: "# foo bar"},
		{name: "comment triple spaces", input: "# a   b   c", want: "# a b c"},
		{name: "non-comment preserved", input: testKeyValue, want: testKeyValue},
		{name: "mixed lines", input: "# a  b\nkey:  val\n# c   d", want: "# a b\nkey:  val\n# c d"},
		{name: cryptoutilSharedMagic.EMPTY, input: "", want: ""},
		{name: "single space comment", input: testOk, want: testOk},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeCommentAlignment(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCompareExact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
		actual   string
		wantDiff bool
	}{
		{name: "identical", expected: testABC, actual: testABC, wantDiff: false},
		{name: "identical crlf normalize", expected: testABC, actual: testABCValueD37A6C0B, wantDiff: false},
		{name: "trailing newline trim", expected: "a\nb\n", actual: testAB, wantDiff: false},
		{name: "different line", expected: testABC, actual: "a\nX\nc", wantDiff: true},
		{name: "extra actual line", expected: testAB, actual: testABC, wantDiff: true},
		{name: "missing actual line", expected: testABC, actual: testAB, wantDiff: true},
		{name: "empty both", expected: "", actual: "", wantDiff: false},
		{name: testEmptyExpected, expected: "", actual: testAValueCA978112, wantDiff: true},
		{name: "empty actual", expected: testAValueCA978112, actual: "", wantDiff: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diff := compareExact(tt.expected, tt.actual)

			if tt.wantDiff {
				require.NotEmpty(t, diff, "expected diff but got none")
			} else {
				require.Empty(t, diff, "expected no diff but got: %s", diff)
			}
		})
	}
}

func TestCompareExact_DiffContent(t *testing.T) {
	t.Parallel()

	t.Run("shows line number for mismatch", func(t *testing.T) {
		t.Parallel()

		diff := compareExact(testABC, "a\nX\nc")
		require.Contains(t, diff, "line 2")
		require.Contains(t, diff, "want")
		require.Contains(t, diff, "got")
	})

	t.Run("shows extra line message", func(t *testing.T) {
		t.Parallel()

		diff := compareExact(testAValueCA978112, testAB)
		require.Contains(t, diff, "unexpected extra line")
	})

	t.Run("shows missing line message", func(t *testing.T) {
		t.Parallel()

		diff := compareExact(testAB, testAValueCA978112)
		require.Contains(t, diff, "missing expected line")
	})
}

func TestComparePrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
		actual   string
		wantDiff bool
	}{
		{name: testExactMatch, expected: testAB, actual: testAB, wantDiff: false},
		{name: "actual has extra suffix", expected: testAB, actual: "a\nb\nc\nd", wantDiff: false},
		{name: "prefix mismatch line 1", expected: testAB, actual: "X\nb", wantDiff: true},
		{name: "prefix mismatch line 2", expected: testAB, actual: "a\nX", wantDiff: true},
		{name: "actual too short", expected: testABC, actual: testAB, wantDiff: true},
		{name: testEmptyExpected, expected: "", actual: testAB, wantDiff: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diff := comparePrefix(tt.expected, tt.actual)

			if tt.wantDiff {
				require.NotEmpty(t, diff, "expected diff but got none")
			} else {
				require.Empty(t, diff, "expected no diff but got: %s", diff)
			}
		})
	}
}

func TestComparePrefix_DiffContent(t *testing.T) {
	t.Parallel()

	t.Run("shows line count when too short", func(t *testing.T) {
		t.Parallel()

		diff := comparePrefix(testABC, testAValueCA978112)
		require.Contains(t, diff, "1 lines")
		require.Contains(t, diff, "at least 3")
	})

	t.Run("shows line diff for mismatch", func(t *testing.T) {
		t.Parallel()

		diff := comparePrefix(testAB, "a\nX")
		require.Contains(t, diff, "line 2")
		require.Contains(t, diff, "want")
		require.Contains(t, diff, "got")
	})
}

func TestCompareSupersetOrdered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
		actual   string
		wantDiff bool
	}{
		{name: testExactMatch, expected: testABC, actual: testABC, wantDiff: false},
		{name: "actual has extra interspersed", expected: testAC, actual: testABC, wantDiff: false},
		{name: "missing expected line", expected: testABC, actual: testAC, wantDiff: true},
		{name: "wrong order", expected: testAB, actual: "b\na", wantDiff: true},
		{name: "all present with extras", expected: "x\nz", actual: "w\nx\ny\nz\nend", wantDiff: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diff := compareSupersetOrdered(tt.expected, tt.actual)

			if tt.wantDiff {
				require.NotEmpty(t, diff, "expected diff but got none")
			} else {
				require.Empty(t, diff, "expected no diff but got: %s", diff)
			}
		})
	}
}

func TestCompareSupersetOrdered_DiffContent(t *testing.T) {
	t.Parallel()

	diff := compareSupersetOrdered(testABC, testAC)
	require.Contains(t, diff, "missing expected line")
	require.Contains(t, diff, "b")
}

func TestCompareBase64Placeholder(t *testing.T) {
	t.Parallel()

	longB64 := testBase64String // 47 chars

	tests := []struct {
		name     string
		expected string
		actual   string
		wantDiff bool
	}{
		{
			name:     "valid replacement",
			expected: "prefix-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-suffix",
			actual:   "prefix-" + longB64 + "-suffix",
			wantDiff: false,
		},
		{
			name:     "too short replacement",
			expected: "prefix-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-suffix",
			actual:   "prefix-SHORT-suffix",
			wantDiff: true,
		},
		{
			name:     "missing fixed segment",
			expected: "prefix-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-suffix",
			actual:   "wrong-" + longB64 + "-suffix",
			wantDiff: true,
		},
		{
			name:     "multiple placeholders",
			expected: "a-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-b-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-c",
			actual:   "a-" + longB64 + "-b-" + longB64 + "-c",
			wantDiff: false,
		},
		{
			name:     "no placeholder exact match",
			expected: testNoPlaceholder,
			actual:   testNoPlaceholder,
			wantDiff: false,
		},
		{
			name:     "no placeholder mismatch",
			expected: "expected",
			actual:   "actual",
			wantDiff: true,
		},
		{
			name:     "exactly 43 chars",
			expected: "x-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-y",
			actual:   "x-" + longB64[:cryptoutilSharedMagic.DefaultCodeChallengeLength] + "-y",
			wantDiff: false,
		},
		{
			name:     "42 chars too short",
			expected: "x-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder + "-y",
			actual:   "x-" + longB64[:cryptoutilSharedMagic.AnswerToLifeUniverseEverything] + "-y",
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diff := compareBase64Placeholder(tt.expected, tt.actual)

			if tt.wantDiff {
				require.NotEmpty(t, diff, "expected diff but got none")
			} else {
				require.Empty(t, diff, "expected no diff but got: %s", diff)
			}
		})
	}
}
