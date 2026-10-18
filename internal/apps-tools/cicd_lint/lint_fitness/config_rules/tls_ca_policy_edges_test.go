// Copyright (c) 2025-2026 Justin Cranford.
package config_rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testTLSCAConfigKey = "key"
	testReadError      = "read error"
	testParseError     = "parse error"
)

func TestHasNonEmptyStringKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]any
		want   bool
	}{
		{name: "missing", config: map[string]any{}, want: false},
		{name: cryptoutilSharedMagic.EMPTY, config: map[string]any{testTLSCAConfigKey: ""}, want: false},
		{name: "whitespace", config: map[string]any{testTLSCAConfigKey: " \t "}, want: false},
		{name: "non-empty string", config: map[string]any{testTLSCAConfigKey: "value"}, want: true},
		{name: "non-string value counts as present", config: map[string]any{testTLSCAConfigKey: true}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, hasNonEmptyStringKey(tc.config, testTLSCAConfigKey))
		})
	}
}

func TestCheckTLSCAPolicyPairs_ReadAndParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*testing.T, string)
		wantReason string
	}{
		{
			name: testReadError,
			setup: func(_ *testing.T, _ string) {
			},
			wantReason: testReadError,
		},
		{
			name: testParseError,
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("[broken"), cryptoutilSharedMagic.CacheFilePermissions))
			},
			wantReason: testParseError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "overlay.yml")
			if tc.name != testReadError {
				tc.setup(t, path)
			}

			violations := checkTLSCAPolicyPairs(path)

			require.Len(t, violations, 1)
			require.Contains(t, violations[0], tc.wantReason)
		})
	}
}

func TestCheckTLSCAPolicyCouplingInDir_OverlayReadFailure(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	overlayDir := filepath.Join(rootDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS, "config")
	require.NoError(t, os.MkdirAll(overlayDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.Mkdir(
		filepath.Join(overlayDir, cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-framework-sqlite-1.yml"),
		cryptoutilSharedMagic.DirPermissions,
	))

	err := checkTLSCAPolicyCouplingInDir(newTestLogger(), rootDir)

	require.ErrorContains(t, err, "config-tls-ca-policy-coupling violations")
	require.ErrorContains(t, err, testReadError)
}
