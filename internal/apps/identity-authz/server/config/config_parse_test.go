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
		name string
		args []string
		want string
		port uint16
	}{
		{name: "defaults", want: defaultIssuer, port: cryptoutilSharedMagic.IdentityAuthzServicePort},
		{name: "overrides", args: []string{"--issuer=https://issuer.example", "--bind-public-port=12345"}, want: "https://issuer.example", port: 12345},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			args := append([]string{testConfigStartCommand}, tc.args...)
			settings, err := ParseWithFlagSet(pflag.NewFlagSet("test", pflag.ContinueOnError), args, false)
			require.NoError(t, err)
			require.Equal(t, tc.want, settings.Issuer)
			require.Equal(t, tc.port, settings.BindPublicPort)
			require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityAuthz, settings.OTLPService)
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
	require.Equal(t, defaultIssuer, settings.Issuer)
}
