// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
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
				BindAddress:  "127.0.0.1",
				Port:         8080,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			expectError: false,
		},
		{
			name: "valid_with_tls",
			config: &ServerConfig{
				Name:         "test-server",
				BindAddress:  "127.0.0.1",
				Port:         8100,
				TLSEnabled:   true,
				TLSCertFile:  "/path/to/cert.pem",
				TLSKeyFile:   "/path/to/key.pem",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			expectError: false,
		},
		{
			name: "valid_with_admin",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      "127.0.0.1",
				Port:             8080,
				AdminEnabled:     true,
				AdminBindAddress: "127.0.0.1",
				AdminPort:        9090,
				ReadTimeout:      30 * time.Second,
				WriteTimeout:     30 * time.Second,
				IdleTimeout:      120 * time.Second,
			},
			expectError: false,
		},
		{
			name: "missing_name",
			config: &ServerConfig{
				Name:        "",
				BindAddress: "127.0.0.1",
				Port:        8080,
			},
			expectError: true,
			errorMsg:    "server name is required",
		},
		{
			name: "missing_bind_address",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: "",
				Port:        8080,
			},
			expectError: true,
			errorMsg:    "bind address is required",
		},
		{
			name: "invalid_port_zero",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: "127.0.0.1",
				Port:        0,
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "invalid_port_negative",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: "127.0.0.1",
				Port:        -1,
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "invalid_port_too_high",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: "127.0.0.1",
				Port:        65536,
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "tls_enabled_missing_cert",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: "127.0.0.1",
				Port:        8100,
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
				BindAddress: "127.0.0.1",
				Port:        8100,
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
				BindAddress:      "127.0.0.1",
				Port:             8080,
				AdminEnabled:     true,
				AdminBindAddress: "",
				AdminPort:        9090,
			},
			expectError: true,
			errorMsg:    "admin bind address is required when admin is enabled",
		},
		{
			name: "admin_enabled_invalid_port_zero",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      "127.0.0.1",
				Port:             8080,
				AdminEnabled:     true,
				AdminBindAddress: "127.0.0.1",
				AdminPort:        0,
			},
			expectError: true,
			errorMsg:    "admin port must be between 1 and 65535",
		},
		{
			name: "admin_enabled_invalid_port_negative",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      "127.0.0.1",
				Port:             8080,
				AdminEnabled:     true,
				AdminBindAddress: "127.0.0.1",
				AdminPort:        -1,
			},
			expectError: true,
			errorMsg:    "admin port must be between 1 and 65535",
		},
		{
			name: "admin_enabled_invalid_port_too_high",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      "127.0.0.1",
				Port:             8080,
				AdminEnabled:     true,
				AdminBindAddress: "127.0.0.1",
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
				Type:         "sqlite",
				DSN:          ":memory:",
				MaxOpenConns: 5,
				MaxIdleConns: 2,
			},
			expectError: false,
		},
		{
			name: "valid_postgres",
			config: &DatabaseConfig{
				Type:         "postgres",
				DSN:          "postgres://user:pass@localhost/db",
				MaxOpenConns: 25,
				MaxIdleConns: 10,
			},
			expectError: false,
		},
		{
			name: "missing_type",
			config: &DatabaseConfig{
				Type:         "",
				DSN:          ":memory:",
				MaxOpenConns: 5,
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
				MaxOpenConns: 25,
				MaxIdleConns: 10,
			},
			expectError: true,
			errorMsg:    "database type must be 'postgres' or 'sqlite'",
		},
		{
			name: "missing_dsn",
			config: &DatabaseConfig{
				Type:         "sqlite",
				DSN:          "",
				MaxOpenConns: 5,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "database DSN is required",
		},
		{
			name: "invalid_max_open_conns_zero",
			config: &DatabaseConfig{
				Type:         "sqlite",
				DSN:          ":memory:",
				MaxOpenConns: 0,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "max open connections must be positive",
		},
		{
			name: "invalid_max_open_conns_negative",
			config: &DatabaseConfig{
				Type:         "sqlite",
				DSN:          ":memory:",
				MaxOpenConns: -1,
				MaxIdleConns: 2,
			},
			expectError: true,
			errorMsg:    "max open connections must be positive",
		},
		{
			name: "invalid_max_idle_conns_zero",
			config: &DatabaseConfig{
				Type:         "sqlite",
				DSN:          ":memory:",
				MaxOpenConns: 5,
				MaxIdleConns: 0,
			},
			expectError: true,
			errorMsg:    "max idle connections must be positive",
		},
		{
			name: "invalid_max_idle_conns_negative",
			config: &DatabaseConfig{
				Type:         "sqlite",
				DSN:          ":memory:",
				MaxOpenConns: 5,
				MaxIdleConns: -1,
			},
			expectError: true,
			errorMsg:    "max idle connections must be positive",
		},
		{
			name: "idle_exceeds_max_open",
			config: &DatabaseConfig{
				Type:         "sqlite",
				DSN:          ":memory:",
				MaxOpenConns: 5,
				MaxIdleConns: 10,
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
