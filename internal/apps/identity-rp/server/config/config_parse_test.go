// Copyright (c) 2025-2026 Justin Cranford.
package config

import (
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

const (
	testConfigStartCommand          = "start"
	testAuthzServerURLRequiredError = "authz-server-url is required in production mode"
	testSPAOriginInvalid            = "invalid-origin"
	testSPAOriginHTTPRequiredError  = "spa-origin must start with http:// or https://"
)

func TestParseWithFlagSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantID   string
		wantPort uint16
	}{
		{name: "defaults", wantID: defaultClientID, wantPort: cryptoutilSharedMagic.IdentityRPServicePort},
		{name: "overrides", args: []string{"--client-id=rp-client", "--client-secret=sample-secret-value", "--bind-public-port=12345"}, wantID: "rp-client", wantPort: 12345},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			args := append([]string{testConfigStartCommand}, tc.args...)
			settings, err := ParseWithFlagSet(pflag.NewFlagSet("test", pflag.ContinueOnError), args, false)
			require.NoError(t, err)
			require.Equal(t, tc.wantID, settings.ClientID)
			require.Equal(t, tc.wantPort, settings.BindPublicPort)
			require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityRP, settings.OTLPService)
		})
	}
}

func TestParseWithFlagSet_UnknownFlag(t *testing.T) {
	t.Parallel()

	_, err := ParseWithFlagSet(
		pflag.NewFlagSet("test", pflag.ContinueOnError),
		[]string{testConfigStartCommand, "--unknown"},
		false,
	)
	require.ErrorContains(t, err, "failed to parse template settings")
}

// Sequential: temporarily replaces pflag.CommandLine global state.
func TestParse(t *testing.T) {
	originalFlagSet := pflag.CommandLine

	t.Cleanup(func() {
		pflag.CommandLine = originalFlagSet
	})

	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)

	settings, err := Parse([]string{testConfigStartCommand}, false)
	require.NoError(t, err)
	require.Equal(t, defaultAuthzServerURL, settings.AuthzServerURL)
}

func TestValidateIdentityRPSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		settings   *IdentityRPServerSettings
		wantErrors []string
	}{
		{
			name: "development allows empty authorization URL",
			settings: &IdentityRPServerSettings{
				ServiceFrameworkServerSettings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{DevMode: true},
			},
		},
		{
			name: "production requires authorization URL",
			settings: &IdentityRPServerSettings{
				ServiceFrameworkServerSettings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
			},
			wantErrors: []string{testAuthzServerURLRequiredError},
		},
		{
			name: "SPA origin must be HTTP",
			settings: &IdentityRPServerSettings{
				ServiceFrameworkServerSettings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{DevMode: true},
				SPAOrigin:                      testSPAOriginInvalid,
			},
			wantErrors: []string{testSPAOriginHTTPRequiredError},
		},
		{
			name: "aggregates configuration errors",
			settings: &IdentityRPServerSettings{
				ServiceFrameworkServerSettings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
				SPAOrigin:                      testSPAOriginInvalid,
			},
			wantErrors: []string{
				testAuthzServerURLRequiredError,
				testSPAOriginHTTPRequiredError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateIdentityRPSettings(tc.settings)
			if len(tc.wantErrors) == 0 {
				require.NoError(t, err)

				return
			}

			for _, want := range tc.wantErrors {
				require.ErrorContains(t, err, want)
			}
		})
	}
}
