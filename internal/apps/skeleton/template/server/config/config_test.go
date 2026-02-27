// Copyright (c) 2025 Justin Cranford
//

package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestParseWithFlagSet_Defaults verifies default skeleton-template configuration.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestParseWithFlagSet_Defaults(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-defaults", pflag.ContinueOnError)

	cfg, err := ParseWithFlagSet(fs, []string{"start", "--profile=test"}, false)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, cfg.OTLPService)
	require.Equal(t, uint16(cryptoutilSharedMagic.SkeletonTemplateServicePort), cfg.BindPublicPort)
}

// TestParseWithFlagSet_CustomPort verifies custom port override.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestParseWithFlagSet_CustomPort(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-custom-port", pflag.ContinueOnError)

	cfg, err := ParseWithFlagSet(fs, []string{"start", "--profile=test", "--bind-public-port=0"}, false)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, uint16(0), cfg.BindPublicPort)
}

// TestParseWithFlagSet_InvalidFlag verifies error on unknown flags.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestParseWithFlagSet_InvalidFlag(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-invalid", pflag.ContinueOnError)

	_, err := ParseWithFlagSet(fs, []string{"start", "--nonexistent-flag=true"}, false)
	require.Error(t, err)
}

func TestNewTestConfig(t *testing.T) {
	t.Parallel()


	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, cfg.OTLPService)
	require.Equal(t, uint16(0), cfg.BindPublicPort)
}

func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, cfg.OTLPService)
}

func TestValidateSettings_NilBase(t *testing.T) {
	t.Parallel()


	err := validateSettings(&SkeletonTemplateServerSettings{
		ServiceTemplateServerSettings: nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "base template settings cannot be nil")
}

// TestParse_Delegates verifies Parse delegates to ParseWithFlagSet with global CommandLine.
// Sequential: uses viper global state via Parse.
func TestParse_Delegates(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	// Parse with valid args to verify delegation works (pflag.CommandLine
	// uses ExitOnError mode, so we cannot test invalid flags here).
	cfg, err := Parse([]string{"start", "--profile=test"}, false)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, cfg.OTLPService)
}
