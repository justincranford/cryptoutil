// Copyright (c) 2025-2026 Justin Cranford.
package dockerfile_labels

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
)

func TestParseEntrypointLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single binary",
			input: `["/app/sm-kms"]`,
			want:  []string{"/app/sm-kms"},
		},
		{
			name:  "tini with suite binary no args",
			input: `["/sbin/tini", "--", "/app/cryptoutil"]`,
			want:  []string{"/sbin/tini", "--", "/app/cryptoutil"},
		},
		{
			name:  "tini with suite binary and subcommand",
			input: `["/sbin/tini", "--", "/app/cryptoutil", "identity-authz", "start"]`,
			want:  []string{"/sbin/tini", "--", "/app/cryptoutil", cryptoutilSharedMagic.OTLPServiceIdentityAuthz, "start"},
		},
		{
			name:  "not json array returns nil",
			input: `/app/sm-kms`,
			want:  nil,
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parseEntrypointLine(tc.input)

			if tc.want == nil {
				if got != nil {
					assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected nil, got %v", got)
				}

				return
			}

			if len(got) != len(tc.want) {
				assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected %v, got %v", tc.want, got)

				return
			}

			for i := range tc.want {
				if got[i] != tc.want[i] {
					assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "element %d: expected %q, got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

func TestEntrypointEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{name: "equal single", a: []string{"/app/sm-kms"}, b: []string{"/app/sm-kms"}, want: true},
		{name: "equal multi", a: []string{"/sbin/tini", "--", "/app/cryptoutil"}, b: []string{"/sbin/tini", "--", "/app/cryptoutil"}, want: true},
		{name: "different length", a: []string{"/app/a"}, b: []string{"/app/a", "extra"}, want: false},
		{name: "different element", a: []string{"/app/a"}, b: []string{"/app/b"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := entrypointEqual(tc.a, tc.b)
			if got != tc.want {
				assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "entrypointEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestParseLabelsFromLine_UnquotedNoSpace(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	// Unquoted value with no trailing space (last key=value pair).
	parseLabelsFromLine("key=unquoted-value", labels)

	got, ok := labels["key"]
	if !ok {
		t.Fatal("expected key to be present")
	}

	if got != "unquoted-value" {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected %q, got %q", "unquoted-value", got)
	}
}

func TestParseLabelsFromLine_NoEqualSign(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	// No equal sign — should not panic, should skip.
	parseLabelsFromLine("no-equal-sign-here", labels)

	if len(labels) != 0 {
		assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "expected no labels, got %v", labels)
	}
}
