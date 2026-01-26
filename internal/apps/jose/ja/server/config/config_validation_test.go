// Copyright (c) 2025 Justin Cranford

package config

import (
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
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

// TestParseWithFlagSet_DefaultValues tests ParseWithFlagSet returns correct default values.
func TestParseWithFlagSet_DefaultValues(t *testing.T) {
	t.Parallel()

	// Create fresh FlagSet for this test (enables parallel execution).
	fs := pflag.NewFlagSet("test-defaults", pflag.ContinueOnError)

	// Parse with subcommand + no additional flags.
	args := []string{"start"}
	settings, err := ParseWithFlagSet(fs, args, false)
	if err != nil {
		t.Fatalf("ParseWithFlagSet() error = %v, want nil", err)
	}

	// Verify jose-ja defaults.
	if settings.DefaultMaxMaterials != 10 {
		t.Errorf("DefaultMaxMaterials = %d, want 10", settings.DefaultMaxMaterials)
	}
	if settings.AuditEnabled != true {
		t.Errorf("AuditEnabled = %v, want true", settings.AuditEnabled)
	}
	if settings.AuditSamplingRate != 100 {
		t.Errorf("AuditSamplingRate = %d, want 100", settings.AuditSamplingRate)
	}

	// Verify template defaults inherited.
	if settings.BindPublicPort != 9443 { // cryptoutilSharedMagic.JoseJAServicePort.
		t.Errorf("BindPublicPort = %d, want 9443", settings.BindPublicPort)
	}
	if settings.OTLPService != "jose-ja" {
		t.Errorf("OTLPService = %q, want %q", settings.OTLPService, "jose-ja")
	}
}

// TestParseWithFlagSet_OverrideDefaults tests ParseWithFlagSet with command line overrides.
func TestParseWithFlagSet_OverrideDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantMax  int
		wantAud  bool
		wantRate int
	}{
		{
			name:     "override max materials",
			args:     []string{"start", "--max-materials", "50"},
			wantMax:  50,
			wantAud:  true,
			wantRate: 100,
		},
		{
			name:     "disable audit",
			args:     []string{"start", "--audit-enabled=false"},
			wantMax:  10,
			wantAud:  false,
			wantRate: 100,
		},
		{
			name:     "override sampling rate",
			args:     []string{"start", "--audit-sampling-rate", "25"},
			wantMax:  10,
			wantAud:  true,
			wantRate: 25,
		},
		{
			name:     "override all jose-ja flags",
			args:     []string{"start", "--max-materials", "100", "--audit-enabled=false", "--audit-sampling-rate", "50"},
			wantMax:  100,
			wantAud:  false,
			wantRate: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create fresh FlagSet for each test.
			fs := pflag.NewFlagSet("test-override-"+tt.name, pflag.ContinueOnError)

			settings, err := ParseWithFlagSet(fs, tt.args, false)
			if err != nil {
				t.Fatalf("ParseWithFlagSet() error = %v, want nil", err)
			}

			if settings.DefaultMaxMaterials != tt.wantMax {
				t.Errorf("DefaultMaxMaterials = %d, want %d", settings.DefaultMaxMaterials, tt.wantMax)
			}
			if settings.AuditEnabled != tt.wantAud {
				t.Errorf("AuditEnabled = %v, want %v", settings.AuditEnabled, tt.wantAud)
			}
			if settings.AuditSamplingRate != tt.wantRate {
				t.Errorf("AuditSamplingRate = %d, want %d", settings.AuditSamplingRate, tt.wantRate)
			}
		})
	}
}

// TestParse_DefaultValues tests Parse returns correct default values using global pflag.
func TestParse_DefaultValues(t *testing.T) {
	// Use ParseWithFlagSet instead - Parse modifies global pflag.CommandLine.
	// Tests should use ParseWithFlagSet for isolation.
	t.Skip("Use TestParseWithFlagSet_DefaultValues instead - this test would modify global pflag state")
}

// TestParse_OverrideDefaults tests Parse with command line overrides using global pflag.
func TestParse_OverrideDefaults(t *testing.T) {
	// Use ParseWithFlagSet instead - Parse modifies global pflag.CommandLine.
	t.Skip("Use TestParseWithFlagSet_OverrideDefaults instead - this test would modify global pflag state")
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
