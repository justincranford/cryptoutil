// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

)

// TestNewFromFile_Success tests successful config loading from file.
// NOTE: Cannot use t.Parallel() - NewFromFile accesses global viper state.
func TestNewFromFile_Success(t *testing.T) {
	yamlContent := `
dev: true
bind-public-address: 127.0.0.1
bind-public-port: 8080
bind-private-address: 127.0.0.1
bind-private-port: 9090
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// NewFromFile now correctly includes "start" subcommand
	settings, err := NewFromFile(configPath)
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.True(t, settings.DevMode)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, settings.BindPublicAddress)
	require.Equal(t, uint16(cryptoutilSharedMagic.DemoServerPort), settings.BindPublicPort)
}

// TestNewFromFile_FileNotFound tests behavior when config file does not exist.
// Note: Viper gracefully handles missing config files by skipping them, so this
// does not return an error. The function returns valid settings with defaults.
// NOTE: Cannot use t.Parallel() - NewFromFile accesses global viper state.
func TestNewFromFile_FileNotFound(t *testing.T) {
	settings, err := NewFromFile("/nonexistent/path/config.yml")
	// Viper intentionally doesn't error on missing config files - they're optional
	require.NoError(t, err)
	require.NotNil(t, settings)
	// Verify defaults are applied when config file is missing
	require.Equal(t, "start", settings.SubCommand)
}

// TestNewFromFile_InvalidYAML tests error when config file has invalid YAML.
// NOTE: Cannot use t.Parallel() - NewFromFile accesses global viper state.
func TestNewFromFile_InvalidYAML(t *testing.T) {
	invalidYAML := `
dev: true
bind-public-address: [this is invalid YAML syntax
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yml")
	err := os.WriteFile(configPath, []byte(invalidYAML), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	_, err = NewFromFile(configPath)
	require.Error(t, err)
}

// TestRegisterAsBoolSetting_WrongType tests panic when setting value is not bool.
func TestRegisterAsBoolSetting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not bool")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not bool")
	}()

	setting := Setting{Name: "test-setting", Value: "not-a-bool"}
	_ = RegisterAsBoolSetting(&setting)
}

// TestRegisterAsStringSetting_WrongType tests panic when setting value is not string.
func TestRegisterAsStringSetting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not string")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not string")
	}()

	setting := Setting{Name: "test-setting", Value: 123}
	_ = RegisterAsStringSetting(&setting)
}

// TestRegisterAsUint16Setting_WrongType tests panic when setting value is not uint16.
func TestRegisterAsUint16Setting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not uint16")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not uint16")
	}()

	setting := Setting{Name: "test-setting", Value: "not-a-uint16"}
	_ = RegisterAsUint16Setting(&setting)
}

// TestRegisterAsStringSliceSetting_WrongType tests panic when setting value is not string slice.
func TestRegisterAsStringSliceSetting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not []string")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not []string")
	}()

	setting := Setting{Name: "test-setting", Value: "not-a-slice"}
	_ = RegisterAsStringSliceSetting(&setting)
}

// TestRegisterAsStringArraySetting_WrongType tests panic when setting value is not string array.
func TestRegisterAsStringArraySetting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not []string")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not []string")
	}()

	setting := Setting{Name: "test-setting", Value: 123}
	_ = RegisterAsStringArraySetting(&setting)
}

// TestRegisterAsDurationSetting_WrongType tests panic when setting value is not time.Duration.
func TestRegisterAsDurationSetting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not time.Duration")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not time.Duration")
	}()

	setting := Setting{Name: "test-setting", Value: "not-a-duration"}
	_ = RegisterAsDurationSetting(&setting)
}

// TestRegisterAsIntSetting_WrongType tests panic when setting value is not int.
func TestRegisterAsIntSetting_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic when value is not int")
		panicMsg, ok := r.(string)
		require.True(t, ok, "panic value should be a string")
		require.Contains(t, panicMsg, "value is not int")
	}()

	setting := Setting{Name: "test-setting", Value: "not-an-int"}
	_ = RegisterAsIntSetting(&setting)
}

