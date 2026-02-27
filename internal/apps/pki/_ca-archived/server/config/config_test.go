// Copyright (c) 2025 Justin Cranford

// Package config tests for pki-ca server configuration.
package config

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewTestConfig_BasicCreation(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	require.NotNil(t, cfg, "config should not be nil")
	require.NotNil(t, cfg.ServiceTemplateServerSettings, "template settings should not be nil")

	// Verify bind settings.
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, cfg.BindPublicAddress)
	require.Equal(t, uint16(0), cfg.BindPublicPort)

	// Verify pki-ca specific settings.
	require.Equal(t, cryptoutilSharedMagic.OTLPServicePKICA, cfg.OTLPService)
	require.Empty(t, cfg.CAConfigPath)
	require.Empty(t, cfg.ProfilesPath)
	require.True(t, cfg.EnableEST)
	require.True(t, cfg.EnableOCSP)
	require.True(t, cfg.EnableCRL)
	require.False(t, cfg.EnableTimestamp)
}

func TestNewTestConfig_CustomPortAndAddress(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4AnyAddress, 8050, false)

	require.Equal(t, cryptoutilSharedMagic.IPv4AnyAddress, cfg.BindPublicAddress)
	require.Equal(t, uint16(8050), cfg.BindPublicPort)
	require.False(t, cfg.DevMode)
}

func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	require.NotNil(t, cfg, "default config should not be nil")
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, cfg.BindPublicAddress)
	require.Equal(t, uint16(0), cfg.BindPublicPort)
	require.True(t, cfg.DevMode)
	require.Equal(t, cryptoutilSharedMagic.OTLPServicePKICA, cfg.OTLPService)
}

func TestValidateCASettings_EmptyPaths(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	// Empty paths should pass validation.
	err := validateCASettings(cfg)
	require.NoError(t, err, "empty paths should pass validation")
}

func TestValidateCASettings_NonExistentCAConfigPath(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()
	cfg.CAConfigPath = "/nonexistent/path/to/ca-config.yaml"

	err := validateCASettings(cfg)
	require.Error(t, err, "non-existent CA config path should fail validation")
	require.Contains(t, err.Error(), "ca-config file does not exist")
}

func TestValidateCASettings_NonExistentProfilesPath(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()
	cfg.ProfilesPath = "/nonexistent/path/to/profiles"

	err := validateCASettings(cfg)
	require.Error(t, err, "non-existent profiles path should fail validation")
	require.Contains(t, err.Error(), "profiles-path does not exist")
}

func TestValidateCASettings_ProfilesPathIsFile(t *testing.T) {
	t.Parallel()
	// Create a temporary file (not a directory).
	tmpFile, err := os.CreateTemp("", "test-profiles-*.txt")
	require.NoError(t, err)

	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	cfg := DefaultTestConfig()
	cfg.ProfilesPath = tmpFile.Name()

	err = validateCASettings(cfg)
	require.Error(t, err, "profiles path that is a file should fail validation")
	require.Contains(t, err.Error(), "profiles-path is not a directory")
}

func TestValidateCASettings_ValidPaths(t *testing.T) {
	t.Parallel()
	// Create temporary directory for profiles.
	tmpDir, err := os.MkdirTemp("", "test-profiles-*")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Create temporary CA config file.
	tmpConfigFile := filepath.Join(tmpDir, "ca-config.yaml")
	err = os.WriteFile(tmpConfigFile, []byte("# CA config"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	cfg := DefaultTestConfig()
	cfg.CAConfigPath = tmpConfigFile
	cfg.ProfilesPath = tmpDir

	err = validateCASettings(cfg)
	require.NoError(t, err, "valid paths should pass validation")
}

func TestLogCASettings_NoError(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	// logCASettings should not panic with valid config.
	require.NotPanics(t, func() {
		logCASettings(cfg)
	}, "logCASettings should not panic")
}

func TestCAServerSettings_CASpecificDefaults(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	// Verify CA-specific boolean defaults.
	require.True(t, cfg.EnableEST, "EST should be enabled by default")
	require.True(t, cfg.EnableOCSP, "OCSP should be enabled by default")
	require.True(t, cfg.EnableCRL, "CRL should be enabled by default")
	require.False(t, cfg.EnableTimestamp, "Timestamp should be disabled by default")
}

func TestCAServerSettings_OTLPServiceName(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	// Verify OTLP service name is set to pki-ca.
	require.Equal(t, cryptoutilSharedMagic.OTLPServicePKICA, cfg.OTLPService, "OTLP service should be pki-ca")
}

func TestValidateCASettings_MultipleErrors(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()
	cfg.CAConfigPath = "/nonexistent/ca-config.yaml"
	cfg.ProfilesPath = "/nonexistent/profiles"

	err := validateCASettings(cfg)
	require.Error(t, err, "multiple invalid paths should fail validation")
	require.Contains(t, err.Error(), "ca-config file does not exist")
	require.Contains(t, err.Error(), "profiles-path does not exist")
}
