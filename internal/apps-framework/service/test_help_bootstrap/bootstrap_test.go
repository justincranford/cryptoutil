// Copyright (c) 2025-2026 Justin Cranford.

package test_help_bootstrap

import (
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"

	"github.com/stretchr/testify/require"
)

func TestCloneStringSlice_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
	}{
		{name: "nil slice", input: nil},
		{name: "empty slice", input: []string{}},
		{name: "populated slice", input: []string{"a", "b"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := cloneStringSlice(tc.input)
			if len(tc.input) == 0 {
				require.Nil(t, output)

				return
			}

			require.Equal(t, tc.input, output)
			require.NotSame(t, &tc.input[0], &output[0])
		})
	}
}

func TestNewTestServerSettings_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "returns parallel-safe auto-tls dynamic-port settings"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := NewTestServerSettings(t)
			require.NotNil(t, settings)
			require.Equal(t, uint16(0), settings.BindPublicPort)
			require.Equal(t, uint16(0), settings.BindPrivatePort)
			require.Equal(t, cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeAuto, settings.TLSPublicProvisionMode)
			require.Equal(t, cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeAuto, settings.TLSPrivateProvisionMode)

			origFirst := settings.TLSPublicDNSNames[0]
			settings.TLSPublicDNSNames[0] = "mutated.local"

			settings2 := NewTestServerSettings(t)
			require.NotNil(t, settings2)
			require.Equal(t, origFirst, settings2.TLSPublicDNSNames[0])
		})
	}
}
