// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Sequential: uses viper/pflag global state.
func TestParseWithMultipleConfigFiles(t *testing.T) {
	resetFlags()

	// Create two temporary config files.
	configFile1 := t.TempDir() + "/config1.yaml"
	configFile2 := t.TempDir() + "/config2.yaml"

	// Write first config file (base settings).
	config1Content := `
log-level: INFO
bind-public-port: 8080
browser-ip-rate-limit: 100
`
	err := os.WriteFile(configFile1, []byte(config1Content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Write second config file (overrides).
	config2Content := `
log-level: DEBUG
bind-public-port: 9080
service-rate-limit: 200
`
	err = os.WriteFile(configFile2, []byte(config2Content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Parse with multiple config files (second should override first).
	commandParameters := []string{
		"start",
		"--config=" + configFile1,
		"--config=" + configFile2,
	}

	s, err := Parse(commandParameters, true)
	require.NoError(t, err)

	// Verify second config file values override first.
	require.Equal(t, "DEBUG", s.LogLevel, "second config should override log-level")
	require.Equal(t, uint16(9080), s.BindPublicPort, "second config should override bind-public-port")
	require.Equal(t, uint16(200), s.ServiceIPRateLimit, "second config should set service-rate-limit")
	require.Equal(t, uint16(cryptoutilSharedMagic.JoseJAMaxMaterials), s.BrowserIPRateLimit, "first config browser-ip-rate-limit should remain")
}

// TestFormatDefault_EmptyStringSlice tests formatDefault with empty []string.
// Kills mutation: config.go:1459 (CONDITIONALS_NEGATION: len(v) == 0 vs len(v) != 0).
func TestFormatDefault_EmptyStringSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "empty string slice",
			value:    []string{},
			expected: "[]",
		},
		{
			name:     "single element string slice",
			value:    []string{"one"},
			expected: "[one]",
		},
		{
			name:     "multi element string slice",
			value:    []string{"one", "two", "three"},
			expected: "[one,two,three]",
		},
		{
			name:     "empty string",
			value:    "",
			expected: `""`,
		},
		{
			name:     "non-empty string",
			value:    "test",
			expected: `"test"`,
		},
		{
			name:     "boolean true",
			value:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			value:    false,
			expected: "false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := formatDefault(tc.value)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestParse_BooleanEnvironmentVariableBinding tests that boolean settings are bound to environment variables.
// Kills mutation: config.go:949 (CONDITIONALS_NEGATION: if _, ok := setting.Value.(bool); ok).
// Sequential: uses viper/pflag global state.
func TestParse_BooleanEnvironmentVariableBinding(t *testing.T) {
	resetFlags()

	// Set environment variables for boolean settings
	t.Setenv("CRYPTOUTIL_DEV", "true")
	t.Setenv("CRYPTOUTIL_VERBOSE", "true")
	t.Setenv("CRYPTOUTIL_DRY_RUN", "true")

	// Parse without any command line flags (environment variables should be used)
	commandParameters := []string{"start"}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)

	// Verify boolean environment variables were bound and parsed correctly
	require.True(t, s.DevMode, "CRYPTOUTIL_DEV should set DevMode to true")
	require.True(t, s.VerboseMode, "CRYPTOUTIL_VERBOSE should set VerboseMode to true")
	require.True(t, s.DryRun, "CRYPTOUTIL_DRY_RUN should set DryRun to true")

	// Verify that flag overrides environment variable (precedence test)
	resetFlags()
	t.Setenv("CRYPTOUTIL_DEV", "true")

	commandParameters = []string{"start", "--dev=false"}
	s, err = Parse(commandParameters, true)
	require.NoError(t, err)
	require.False(t, s.DevMode, "flag should override environment variable")
}
