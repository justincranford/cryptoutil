// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidateJoseSettings_ValidSettings tests validation passes for valid settings.
func TestValidateJoseSettings_ValidSettings(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 100,
		AuditEnabled:              true,
		AuditSamplingRate:         0.5,
	}

	err := validateJoseSettings(settings)
	require.NoError(t, err)
}

// TestValidateJoseSettings_MinMaterials tests validation passes for minimum materials value.
func TestValidateJoseSettings_MinMaterials(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 1,
		AuditEnabled:              false,
		AuditSamplingRate:         0.0,
	}

	err := validateJoseSettings(settings)
	require.NoError(t, err)
}

// TestValidateJoseSettings_MaxSamplingRate tests validation passes for maximum sampling rate.
func TestValidateJoseSettings_MaxSamplingRate(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 10,
		AuditEnabled:              true,
		AuditSamplingRate:         1.0,
	}

	err := validateJoseSettings(settings)
	require.NoError(t, err)
}

// TestValidateJoseSettings_ZeroMaterials tests validation fails for zero materials.
func TestValidateJoseSettings_ZeroMaterials(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 0,
		AuditEnabled:              true,
		AuditSamplingRate:         0.5,
	}

	err := validateJoseSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials-per-elastic-key must be >= 1")
}

// TestValidateJoseSettings_NegativeMaterials tests validation fails for negative materials.
func TestValidateJoseSettings_NegativeMaterials(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: -1,
		AuditEnabled:              true,
		AuditSamplingRate:         0.5,
	}

	err := validateJoseSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials-per-elastic-key must be >= 1")
}

// TestValidateJoseSettings_NegativeSamplingRate tests validation fails for negative sampling rate.
func TestValidateJoseSettings_NegativeSamplingRate(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 10,
		AuditEnabled:              true,
		AuditSamplingRate:         -0.1,
	}

	err := validateJoseSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "audit-sampling-rate must be between 0.0 and 1.0")
}

// TestValidateJoseSettings_SamplingRateOverOne tests validation fails for sampling rate over 1.
func TestValidateJoseSettings_SamplingRateOverOne(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 10,
		AuditEnabled:              true,
		AuditSamplingRate:         1.5,
	}

	err := validateJoseSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "audit-sampling-rate must be between 0.0 and 1.0")
}

// TestValidateJoseSettings_MultipleErrors tests validation collects multiple errors.
func TestValidateJoseSettings_MultipleErrors(t *testing.T) {
	settings := &JoseServerSettings{
		MaxMaterialsPerElasticKey: 0,
		AuditEnabled:              true,
		AuditSamplingRate:         2.0,
	}

	err := validateJoseSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials-per-elastic-key must be >= 1")
	require.Contains(t, err.Error(), "audit-sampling-rate must be between 0.0 and 1.0")
}

// TestNewDevSettings_Defaults tests NewDevSettings creates settings with defaults.
func TestNewDevSettings_Defaults(t *testing.T) {
	t.Parallel()
	settings := NewDevSettings()

	require.NotNil(t, settings)
	require.NotNil(t, settings.ServiceTemplateServerSettings)
	require.Equal(t, defaultMaxMaterialsPerElasticKey, settings.MaxMaterialsPerElasticKey)
	require.Equal(t, defaultAuditEnabled, settings.AuditEnabled)
	require.Equal(t, defaultAuditSamplingRate, settings.AuditSamplingRate)

	// Also validate the settings are valid.
	err := validateJoseSettings(settings)
	require.NoError(t, err)
}

// TestNewTestSettings_Defaults tests NewTestSettings creates settings with defaults.
func TestNewTestSettings_Defaults(t *testing.T) {
	t.Parallel()
	settings := NewTestSettings()

	require.NotNil(t, settings)
	require.NotNil(t, settings.ServiceTemplateServerSettings)
	require.Equal(t, defaultMaxMaterialsPerElasticKey, settings.MaxMaterialsPerElasticKey)
	require.Equal(t, defaultAuditEnabled, settings.AuditEnabled)
	require.Equal(t, defaultAuditSamplingRate, settings.AuditSamplingRate)

	// Port should be 0 for dynamic allocation in tests.
	require.Equal(t, uint16(0), settings.BindPublicPort)

	// Also validate the settings are valid.
	err := validateJoseSettings(settings)
	require.NoError(t, err)
}

// TestLogJoseSettings_Logs tests logJoseSettings does not panic.
func TestLogJoseSettings_Logs(t *testing.T) {
	t.Parallel()
	settings := NewTestSettings()

	// logJoseSettings should not panic.
	require.NotPanics(t, func() {
		logJoseSettings(settings)
	})
}