// TestGetTLSPEMBytes_Base64DecodeError tests getTLSPEMBytes with invalid base64 string.
func TestGetTLSPEMBytes_Base64DecodeError(t *testing.T) {
	t.Parallel()

	viper.Set("test-invalid-base64", "!!!invalid-base64!!!")

	defer viper.Reset()

	result := getTLSPEMBytes("test-invalid-base64")
	require.Nil(t, result)
}

// TestGetTLSPEMBytes_EmptyString tests getTLSPEMBytes with empty string.
func TestGetTLSPEMBytes_EmptyString(t *testing.T) {
	viper.Set("test-empty-string", "")

	defer viper.Reset()

	result := getTLSPEMBytes("test-empty-string")
	require.Nil(t, result)
}

// TestGetTLSPEMBytes_ByteSliceValue tests getTLSPEMBytes with []byte value.
func TestGetTLSPEMBytes_ByteSliceValue(t *testing.T) {
	expected := []byte("test-bytes")
	viper.Set("test-byte-slice", expected)

	defer viper.Reset()

	result := getTLSPEMBytes("test-byte-slice")
	require.Equal(t, expected, result)
}

// TestGetTLSPEMBytes_ValidBase64 tests getTLSPEMBytes with valid base64 string.
func TestGetTLSPEMBytes_ValidBase64(t *testing.T) {
	original := []byte("test-pem-content")
	encoded := base64.StdEncoding.EncodeToString(original)
	viper.Set("test-valid-base64", encoded)

	defer viper.Reset()

	result := getTLSPEMBytes("test-valid-base64")
	require.Equal(t, original, result)
}

// TestNewTestConfig_DevMode tests NewTestConfig with dev mode enabled.
func TestNewTestConfig_DevMode(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, cryptoutilSharedMagic.DemoServerPort, true)
	require.NotNil(t, cfg)
	require.True(t, cfg.DevMode)
	require.Contains(t, cfg.DatabaseURL, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
}

// TestNewTestConfig_ProdMode tests NewTestConfig with dev mode disabled.
func TestNewTestConfig_ProdMode(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, cryptoutilSharedMagic.DemoServerPort, false)
	require.NotNil(t, cfg)
	require.False(t, cfg.DevMode)
	require.Contains(t, cfg.DatabaseURL, "postgres://")
}

// TestParseWithFlagSet_EmptyCommandParameters tests error when command parameters is empty.
func TestParseWithFlagSet_EmptyCommandParameters(t *testing.T) {
	// NOTE: Cannot use t.Parallel() here - ParseWithFlagSet modifies global viper state
	viper.Reset()

	defer viper.Reset()

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, err := ParseWithFlagSet(fs, []string{}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing subcommand")
}

// TestParseWithFlagSet_InvalidFlagSyntax tests error when parsing invalid flag syntax.
func TestParseWithFlagSet_InvalidFlagSyntax(t *testing.T) {
	// NOTE: Cannot use t.Parallel() here - ParseWithFlagSet modifies global viper state
	viper.Reset()

	defer viper.Reset()

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	// Invalid flag syntax: --unknown-flag will cause pflag.Parse to fail when flag doesn't exist
	// Actually need to test flag parsing error - use --bind-public-port with invalid value type
	_, err := ParseWithFlagSet(fs, []string{"start", "--bind-public-port=not-a-number"}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error parsing flags")
}

// TestParseWithFlagSet_ProfileKnown tests configuration profiles.
func TestParseWithFlagSet_ProfileKnown(t *testing.T) {
	// NOTE: Cannot use t.Parallel() here - ParseWithFlagSet modifies global viper state
	viper.Reset()

	defer viper.Reset()

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	// Use a known profile (dev, stg, prod, test)
	settings, err := ParseWithFlagSet(fs, []string{"start", "--profile=dev"}, false)
	require.NoError(t, err)
	require.NotNil(t, settings)
}

// TestParseWithFlagSet_ProfileUnknown tests error for unknown configuration profile.
