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
func TestParse_EnvironmentVariables_DatabaseURL(t *testing.T) {
	resetFlags()

	// Set environment variable with CRYPTOUTIL_ prefix and underscores
	t.Setenv("CRYPTOUTIL_DATABASE_URL", "postgres://envuser:envpass@envhost:5432/envdb")

	commandParameters := []string{"start"}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, "postgres://envuser:envpass@envhost:5432/envdb", s.DatabaseURL, "environment variable should override default")
}

// Sequential: uses viper/pflag global state.
func TestParse_EnvironmentVariables_BindPublicPort(t *testing.T) {
	resetFlags()

	// Set environment variable for public port
	t.Setenv("CRYPTOUTIL_BIND_PUBLIC_PORT", "9999")

	commandParameters := []string{"start"}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, uint16(9999), s.BindPublicPort, "environment variable should override default")
}

// Sequential: uses viper/pflag global state.
func TestParse_EnvironmentVariables_LogLevel(t *testing.T) {
	resetFlags()

	// Set environment variable for log level
	t.Setenv("CRYPTOUTIL_LOG_LEVEL", "TRACE")

	commandParameters := []string{"start"}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, "TRACE", s.LogLevel, "environment variable should override default")
}

// Sequential: uses viper/pflag global state.
func TestParse_Precedence_FlagOverridesEnvVar(t *testing.T) {
	resetFlags()

	// Set environment variable
	t.Setenv("CRYPTOUTIL_BIND_PUBLIC_PORT", "9999")

	// Flag should take precedence over environment variable
	commandParameters := []string{"start", "--bind-public-port=7777"}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, uint16(7777), s.BindPublicPort, "flag should override environment variable")
}

func TestResolveFileURL_WithFilePrefix(t *testing.T) {
	t.Parallel()

	// Create temporary file with secret content
	tempFile := t.TempDir() + "/database_url.secret"
	secretContent := "postgres://secretuser:secretpass@secrethost:5432/secretdb"
	err := os.WriteFile(tempFile, []byte(secretContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Test file URL resolution
	result := resolveFileURL(cryptoutilSharedMagic.FileURIScheme + tempFile)
	require.Equal(t, secretContent, result, "file URL should resolve to file content")
}

func TestResolveFileURL_WithoutFilePrefix(t *testing.T) {
	t.Parallel()

	// Value without file:// prefix should return unchanged
	result := resolveFileURL("postgres://localhost:5432/db")
	require.Equal(t, "postgres://localhost:5432/db", result, "non-file URL should return unchanged")
}

func TestResolveFileURL_FileNotFound(t *testing.T) {
	t.Parallel()

	// Non-existent file should return original value and log warning
	result := resolveFileURL("file:///nonexistent/path/to/secret")
	require.Equal(t, "file:///nonexistent/path/to/secret", result, "missing file should return original value")
}

func TestResolveFileURL_WhitespaceTrimming(t *testing.T) {
	t.Parallel()

	// Create temporary file with whitespace around content
	tempFile := t.TempDir() + "/whitespace.secret"
	secretContent := "\n\t  postgres://trimmeduser:trimmedpass@trimmedhost:5432/trimmeddb  \t\n"
	err := os.WriteFile(tempFile, []byte(secretContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Test whitespace trimming
	result := resolveFileURL(cryptoutilSharedMagic.FileURIScheme + tempFile)
	require.Equal(t, "postgres://trimmeduser:trimmedpass@trimmedhost:5432/trimmeddb", result, "file content should be trimmed")
}

// Sequential: uses viper/pflag global state.
func TestParse_FileURL_DatabaseURL(t *testing.T) {
	resetFlags()

	// Create temporary file with database URL
	tempFile := t.TempDir() + "/database_url.secret"
	secretContent := "postgres://fileuser:filepass@filehost:5432/filedb"
	err := os.WriteFile(tempFile, []byte(secretContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Pass file URL via flag
	commandParameters := []string{"start", "--database-url=file://" + tempFile}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, secretContent, s.DatabaseURL, "file URL should resolve to file content")
}

// Sequential: uses viper/pflag global state.
func TestParse_Precedence_FullStack(t *testing.T) {
	resetFlags()

	// Create config file with database URL
	configDir := t.TempDir()
	configFile := configDir + "/config.yml"
	configContent := "database-url: postgres://configuser:configpass@confighost:5432/configdb\n"
	err := os.WriteFile(configFile, []byte(configContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Set environment variable (should override config file)
	t.Setenv("CRYPTOUTIL_DATABASE_URL", "postgres://envuser:envpass@envhost:5432/envdb")

	// Flag should override environment variable
	commandParameters := []string{"start", "--config=" + configFile, "--database-url=postgres://flaguser:flagpass@flaghost:5432/flagdb"}
	s, err := Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, "postgres://flaguser:flagpass@flaghost:5432/flagdb", s.DatabaseURL, "flag should have highest precedence")

	// Without flag, env var should override config
	resetFlags()

	commandParameters = []string{"start", "--config=" + configFile}
	s, err = Parse(commandParameters, true)
	require.NoError(t, err)
	require.Equal(t, "postgres://envuser:envpass@envhost:5432/envdb", s.DatabaseURL, "env var should override config file")
}
