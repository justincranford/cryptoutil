// Copyright (c) 2025 Justin Cranford
//
//

package testutil_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

func TestCaptureOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fn       func()
		expected string
	}{
		{
			name: "captures stdout",
			fn: func() {
				_, _ = os.Stdout.WriteString("stdout output")
			},
			expected: "stdout output",
		},
		{
			name: "captures stderr",
			fn: func() {
				_, _ = os.Stderr.WriteString("stderr output")
			},
			expected: "stderr output",
		},
		{
			name: "captures both stdout and stderr",
			fn: func() {
				_, _ = os.Stdout.WriteString("stdout")
				_, _ = os.Stderr.WriteString("stderr")
			},
			expected: "stdout",
		},
		{
			name: "captures empty output",
			fn: func() {
				// Do nothing
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := cryptoutilSharedTestutil.CaptureOutput(t, tc.fn)
			require.Contains(t, output, tc.expected)
		})
	}
}

// TestContainsAny tests the ContainsAny function.
func TestContainsAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		haystack string
		needles  []string
		want     bool
	}{
		{
			name:     "found first needle",
			haystack: "hello world",
			needles:  []string{"hello", "foo", "bar"},
			want:     true,
		},
		{
			name:     "found middle needle",
			haystack: "hello world",
			needles:  []string{"foo", "world", "bar"},
			want:     true,
		},
		{
			name:     "found last needle",
			haystack: "hello world",
			needles:  []string{"foo", "bar", "world"},
			want:     true,
		},
		{
			name:     "no needle found",
			haystack: "hello world",
			needles:  []string{"foo", "bar", "baz"},
			want:     false,
		},
		{
			name:     "empty needles",
			haystack: "hello world",
			needles:  []string{},
			want:     false,
		},
		{
			name:     "empty haystack",
			haystack: "",
			needles:  []string{"hello"},
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := cryptoutilSharedTestutil.ContainsAny(tc.haystack, tc.needles)
			require.Equal(t, tc.want, got)
		})
	}
}
