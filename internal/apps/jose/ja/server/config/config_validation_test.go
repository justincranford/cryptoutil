// Copyright (c) 2025 Justin Cranford

package config

import (
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestValidateJoseJASettings_NegativeMaxMaterials tests negative max materials validation.
func TestValidateJoseJASettings_NegativeMaxMaterials(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()
	settings.DefaultMaxMaterials = -1

	err := validateJoseJASettings(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials")
}

// TestValidateJoseJASettings_ZeroMaxMaterials tests zero max materials validation.
func TestValidateJoseJASettings_ZeroMaxMaterials(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()
	settings.DefaultMaxMaterials = 0

	err := validateJoseJASettings(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials")
}

// TestValidateJoseJASettings_NegativeSamplingRate tests negative audit sampling rate validation.
func TestValidateJoseJASettings_NegativeSamplingRate(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()
	settings.AuditSamplingRate = -1

	err := validateJoseJASettings(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "audit-sampling-rate")
}

// TestValidateJoseJASettings_SamplingRateAbove100 tests sampling rate above 100.
func TestValidateJoseJASettings_SamplingRateAbove100(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()
	settings.AuditSamplingRate = 101

	err := validateJoseJASettings(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "audit-sampling-rate")
}

// TestValidateJoseJASettings_CombinedErrors tests multiple validation errors combined.
func TestValidateJoseJASettings_CombinedErrors(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()
	settings.DefaultMaxMaterials = -1
	settings.AuditSamplingRate = 150

	err := validateJoseJASettings(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials")
	require.Contains(t, err.Error(), "audit-sampling-rate")
}

// TestLogOutputFormat tests log output format smoke test.
func TestLogOutputFormat(t *testing.T) {
	t.Parallel()

	// Note: logJoseJASettings writes to os.Stderr.
	// Testing exact output would require capturing stderr.
	// This is a smoke test to ensure no panic.

	settings := DefaultTestConfig()

	// Should not panic.
	logJoseJASettings(settings)
}

// TestParse_DefaultValues tests Parse returns correct default values.
func TestParse_DefaultValues(t *testing.T) {
	// Cannot run in parallel due to global flag state.

	// Note: This test would require resetting pflag state.
	// Skip for now as Parse modifies global state.
	t.Skip("TODO P2.4: Add Parse tests with flag state isolation")
}

// TestParse_OverrideDefaults tests Parse with command line overrides.
func TestParse_OverrideDefaults(t *testing.T) {
	// Cannot run in parallel due to global flag state.

	t.Skip("TODO P2.4: Add Parse tests with flag state isolation")
}

// TestPublicBaseURL tests URL construction for public server.
func TestPublicBaseURL(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()

	url := settings.PublicBaseURL()

	require.NotEmpty(t, url)
	require.True(t, strings.HasPrefix(url, "https://"))
}

// TestPrivateBaseURL tests URL construction for private server.
func TestPrivateBaseURL(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()
	settings.BindPrivatePort = 9090 // Set explicit port for test.

	url := settings.PrivateBaseURL()

	require.NotEmpty(t, url)
	require.True(t, strings.HasPrefix(url, "https://"))
	require.Contains(t, url, ":9090")
}

// TestVerifySettingsRegistration verifies jose-ja settings are registered.
func TestVerifySettingsRegistration(t *testing.T) {
	t.Parallel()

	// Verify jose-ja settings are registered.
	require.NotNil(t, maxMaterialsSetting)
	require.NotNil(t, auditEnabledSetting)
	require.NotNil(t, auditSamplingRateSetting)

	require.Equal(t, "max-materials", maxMaterialsSetting.Name)
	require.Equal(t, "audit-enabled", auditEnabledSetting.Name)
	require.Equal(t, "audit-sampling-rate", auditSamplingRateSetting.Name)
}

// TestVerifyDefaultConstants verifies default value constants.
func TestVerifyDefaultConstants(t *testing.T) {
	t.Parallel()

	// Verify defaults match magic constants.
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, defaultMaxMaterials)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultEnabled, defaultAuditEnabled)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate, defaultAuditSamplingRate)
}
