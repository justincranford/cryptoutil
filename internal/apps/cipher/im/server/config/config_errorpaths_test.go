// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestValidateCipherImSettings_Errors tests validation error paths.
func TestValidateCipherImSettings_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *CipherImServerSettings
		wantErr  string
	}{
		{
			name: "empty_jwe_algorithm",
			settings: &CipherImServerSettings{
				MessageJWEAlgorithm: "",
				MessageMinLength:    1,
				MessageMaxLength:    100,
				RecipientsMinCount:  1,
				RecipientsMaxCount:  10,
			},
			wantErr: "message-jwe-algorithm cannot be empty",
		},
		{
			name: "message_min_length_zero",
			settings: &CipherImServerSettings{
				MessageJWEAlgorithm: "A256GCM",
				MessageMinLength:    0,
				MessageMaxLength:    100,
				RecipientsMinCount:  1,
				RecipientsMaxCount:  10,
			},
			wantErr: "message-min-length must be >= 1, got 0",
		},
		{
			name: "message_max_less_than_min",
			settings: &CipherImServerSettings{
				MessageJWEAlgorithm: "A256GCM",
				MessageMinLength:    100,
				MessageMaxLength:    50,
				RecipientsMinCount:  1,
				RecipientsMaxCount:  10,
			},
			wantErr: "message-max-length (50) must be >= message-min-length (100)",
		},
		{
			name: "recipients_min_count_zero",
			settings: &CipherImServerSettings{
				MessageJWEAlgorithm: "A256GCM",
				MessageMinLength:    1,
				MessageMaxLength:    100,
				RecipientsMinCount:  0,
				RecipientsMaxCount:  10,
			},
			wantErr: "recipients-min-count must be >= 1, got 0",
		},
		{
			name: "recipients_max_less_than_min",
			settings: &CipherImServerSettings{
				MessageJWEAlgorithm: "A256GCM",
				MessageMinLength:    1,
				MessageMaxLength:    100,
				RecipientsMinCount:  10,
				RecipientsMaxCount:  5,
			},
			wantErr: "recipients-max-count (5) must be >= recipients-min-count (10)",
		},
		{
			name: "multiple_errors_aggregated",
			settings: &CipherImServerSettings{
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

			err := validateCipherImSettings(tc.settings)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestValidateCipherImSettings_Valid tests that valid settings pass validation.
func TestValidateCipherImSettings_Valid(t *testing.T) {
	t.Parallel()

	settings := DefaultTestConfig()

	err := validateCipherImSettings(settings)
	require.NoError(t, err)
}

// TestParseWithFlagSet_DefaultPortOverride verifies that when --bind-public-port is NOT specified,
// the port is overridden to CipherServicePort.
func TestParseWithFlagSet_DefaultPortOverride(t *testing.T) {
	// Sequential: uses viper global state.
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })

	fs := pflag.NewFlagSet("test-default-port", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, []string{"start"}, false)

	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, uint16(cryptoutilSharedMagic.CipherServicePort), settings.BindPublicPort)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceCipherIM, settings.OTLPService)
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

// TestParseWithFlagSet_ValidationFail verifies error wrapping when cipher-im validation
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
