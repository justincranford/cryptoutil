// Copyright (c) 2025 Justin Cranford
//

package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()

	settings := config.DefaultTestConfig()

	require.NotNil(t, settings)
	require.NotNil(t, settings.ServiceTemplateServerSettings)

	// Verify cipher-im specific defaults.
	require.Equal(t, cryptoutilSharedMagic.CipherJWEAlgorithm, settings.MessageJWEAlgorithm)
	require.Equal(t, cryptoutilSharedMagic.CipherMessageMinLength, settings.MessageMinLength)
	require.Equal(t, cryptoutilSharedMagic.CipherMessageMaxLength, settings.MessageMaxLength)
	require.Equal(t, cryptoutilSharedMagic.CipherRecipientsMinCount, settings.RecipientsMinCount)
	require.Equal(t, cryptoutilSharedMagic.CipherRecipientsMaxCount, settings.RecipientsMaxCount)

	// Note: BrowserRealms and ServiceRealms are populated by Parse() from config file or flags.
	// DefaultTestConfig provides minimal valid settings for unit tests, not full runtime config.

	// Verify development mode defaults.
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, settings.BindPublicAddress)
	require.Equal(t, uint16(0), settings.BindPublicPort) // Dynamic port in dev mode.
	require.True(t, settings.DevMode)
}

func TestNewTestConfig_CustomValues(t *testing.T) {
	t.Parallel()

	bindAddr := "192.168.1.100"
	bindPort := uint16(8888)
	devMode := false

	settings := config.NewTestConfig(bindAddr, bindPort, devMode)

	require.NotNil(t, settings)
	require.Equal(t, bindAddr, settings.BindPublicAddress)
	require.Equal(t, bindPort, settings.BindPublicPort)
	require.Equal(t, devMode, settings.DevMode)

	// Cipher-im defaults still apply.
	require.Equal(t, cryptoutilSharedMagic.CipherJWEAlgorithm, settings.MessageJWEAlgorithm)
	require.Equal(t, cryptoutilSharedMagic.CipherMessageMinLength, settings.MessageMinLength)
	require.Equal(t, cryptoutilSharedMagic.CipherMessageMaxLength, settings.MessageMaxLength)
}

func TestNewTestConfig_OTLPServiceOverride(t *testing.T) {
	t.Parallel()

	settings := config.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	require.NotNil(t, settings)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceCipherIM, settings.OTLPService)
}

func TestNewTestConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	// NewTestConfig validates addresses aren't empty (security requirement).
	// Use minimal valid addresses instead of empty strings.
	settings := config.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, false)

	require.NotNil(t, settings)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, settings.BindPublicAddress)
	require.Equal(t, uint16(0), settings.BindPublicPort)
	require.False(t, settings.DevMode)

	// Cipher-im defaults still populated.
	require.NotEmpty(t, settings.MessageJWEAlgorithm)
	require.Greater(t, settings.MessageMinLength, 0)
	require.Greater(t, settings.MessageMaxLength, 0)
	require.Greater(t, settings.RecipientsMinCount, 0)
	require.Greater(t, settings.RecipientsMaxCount, 0)
}

func TestDefaultTestConfig_PortAllocation(t *testing.T) {
	t.Parallel()

	// Create two instances to verify they don't conflict.
	settings1 := config.DefaultTestConfig()
	settings2 := config.DefaultTestConfig()

	require.NotNil(t, settings1)
	require.NotNil(t, settings2)

	// Both should use dynamic ports (0).
	require.Equal(t, uint16(0), settings1.BindPublicPort)
	require.Equal(t, uint16(0), settings2.BindPublicPort)
}

func TestParse_HappyPath(t *testing.T) {
	// Don't use t.Parallel() - Parse modifies global state (pflag).
	
	// Parse uses template defaults for cipher settings since flags can't be tested directly
	// due to pflag.Parse() being called twice (once in template, once in cipher).
	args := []string{
		"start", // Required subcommand.
		"--bind-public-address", "127.0.0.1",
		"--bind-public-port", "8080",
	}
	
	settings, err := config.Parse(args, false)
	
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "127.0.0.1", settings.BindPublicAddress)
	require.Equal(t, uint16(cryptoutilSharedMagic.CipherServicePort), settings.BindPublicPort) // Overridden to cipher default.
	require.Equal(t, cryptoutilSharedMagic.CipherJWEAlgorithm, settings.MessageJWEAlgorithm)
	require.Equal(t, cryptoutilSharedMagic.CipherMessageMinLength, settings.MessageMinLength)
	require.Equal(t, cryptoutilSharedMagic.CipherMessageMaxLength, settings.MessageMaxLength)
	require.Equal(t, cryptoutilSharedMagic.CipherRecipientsMinCount, settings.RecipientsMinCount)
	require.Equal(t, cryptoutilSharedMagic.CipherRecipientsMaxCount, settings.RecipientsMaxCount)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceCipherIM, settings.OTLPService)
}

// Note: Testing validation through Parse() with invalid defaults in magic constants
// is not feasible since defaults are always valid. Direct validation function testing
// would require exporting validateCipherImSettings() or using reflection.
// The validation logic is indirectly tested via Parse() above with valid defaults.

func TestNewTestConfig_InheritedTemplateSettings(t *testing.T) {
	t.Parallel()

	settings := config.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 8888, false)

	require.NotNil(t, settings)
	require.NotNil(t, settings.ServiceTemplateServerSettings)

	// Note: BrowserRealms and ServiceRealms are populated by Parse() from config file or flags.
	// NewTestConfig provides minimal valid settings, not full runtime configuration.

	// Verify cipher-im port overrides.
	require.Equal(t, uint16(8888), settings.BindPublicPort)

	// BindPrivatePort uses dynamic allocation (0) in tests to avoid port conflicts.
	require.Equal(t, uint16(0), settings.BindPrivatePort)
}

func TestNewTestConfig_MessageConstraints(t *testing.T) {
	t.Parallel()

	settings := config.DefaultTestConfig()

	require.NotNil(t, settings)

	// Verify message constraints are valid.
	require.Greater(t, settings.MessageMinLength, 0)
	require.Greater(t, settings.MessageMaxLength, settings.MessageMinLength)

	// Verify recipient constraints are valid.
	require.Greater(t, settings.RecipientsMinCount, 0)
	require.Greater(t, settings.RecipientsMaxCount, settings.RecipientsMinCount)
}

func TestNewTestConfig_MessageJWEAlgorithm(t *testing.T) {
	t.Parallel()

	settings := config.DefaultTestConfig()

	require.NotNil(t, settings)
	require.NotEmpty(t, settings.MessageJWEAlgorithm)
	require.Equal(t, cryptoutilSharedMagic.CipherJWEAlgorithm, settings.MessageJWEAlgorithm)
}
