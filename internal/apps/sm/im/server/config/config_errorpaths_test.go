// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestValidateSmIMSettings_Errors tests validation error paths.
func TestValidateSmIMSettings_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *SmIMServerSettings
		wantErr  string
	}{
		{
			name: "empty_jwe_algorithm",
			settings: &SmIMServerSettings{
				MessageJWEAlgorithm: "",
				MessageMinLength:    1,
				MessageMaxLength:    cryptoutilSharedMagic.JoseJAMaxMaterials,
				RecipientsMinCount:  1,
				RecipientsMaxCount:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			wantErr: "message-jwe-algorithm cannot be empty",
		},
		{
			name: "message_min_length_zero",
			settings: &SmIMServerSettings{
				MessageJWEAlgorithm: cryptoutilSharedMagic.JoseEncA256GCM,
				MessageMinLength:    0,
				MessageMaxLength:    cryptoutilSharedMagic.JoseJAMaxMaterials,
				RecipientsMinCount:  1,
				RecipientsMaxCount:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			wantErr: "message-min-length must be >= 1, got 0",
		},
		{
			name: "message_max_less_than_min",
			settings: &SmIMServerSettings{
				MessageJWEAlgorithm: cryptoutilSharedMagic.JoseEncA256GCM,
				MessageMinLength:    cryptoutilSharedMagic.JoseJAMaxMaterials,
				MessageMaxLength:    cryptoutilSharedMagic.IMMaxUsernameLength,
				RecipientsMinCount:  1,
				RecipientsMaxCount:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			wantErr: "message-max-length (50) must be >= message-min-length (100)",
		},
		{
			name: "recipients_min_count_zero",
			settings: &SmIMServerSettings{
				MessageJWEAlgorithm: cryptoutilSharedMagic.JoseEncA256GCM,
				MessageMinLength:    1,
				MessageMaxLength:    cryptoutilSharedMagic.JoseJAMaxMaterials,
				RecipientsMinCount:  0,
				RecipientsMaxCount:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			wantErr: "recipients-min-count must be >= 1, got 0",
		},
		{
			name: "recipients_max_less_than_min",
			settings: &SmIMServerSettings{
				MessageJWEAlgorithm: cryptoutilSharedMagic.JoseEncA256GCM,
				MessageMinLength:    1,
				MessageMaxLength:    cryptoutilSharedMagic.JoseJAMaxMaterials,
				RecipientsMinCount:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
				RecipientsMaxCount:  cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			},
			wantErr: "recipients-max-count (5) must be >= recipients-min-count (10)",
		},
		{
			name: "multiple_errors_aggregated",
			settings: &SmIMServerSettings{
				MessageJWEAlgorithm: "",
				MessageMinLength:    0,
				MessageMaxLength:    0,
				RecipientsMinCount:  0,
				RecipientsMaxCount:  0,
			},
			wantErr: "validation errors:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateSmIMSettings(tc.settings)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestValidateSmIMSettings_Valid tests that valid settings pass validation.
func TestValidateSmIMSettings_Valid(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()

	err := validateSmIMSettings(settings)
	require.NoError(t, err)
}

// TestParseWithFlagSet_DefaultPortOverride verifies that when --bind-public-port is NOT specified,
// the port is overridden to IMServicePort.
func TestParseWithFlagSet_DefaultPortOverride(t *testing.T) {
	// Sequential: uses viper global state.
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-default-port", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, []string{"start"}, false)

	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, uint16(cryptoutilSharedMagic.IMServicePort), settings.BindPublicPort)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSMIM, settings.OTLPService)
}

// TestParseWithFlagSet_TemplateParseFail verifies error wrapping when the template
// parser encounters an unknown flag.
func TestParseWithFlagSet_TemplateParseFail(t *testing.T) {
	// Sequential: uses viper global state.
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-parse-error", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, []string{"start", "--nonexistent-flag-xyz"}, false)

	require.Error(t, err)
	require.Nil(t, settings)
	require.Contains(t, err.Error(), "failed to parse template settings")
}

// TestParseWithFlagSet_ValidationFail verifies error wrapping when sm-im validation
// fails due to an invalid flag value passed through the CLI.
func TestParseWithFlagSet_ValidationFail(t *testing.T) {
	// Sequential: uses viper global state.
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-validation-error", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, []string{"start", "--message-min-length=0"}, false)

	require.Error(t, err)
	require.Nil(t, settings)
	require.Contains(t, err.Error(), "validation errors")
}
