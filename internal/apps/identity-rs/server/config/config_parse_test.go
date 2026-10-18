// Copyright (c) 2025-2026 Justin Cranford.
package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

const testConfigStartCommand = "start"

func TestParseWithFlagSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		args              []string
		wantIntrospection string
		wantPort          uint16
	}{
		{name: "defaults", wantIntrospection: defaultIntrospectionURL, wantPort: cryptoutilSharedMagic.IdentityRSServicePort},
		{name: "overrides", args: []string{"--introspection-url=https://authz.example/introspect", "--bind-public-port=12345"}, wantIntrospection: "https://authz.example/introspect", wantPort: 12345},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			args := append([]string{testConfigStartCommand}, tc.args...)
			settings, err := ParseWithFlagSet(pflag.NewFlagSet("test", pflag.ContinueOnError), args, false)
			require.NoError(t, err)
			require.Equal(t, defaultRSAuthzServerURL, settings.AuthzServerURL)
			require.Equal(t, tc.wantIntrospection, settings.IntrospectionURL)
			require.Equal(t, tc.wantPort, settings.BindPublicPort)
			require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityRS, settings.OTLPService)
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
	require.Equal(t, defaultRSAuthzServerURL, settings.AuthzServerURL)
}

func TestMaskEmpty(t *testing.T) {
	t.Parallel()

	require.Equal(t, "(not configured)", maskEmpty("", "(not configured)"))
	require.Equal(t, "configured", maskEmpty("configured", "(not configured)"))
}
