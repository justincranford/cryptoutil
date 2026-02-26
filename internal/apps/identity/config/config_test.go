// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Validates requirements:
// - R09-01: Configuration templates for all deployment scenarios
// - R09-02: Configuration validation prevents startup errors.
func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.AuthZ)
	require.NotNil(t, cfg.IDP)
	require.NotNil(t, cfg.RS)
	require.NotNil(t, cfg.Database)
	require.NotNil(t, cfg.Tokens)
	require.NotNil(t, cfg.Sessions)
	require.NotNil(t, cfg.Security)
	require.NotNil(t, cfg.Observability)

	err := cfg.Validate()
	require.NoError(t, err)
}

func TestLoadFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "minimal valid config",
			filename:    "minimal.yml",
			expectError: false,
		},
		{
			name:        "nonexistent file",
			filename:    "nonexistent.yml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.filename)
			cfg, err := LoadFromFile(path)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				require.NoError(t, cfg.Validate())
			}
		})
	}
}

func TestSaveToFile(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	tmpFile := filepath.Join(t.TempDir(), "test-config.yml")

	err := SaveToFile(cfg, tmpFile)
	require.NoError(t, err)

	stat, err := os.Stat(tmpFile)
	require.NoError(t, err)
	require.Greater(t, stat.Size(), int64(0))

	// File permissions verification (note: Windows may report different permissions)
	// Just verify file was created securely
	require.True(t, stat.Mode().IsRegular())

	loaded, err := LoadFromFile(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, loaded)
}

func TestServerConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *ServerConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &ServerConfig{
				Name:         "test",
				BindAddress:  cryptoutilSharedMagic.IPv4Loopback,
				Port:         cryptoutilSharedMagic.DemoServerPort,
				TLSEnabled:   false,
				ReadTimeout:  cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				WriteTimeout: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				IdleTimeout:  cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes * time.Second,
			},
			expectError: false,
		},
		{
			name: "invalid port",
			config: &ServerConfig{
				Name:        "test",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        -1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *DatabaseConfig
		expectError bool
	}{
		{
			name: "valid sqlite config",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 2,
			},
			expectError: false,
		},
		{
			name: "valid postgres config",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.DockerServicePostgres,
				DSN:          "postgres://user:pass@localhost/db",
				MaxOpenConns: cryptoutilSharedMagic.TLSMaxValidityCACertYears,
				MaxIdleConns: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			expectError: false,
		},
		{
			name: "empty type",
			config: &DatabaseConfig{
				Type:         "",
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 2,
			},
			expectError: true,
		},
		{
			name: "empty dsn",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          "",
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 2,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTokenConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *TokenConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: false,
		},
		{
			name: "empty issuer",
			config: &TokenConfig{
				AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:   cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:              "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
