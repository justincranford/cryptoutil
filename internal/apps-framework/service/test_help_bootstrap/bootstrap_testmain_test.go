// Copyright (c) 2025-2026 Justin Cranford.

package test_help_bootstrap

import (
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewTestServerSettingsForTestMain_Isolated(t *testing.T) {
	t.Parallel()

	settings := NewTestServerSettingsForTestMain()
	otherSettings := NewTestServerSettingsForTestMain()

	require.Equal(t, uint16(0), settings.BindPublicPort)
	require.Equal(t, uint16(0), settings.BindPrivatePort)
	require.Equal(t, cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeAuto, settings.TLSPublicProvisionMode)
	require.Equal(t, cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeAuto, settings.TLSPrivateProvisionMode)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, settings.BindPublicAddress)

	tests := []struct {
		name   string
		values *[]string
		other  []string
	}{
		{name: "public dns names", values: &settings.TLSPublicDNSNames, other: otherSettings.TLSPublicDNSNames},
		{name: "public ip addresses", values: &settings.TLSPublicIPAddresses, other: otherSettings.TLSPublicIPAddresses},
		{name: "private dns names", values: &settings.TLSPrivateDNSNames, other: otherSettings.TLSPrivateDNSNames},
		{name: "private ip addresses", values: &settings.TLSPrivateIPAddresses, other: otherSettings.TLSPrivateIPAddresses},
		{name: "cors origins", values: &settings.CORSAllowedOrigins, other: otherSettings.CORSAllowedOrigins},
		{name: "cors methods", values: &settings.CORSAllowedMethods, other: otherSettings.CORSAllowedMethods},
		{name: "cors headers", values: &settings.CORSAllowedHeaders, other: otherSettings.CORSAllowedHeaders},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, *tc.values)
			require.Equal(t, *tc.values, tc.other)
			(*tc.values)[0] = "mutated"
			require.NotEqual(t, "mutated", tc.other[0])
		})
	}
}
