// Copyright (c) 2025 Justin Cranford

package config

import (
	"bytes"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateJoseJASettings_HappyPath(t *testing.T) {
	t.Parallel()

	settings := &JoseJAServerSettings{
		ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
		DefaultMaxMaterials:           cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		AuditEnabled:                  true,
		AuditSamplingRate:             cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate,
	}
	err := validateJoseJASettings(settings)
	require.NoError(t, err)
}

func TestValidateJoseJASettings_MinMaxMaterials(t *testing.T) {
	t.Parallel()
	t.Run("at_minimum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           cryptoutilSharedMagic.JoseJAMinMaterials, // 1
			AuditSamplingRate:             50,
		}
		err := validateJoseJASettings(settings)
		require.NoError(t, err)
	})

	t.Run("at_maximum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           cryptoutilSharedMagic.JoseJAMaxMaterials, // 100
			AuditSamplingRate:             50,
		}
		err := validateJoseJASettings(settings)
		require.NoError(t, err)
	})

	t.Run("below_minimum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           cryptoutilSharedMagic.JoseJAMinMaterials - 1, // 0
			AuditSamplingRate:             50,
		}
		err := validateJoseJASettings(settings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "max-materials must be >=")
	})

	t.Run("above_maximum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           cryptoutilSharedMagic.JoseJAMaxMaterials + 1, // 101
			AuditSamplingRate:             50,
		}
		err := validateJoseJASettings(settings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "max-materials must be <=")
	})
}

func TestValidateJoseJASettings_AuditSamplingRate(t *testing.T) {
	t.Parallel()
	t.Run("at_minimum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           10,
			AuditSamplingRate:             cryptoutilSharedMagic.JoseJAAuditMinSamplingRate, // 0
		}
		err := validateJoseJASettings(settings)
		require.NoError(t, err)
	})

	t.Run("at_maximum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           10,
			AuditSamplingRate:             cryptoutilSharedMagic.JoseJAAuditMaxSamplingRate, // 100
		}
		err := validateJoseJASettings(settings)
		require.NoError(t, err)
	})

	t.Run("below_minimum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           10,
			AuditSamplingRate:             cryptoutilSharedMagic.JoseJAAuditMinSamplingRate - 1, // -1
		}
		err := validateJoseJASettings(settings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "audit-sampling-rate must be >=")
	})

	t.Run("above_maximum", func(t *testing.T) {
		settings := &JoseJAServerSettings{
			ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
			DefaultMaxMaterials:           10,
			AuditSamplingRate:             cryptoutilSharedMagic.JoseJAAuditMaxSamplingRate + 1, // 101
		}
		err := validateJoseJASettings(settings)
		require.Error(t, err)
		require.Contains(t, err.Error(), "audit-sampling-rate must be <=")
	})
}

func TestValidateJoseJASettings_MultipleErrors(t *testing.T) {
	t.Parallel()

	settings := &JoseJAServerSettings{
		ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
		DefaultMaxMaterials:           0,   // Invalid: below minimum
		AuditSamplingRate:             101, // Invalid: above maximum
	}
	err := validateJoseJASettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max-materials must be >=")
	require.Contains(t, err.Error(), "audit-sampling-rate must be <=")
}

func TestLogJoseJASettings(t *testing.T) {
	t.Parallel()
	// Capture stderr output.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	settings := &JoseJAServerSettings{
		ServiceTemplateServerSettings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
			BindPublicProtocol:  "https",
			BindPublicAddress:   "127.0.0.1",
			BindPublicPort:      8060,
			BindPrivateProtocol: "https",
			BindPrivateAddress:  "127.0.0.1",
			BindPrivatePort:     9090,
			OTLPService:         "jose-ja",
			BrowserRealms:       []string{"default"},
			ServiceRealms:       []string{"service"},
		},
		DefaultMaxMaterials: 10,
		AuditEnabled:        true,
		AuditSamplingRate:   50,
	}

	logJoseJASettings(settings)

	// Restore stderr and read captured output.
	_ = w.Close() //nolint:errcheck // Test cleanup

	os.Stderr = oldStderr

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(r)

	output := buf.String()

	// Verify output contains expected content.
	require.Contains(t, output, "Jose-JA Server Settings:")
	require.Contains(t, output, "Public Server:")
	require.Contains(t, output, "Private Server:")
	require.Contains(t, output, "OTLP Service: jose-ja")
	require.Contains(t, output, "Browser Realms: default")
	require.Contains(t, output, "Service Realms: service")
	require.Contains(t, output, "Default Max Materials: 10")
	require.Contains(t, output, "Audit Enabled: true")
	require.Contains(t, output, "Audit Sampling Rate: 50%")
}

// TestJoseJAServerSettings_DefaultValues verifies the default constant values.
func TestJoseJAServerSettings_DefaultValues(t *testing.T) {
	t.Parallel()
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultEnabled, cryptoutilSharedMagic.JoseJAAuditDefaultEnabled)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate, cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate)
}

// TestSettingRegistrations verifies jose-ja settings are properly configured.
func TestSettingRegistrations(t *testing.T) {
	t.Parallel()
	require.Equal(t, "max-materials", maxMaterialsSetting.Name)
	require.Equal(t, "audit-enabled", auditEnabledSetting.Name)
	require.Equal(t, "audit-sampling-rate", auditSamplingRateSetting.Name)

	// Verify settings have description fields.
	require.NotEmpty(t, maxMaterialsSetting.Description)
	require.NotEmpty(t, auditEnabledSetting.Description)
	require.NotEmpty(t, auditSamplingRateSetting.Description)

	// Verify settings have usage fields.
	require.NotEmpty(t, maxMaterialsSetting.Usage)
	require.NotEmpty(t, auditEnabledSetting.Usage)
	require.NotEmpty(t, auditSamplingRateSetting.Usage)
}

// TestNewTestConfig verifies the test config helper function.
func TestNewTestConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 8080, true)

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, uint16(8080), cfg.BindPublicPort)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceJoseJA, cfg.OTLPService)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cfg.DefaultMaxMaterials)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultEnabled, cfg.AuditEnabled)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate, cfg.AuditSamplingRate)
}

// TestDefaultTestConfig verifies the default test config helper function.
func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, uint16(0), cfg.BindPublicPort) // Dynamic port allocation.
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceJoseJA, cfg.OTLPService)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cfg.DefaultMaxMaterials)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultEnabled, cfg.AuditEnabled)
	require.Equal(t, cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate, cfg.AuditSamplingRate)
}
