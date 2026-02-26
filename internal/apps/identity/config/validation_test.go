// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *ServerConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_minimal_config",
			config: &ServerConfig{
				Name:         "test-server",
				BindAddress:  cryptoutilSharedMagic.IPv4Loopback,
				Port:         cryptoutilSharedMagic.DemoServerPort,
				ReadTimeout:  cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				WriteTimeout: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				IdleTimeout:  cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes * time.Second,
			},
			expectError: false,
		},
		{
			name: "valid_with_tls",
			config: &ServerConfig{
				Name:         "test-server",
				BindAddress:  cryptoutilSharedMagic.IPv4Loopback,
				Port:         cryptoutilSharedMagic.PKICAServicePort,
				TLSEnabled:   true,
				TLSCertFile:  "/path/to/cert.pem",
				TLSKeyFile:   "/path/to/key.pem",
				ReadTimeout:  cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				WriteTimeout: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				IdleTimeout:  cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes * time.Second,
			},
			expectError: false,
		},
		{
			name: "valid_with_admin",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             cryptoutilSharedMagic.DemoServerPort,
				AdminEnabled:     true,
				AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
				AdminPort:        cryptoutilSharedMagic.JoseJAAdminPort,
				ReadTimeout:      cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				WriteTimeout:     cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
				IdleTimeout:      cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes * time.Second,
			},
			expectError: false,
		},
		{
			name: "missing_name",
			config: &ServerConfig{
				Name:        "",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        cryptoutilSharedMagic.DemoServerPort,
			},
			expectError: true,
			errorMsg:    "server name is required",
		},
		{
			name: "missing_bind_address",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: "",
				Port:        cryptoutilSharedMagic.DemoServerPort,
			},
			expectError: true,
			errorMsg:    "bind address is required",
		},
		{
			name: "invalid_port_zero",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        0,
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "invalid_port_negative",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        -1,
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "invalid_port_too_high",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        65536,
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "tls_enabled_missing_cert",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        cryptoutilSharedMagic.PKICAServicePort,
				TLSEnabled:  true,
				TLSCertFile: "",
				TLSKeyFile:  "/path/to/key.pem",
			},
			expectError: true,
			errorMsg:    "TLS cert file is required when TLS is enabled",
		},
		{
			name: "tls_enabled_missing_key",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        cryptoutilSharedMagic.PKICAServicePort,
				TLSEnabled:  true,
				TLSCertFile: "/path/to/cert.pem",
				TLSKeyFile:  "",
			},
			expectError: true,
			errorMsg:    "TLS key file is required when TLS is enabled",
		},
		{
			name: "admin_enabled_missing_bind_address",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             cryptoutilSharedMagic.DemoServerPort,
				AdminEnabled:     true,
				AdminBindAddress: "",
				AdminPort:        cryptoutilSharedMagic.JoseJAAdminPort,
			},
			expectError: true,
			errorMsg:    "admin bind address is required when admin is enabled",
		},
		{
			name: "admin_enabled_invalid_port_zero",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             cryptoutilSharedMagic.DemoServerPort,
				AdminEnabled:     true,
				AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
				AdminPort:        0,
			},
			expectError: true,
			errorMsg:    "admin port must be between 1 and 65535",
		},
		{
			name: "admin_enabled_invalid_port_negative",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             cryptoutilSharedMagic.DemoServerPort,
				AdminEnabled:     true,
				AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
				AdminPort:        -1,
			},
			expectError: true,
			errorMsg:    "admin port must be between 1 and 65535",
		},
		{
			name: "admin_enabled_invalid_port_too_high",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             cryptoutilSharedMagic.DemoServerPort,
				AdminEnabled:     true,
				AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
				AdminPort:        65536,
			},
			expectError: true,
			errorMsg:    "admin port must be between 1 and 65535",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_sqlite",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 2,
			},
			expectError: false,
		},
		{
			name: "valid_postgres",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.DockerServicePostgres,
				DSN:          "postgres://user:pass@localhost/db",
				MaxOpenConns: cryptoutilSharedMagic.TLSMaxValidityCACertYears,
				MaxIdleConns: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			expectError: false,
		},
		{
			name: "missing_type",
			config: &DatabaseConfig{
				Type:         "",
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "database type is required",
		},
		{
			name: "invalid_type",
			config: &DatabaseConfig{
				Type:         "mysql",
				DSN:          "mysql://user:pass@localhost/db",
				MaxOpenConns: cryptoutilSharedMagic.TLSMaxValidityCACertYears,
				MaxIdleConns: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			expectError: true,
			errorMsg:    "database type must be 'postgres' or 'sqlite'",
		},
		{
			name: "missing_dsn",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          "",
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "database DSN is required",
		},
		{
			name: "invalid_max_open_conns_zero",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: 0,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "max open connections must be positive",
		},
		{
			name: "invalid_max_open_conns_negative",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: -1,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "max open connections must be positive",
		},
		{
			name: "invalid_max_idle_conns_zero",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: 0,
			},
			expectError: true,
			errorMsg:    "max idle connections must be positive",
		},
		{
			name: "invalid_max_idle_conns_negative",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: -1,
			},
			expectError: true,
			errorMsg:    "max idle connections must be positive",
		},
		{
			name: "idle_exceeds_max_open",
			config: &DatabaseConfig{
				Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
				MaxIdleConns: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
			expectError: true,
			errorMsg:    "max idle connections cannot exceed max open connections",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
