// Copyright (c) 2025 Justin Cranford

// Package config tests Parse via ParseWithFlagSet for pki-ca server configuration.
package config

import (
"testing"

"github.com/spf13/pflag"
"github.com/stretchr/testify/require"
)

// TestParse_HappyPath tests ParseWithFlagSet with valid arguments.
// Uses a fresh FlagSet to avoid pflag.CommandLine state conflicts with count > 1 runs.
func TestParse_HappyPath(t *testing.T) {
// Not parallel: pflag FlagSets may share global viper state.
fs := pflag.NewFlagSet("test-pki-happy-path", pflag.ContinueOnError)
args := []string{
"start", // Required subcommand.
"--bind-public-address", "127.0.0.1",
"--dev",
}

settings, err := ParseWithFlagSet(fs, args, false)
require.NoError(t, err)
require.NotNil(t, settings)
require.Equal(t, "127.0.0.1", settings.BindPublicAddress)
// Verify CA-specific defaults.
require.Empty(t, settings.CAConfigPath)
require.Empty(t, settings.ProfilesPath)
require.True(t, settings.EnableEST)        // Default is true.
require.True(t, settings.EnableOCSP)       // Default is true.
require.True(t, settings.EnableCRL)        // Default is true.
require.False(t, settings.EnableTimestamp) // Default is false.
}
