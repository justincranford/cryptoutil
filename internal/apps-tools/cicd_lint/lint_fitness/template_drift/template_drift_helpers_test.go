// Copyright (c) 2025-2026 Justin Cranford.
package template_drift

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestIsStructuralMetaFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		relPath string
		want    bool
	}{
		{name: "manifest yaml is meta", relPath: "internal/apps/__PS_ID__/MANIFEST.yaml", want: true},
		{name: "readme md is meta", relPath: "internal/apps/__PS_ID__/README.md", want: true},
		{name: "usage go in internal is not meta", relPath: "internal/apps/__PS_ID__/__SERVICE___usage.go", want: false},
		{name: "dockerfile is not meta", relPath: "deployments/__PS_ID__/Dockerfile", want: false},
		{name: "compose yml is not meta", relPath: "deployments/__PS_ID__/compose.yml", want: false},
		{name: "cmd main go is meta (non-usage internal)", relPath: "cmd/__PS_ID__/main.go", want: true},
		{name: "internal service go is meta (non-usage internal)", relPath: "internal/apps/__PS_ID__/__SERVICE__.go", want: true},
		{name: "internal client go is meta (non-usage)", relPath: "internal/apps/__PS_ID__/client/client.go", want: true},
		{name: "internal test go is meta (non-usage)", relPath: "internal/apps/__PS_ID__/__SERVICE___cli_test.go", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isStructuralMetaFile(tt.relPath)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStripBuildIgnoreTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips tag and following blank line",
			input: "//go:build ignore\n\npackage main\n",
			want:  "package main\n",
		},
		{
			name:  "strips tag with no following blank line",
			input: "//go:build ignore\npackage main\n",
			want:  "package main\n",
		},
		{
			name:  "no tag present returns unchanged",
			input: "package main\n// some comment\n",
			want:  "package main\n// some comment\n",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := stripBuildIgnoreTag(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestHasUnresolvedPlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "resolved", input: "package sm_kms", want: false},
		{name: "unresolved single", input: "package __SERVICE__", want: true},
		{name: "unresolved multiple", input: "__PRODUCT_NAME_CONST__ = \"__SERVICE__\"", want: true},
		{name: "lowercase not placeholder", input: "var __lower__ = 1", want: false},
		{name: "empty string", input: "", want: false},
		{
			name:  "base64 sentinel is not unresolved",
			input: "prefix-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := hasUnresolvedPlaceholders(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAddGoSourceParams(t *testing.T) {
	t.Parallel()

	params := buildParams(cryptoutilSharedMagic.OTLPServiceSMKMS)
	pss := cryptoutilRegistry.AllProductServices()

	var kmsPS cryptoutilRegistry.ProductService

	for _, ps := range pss {
		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			kmsPS = ps

			break
		}
	}

	addGoSourceParams(params, kmsPS)

	require.NotEmpty(t, params[cryptoutilSharedMagic.CICDTemplateExpansionKeyService])
	require.NotEmpty(t, params["__USAGE_PREFIX__"])
	require.NotEmpty(t, params["__PRODUCT_NAME_CONST__"])
	require.NotEmpty(t, params["__SERVICE_NAME_CONST__"])
	require.NotEmpty(t, params["__SERVICE_ID_CONST__"])
	require.NotEmpty(t, params["__SERVICE_PORT_CONST__"])
	require.NotEmpty(t, params["__SERVICE_DISPLAY_NAME_CONST__"])
}

func TestBuildExpectedFS_GoSourceTemplateExpansion(t *testing.T) {
	t.Parallel()

	// Simulate a usage.go template with all resolvable placeholders.
	// Note: LoadTemplatesDir strips the //go:build ignore tag on load, so templates
	// passed directly to BuildExpectedFS should already have the tag removed.
	tmplContent := "package __PS_ID_UNDERSCORE__\n// __SERVICE_DISPLAY_NAME_CONST__\n"

	templates := map[string]string{
		"internal/apps/" + cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID + "/__SERVICE___usage.go": tmplContent,
	}

	expected := BuildExpectedFS(templates)

	// Should expand once per PS-ID (10 entries).
	require.Len(t, expected, len(cryptoutilRegistry.AllProductServices()))

	// Spot-check sm-kms expansion.
	content, ok := expected["internal/apps/sm-kms/kms_usage.go"]
	require.True(t, ok, "sm-kms usage file must be in expected FS")
	require.NotContains(t, content, "__PS_ID_UNDERSCORE__", "placeholder must be resolved")
	require.Contains(t, content, "sm_kms", "underscore form of psid must appear")
}

func TestBuildExpectedFS_GoSourceTemplateSkipsUnresolved(t *testing.T) {
	t.Parallel()

	// A template with a placeholder not in any buildParams â€” should produce 0 expanded entries.
	templates := map[string]string{
		"internal/apps/" + cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID + "/__SERVICE__.go": "//go:build ignore\n\nfunc __ENTRY_FUNC__() {}",
	}

	expected := BuildExpectedFS(templates)

	// All 10 expansions should be skipped because __ENTRY_FUNC__ is unresolved.
	require.Empty(t, expected, "templates with unresolved placeholders must be skipped")
}
