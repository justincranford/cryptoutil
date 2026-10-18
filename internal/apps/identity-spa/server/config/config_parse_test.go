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
		name     string
		args     []string
		wantRP   string
		wantAge  int
		wantPort uint16
	}{
		{name: "defaults", wantRP: defaultRPOrigin, wantAge: defaultCacheMaxAge, wantPort: cryptoutilSharedMagic.IdentitySPAServicePort},
		{name: "overrides", args: []string{"--rp-origin=https://rp.example", "--cache-control-max-age=10", "--bind-public-port=12345"}, wantRP: "https://rp.example", wantAge: 10, wantPort: 12345},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			args := append([]string{testConfigStartCommand}, tc.args...)
			settings, err := ParseWithFlagSet(pflag.NewFlagSet("test", pflag.ContinueOnError), args, false)
			require.NoError(t, err)
			require.Equal(t, tc.wantRP, settings.RPOrigin)
			require.Equal(t, tc.wantAge, settings.CacheControlMaxAge)
			require.Equal(t, tc.wantPort, settings.BindPublicPort)
			require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentitySPA, settings.OTLPService)
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
	require.Equal(t, defaultStaticFilesPath, settings.StaticFilesPath)
}
