package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParse_HappyPath tests the Parse function with valid arguments.
func TestParse_HappyPath(t *testing.T) {
	// Create temporary config files.
	tempDir := t.TempDir()
	caConfigPath := filepath.Join(tempDir, "ca-config.yaml")
	profilesPath := filepath.Join(tempDir, "profiles")

	// Create the files.
	require.NoError(t, os.WriteFile(caConfigPath, []byte("{}"), 0o644))
	require.NoError(t, os.Mkdir(profilesPath, 0o755))

	args := []string{
		"start", // Required subcommand.
		"--ca-config", caConfigPath,
		"--profiles-path", profilesPath,
		"--enable-est=true",
		"--enable-ocsp=true",
		"--enable-crl=true",
		"--enable-timestamp=true",
		"--dev",
	}

	settings, err := Parse(args, false)
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, caConfigPath, settings.CAConfigPath)
	require.Equal(t, profilesPath, settings.ProfilesPath)
	require.True(t, settings.EnableEST)
	require.True(t, settings.EnableOCSP)
	require.True(t, settings.EnableCRL)
	require.True(t, settings.EnableTimestamp)
}

// TestParse_NonExistentCAConfig tests Parse with non-existent CA config file.
func TestParse_NonExistentCAConfig(t *testing.T) {
	t.Parallel()

	args := []string{
		"start", // Required subcommand.
		"--ca-config", "/nonexistent/ca-config.yaml",
		"--dev",
	}

	settings, err := Parse(args, false)
	require.Error(t, err)
	require.Nil(t, settings)
	require.Contains(t, err.Error(), "pki-ca settings validation failed")
	require.Contains(t, err.Error(), "ca-config file does not exist")
}

// TestParse_NonExistentProfilesPath tests Parse with non-existent profiles directory.
func TestParse_NonExistentProfilesPath(t *testing.T) {
	t.Parallel()

	args := []string{
		"start", // Required subcommand.
		"--profiles-path", "/nonexistent/profiles",
		"--dev",
	}

	settings, err := Parse(args, false)
	require.Error(t, err)
	require.Nil(t, settings)
	require.Contains(t, err.Error(), "pki-ca settings validation failed")
	require.Contains(t, err.Error(), "profiles directory does not exist")
}

// TestParse_ProfilesPathIsFile tests Parse when profiles path points to a file instead of directory.
func TestParse_ProfilesPathIsFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	profilesFile := filepath.Join(tempDir, "profiles.txt")
	require.NoError(t, os.WriteFile(profilesFile, []byte("not a directory"), 0o644))

	args := []string{
		"start", // Required subcommand.
		"--profiles-path", profilesFile,
		"--dev",
	}

	settings, err := Parse(args, false)
	require.Error(t, err)
	require.Nil(t, settings)
	require.Contains(t, err.Error(), "pki-ca settings validation failed")
	require.Contains(t, err.Error(), "profiles path is not a directory")
}

// TestParse_MultipleValidationErrors tests Parse with multiple validation failures.
func TestParse_MultipleValidationErrors(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	profilesFile := filepath.Join(tempDir, "profiles.txt")
	require.NoError(t, os.WriteFile(profilesFile, []byte("not a directory"), 0o644))

	args := []string{
		"start", // Required subcommand.
		"--ca-config", "/nonexistent/ca-config.yaml",
		"--profiles-path", profilesFile,
		"--dev",
	}

	settings, err := Parse(args, false)
	require.Error(t, err)
	require.Nil(t, settings)
	require.Contains(t, err.Error(), "pki-ca settings validation failed")
	// Should contain both errors.
	require.Contains(t, err.Error(), "ca-config file does not exist")
	require.Contains(t, err.Error(), "profiles path is not a directory")
}

// TestParse_DefaultValues tests Parse with default values (no flags).
func TestParse_DefaultValues(t *testing.T) {
	t.Parallel()

	args := []string{"start", "--dev"} // Required subcommand.

	settings, err := Parse(args, false)
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Empty(t, settings.CAConfigPath)
	require.Empty(t, settings.ProfilesPath)
	require.False(t, settings.EnableEST)
	require.True(t, settings.EnableOCSP)
	require.True(t, settings.EnableCRL)
	require.False(t, settings.EnableTimestamp)
}

// TestParse_EmptyPaths tests Parse with empty path strings (should pass validation).
func TestParse_EmptyPaths(t *testing.T) {
	t.Parallel()

	args := []string{
		"start", // Required subcommand.
		"--ca-config", "",
		"--profiles-path", "",
		"--dev",
	}

	settings, err := Parse(args, false)
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Empty(t, settings.CAConfigPath)
	require.Empty(t, settings.ProfilesPath)
}
