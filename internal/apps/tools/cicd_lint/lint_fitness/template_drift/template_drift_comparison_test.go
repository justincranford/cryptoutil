// Copyright (c) 2025 Justin Cranford

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
		{name: "lf only", input: "a\nb\nc", want: "a\nb\nc"},
		{name: "crlf to lf", input: "a\r\nb\r\nc", want: "a\nb\nc"},
		{name: "mixed", input: "a\r\nb\nc\r\n", want: "a\nb\nc\n"},
		{name: "empty", input: "", want: ""},
		{name: "no newlines", input: "abc", want: "abc"},
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
		{name: "non-comment preserved", input: "key:  value", want: "key:  value"},
		{name: "mixed lines", input: "# a  b\nkey:  val\n# c   d", want: "# a b\nkey:  val\n# c d"},
		{name: "empty", input: "", want: ""},
		{name: "single space comment", input: "# ok", want: "# ok"},
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
		{name: "identical", expected: "a\nb\nc", actual: "a\nb\nc", wantDiff: false},
		{name: "identical crlf normalize", expected: "a\nb\nc", actual: "a\r\nb\r\nc", wantDiff: false},
		{name: "trailing newline trim", expected: "a\nb\n", actual: "a\nb", wantDiff: false},
		{name: "different line", expected: "a\nb\nc", actual: "a\nX\nc", wantDiff: true},
		{name: "extra actual line", expected: "a\nb", actual: "a\nb\nc", wantDiff: true},
		{name: "missing actual line", expected: "a\nb\nc", actual: "a\nb", wantDiff: true},
		{name: "empty both", expected: "", actual: "", wantDiff: false},
		{name: "empty expected", expected: "", actual: "a", wantDiff: true},
		{name: "empty actual", expected: "a", actual: "", wantDiff: true},
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

		diff := compareExact("a\nb\nc", "a\nX\nc")
		require.Contains(t, diff, "line 2")
		require.Contains(t, diff, "want")
		require.Contains(t, diff, "got")
	})

	t.Run("shows extra line message", func(t *testing.T) {
		t.Parallel()

		diff := compareExact("a", "a\nb")
		require.Contains(t, diff, "unexpected extra line")
	})

	t.Run("shows missing line message", func(t *testing.T) {
		t.Parallel()

		diff := compareExact("a\nb", "a")
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
		{name: "exact match", expected: "a\nb", actual: "a\nb", wantDiff: false},
		{name: "actual has extra suffix", expected: "a\nb", actual: "a\nb\nc\nd", wantDiff: false},
		{name: "prefix mismatch line 1", expected: "a\nb", actual: "X\nb", wantDiff: true},
		{name: "prefix mismatch line 2", expected: "a\nb", actual: "a\nX", wantDiff: true},
		{name: "actual too short", expected: "a\nb\nc", actual: "a\nb", wantDiff: true},
		{name: "empty expected", expected: "", actual: "a\nb", wantDiff: false},
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

		diff := comparePrefix("a\nb\nc", "a")
		require.Contains(t, diff, "1 lines")
		require.Contains(t, diff, "at least 3")
	})

	t.Run("shows line diff for mismatch", func(t *testing.T) {
		t.Parallel()

		diff := comparePrefix("a\nb", "a\nX")
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
		{name: "exact match", expected: "a\nb\nc", actual: "a\nb\nc", wantDiff: false},
		{name: "actual has extra interspersed", expected: "a\nc", actual: "a\nb\nc", wantDiff: false},
		{name: "missing expected line", expected: "a\nb\nc", actual: "a\nc", wantDiff: true},
		{name: "wrong order", expected: "a\nb", actual: "b\na", wantDiff: true},
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

	diff := compareSupersetOrdered("a\nb\nc", "a\nc")
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
			expected: "prefix-BASE64_CHAR43-suffix",
			actual:   "prefix-" + longB64 + "-suffix",
			wantDiff: false,
		},
		{
			name:     "too short replacement",
			expected: "prefix-BASE64_CHAR43-suffix",
			actual:   "prefix-SHORT-suffix",
			wantDiff: true,
		},
		{
			name:     "missing fixed segment",
			expected: "prefix-BASE64_CHAR43-suffix",
			actual:   "wrong-" + longB64 + "-suffix",
			wantDiff: true,
		},
		{
			name:     "multiple placeholders",
			expected: "a-BASE64_CHAR43-b-BASE64_CHAR43-c",
			actual:   "a-" + longB64 + "-b-" + longB64 + "-c",
			wantDiff: false,
		},
		{
			name:     "no placeholder exact match",
			expected: "no placeholder",
			actual:   "no placeholder",
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
			expected: "x-BASE64_CHAR43-y",
			actual:   "x-" + longB64[:cryptoutilSharedMagic.DefaultCodeChallengeLength] + "-y",
			wantDiff: false,
		},
		{
			name:     "42 chars too short",
			expected: "x-BASE64_CHAR43-y",
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
