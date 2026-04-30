// Copyright (c) 2025-2026 Justin Cranford.
package dockerfile_labels

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
			input: `["/app/jose-ja"]`,
			want:  []string{"/app/jose-ja"},
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
			input: `/app/jose-ja`,
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
					t.Errorf("expected nil, got %v", got)
				}

				return
			}

			if len(got) != len(tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)

				return
			}

			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("element %d: expected %q, got %q", i, tc.want[i], got[i])
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
		{name: "equal single", a: []string{"/app/jose-ja"}, b: []string{"/app/jose-ja"}, want: true},
		{name: "equal multi", a: []string{"/sbin/tini", "--", "/app/cryptoutil"}, b: []string{"/sbin/tini", "--", "/app/cryptoutil"}, want: true},
		{name: "different length", a: []string{"/app/a"}, b: []string{"/app/a", "extra"}, want: false},
		{name: "different element", a: []string{"/app/a"}, b: []string{"/app/b"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := entrypointEqual(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("entrypointEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
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
		t.Errorf("expected %q, got %q", "unquoted-value", got)
	}
}

func TestParseLabelsFromLine_NoEqualSign(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	// No equal sign — should not panic, should skip.
	parseLabelsFromLine("no-equal-sign-here", labels)

	if len(labels) != 0 {
		t.Errorf("expected no labels, got %v", labels)
	}
}
