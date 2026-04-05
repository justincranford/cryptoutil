// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestServerConfig_Validate(t *testing.T) {
	t.Parallel()

	validBase := &ServerConfig{
		Name:        "test-server",
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        int(cryptoutilSharedMagic.TestServerPort),
	}

	tests := []struct {
		name      string
		config    *ServerConfig
		wantError string
	}{
		{
			name:   "valid_minimal",
			config: validBase,
		},
		{
			name: "valid_with_tls",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        int(cryptoutilSharedMagic.TestServerPort),
				TLSEnabled:  true,
				TLSCertFile: "/path/to/cert.pem",
				TLSKeyFile:  "/path/to/key.pem",
			},
		},
		{
			name: "valid_with_admin",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             int(cryptoutilSharedMagic.TestServerPort),
				AdminEnabled:     true,
				AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
				AdminPort:        int(cryptoutilSharedMagic.JoseJAAdminPort),
			},
		},
		{
			name:      "missing_name",
			config:    &ServerConfig{Name: "", BindAddress: cryptoutilSharedMagic.IPv4Loopback, Port: int(cryptoutilSharedMagic.TestServerPort)},
			wantError: "server name is required",
		},
		{
			name:      "missing_bind_address",
			config:    &ServerConfig{Name: "test-server", BindAddress: "", Port: int(cryptoutilSharedMagic.TestServerPort)},
			wantError: "bind address is required",
		},
		{
			name:      "port_zero",
			config:    &ServerConfig{Name: "test-server", BindAddress: cryptoutilSharedMagic.IPv4Loopback, Port: 0},
			wantError: "port must be between 1 and 65535",
		},
		{
			name:      "port_too_high",
			config:    &ServerConfig{Name: "test-server", BindAddress: cryptoutilSharedMagic.IPv4Loopback, Port: int(cryptoutilSharedMagic.MaxPortNumber) + 1},
			wantError: "port must be between 1 and 65535",
		},
		{
			name: "tls_enabled_missing_cert",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        int(cryptoutilSharedMagic.TestServerPort),
				TLSEnabled:  true,
				TLSKeyFile:  "/path/to/key.pem",
			},
			wantError: "TLS cert file is required when TLS is enabled",
		},
		{
			name: "tls_enabled_missing_key",
			config: &ServerConfig{
				Name:        "test-server",
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				Port:        int(cryptoutilSharedMagic.TestServerPort),
				TLSEnabled:  true,
				TLSCertFile: "/path/to/cert.pem",
			},
			wantError: "TLS key file is required when TLS is enabled",
		},
		{
			name: "admin_enabled_missing_address",
			config: &ServerConfig{
				Name:         "test-server",
				BindAddress:  cryptoutilSharedMagic.IPv4Loopback,
				Port:         int(cryptoutilSharedMagic.TestServerPort),
				AdminEnabled: true,
				AdminPort:    int(cryptoutilSharedMagic.JoseJAAdminPort),
			},
			wantError: "admin bind address is required when admin is enabled",
		},
		{
			name: "admin_enabled_invalid_port",
			config: &ServerConfig{
				Name:             "test-server",
				BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
				Port:             int(cryptoutilSharedMagic.TestServerPort),
				AdminEnabled:     true,
				AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
				AdminPort:        0,
			},
			wantError: "admin port must be between 1 and 65535",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    *DatabaseConfig
		wantError string
	}{
		{
			name:   "valid_postgres",
			config: &DatabaseConfig{Type: cryptoutilSharedMagic.DockerServicePostgres, DSN: "postgres://user:pass@host/db", MaxOpenConns: cryptoutilSharedMagic.PostgreSQLMaxIdleConns, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
		},
		{
			name:   "valid_sqlite",
			config: &DatabaseConfig{Type: cryptoutilSharedMagic.TestDatabaseSQLite, DSN: "file::memory:", MaxOpenConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
		},
		{
			name:      "empty_type",
			config:    &DatabaseConfig{DSN: "postgres://user:pass@host/db", MaxOpenConns: cryptoutilSharedMagic.PostgreSQLMaxIdleConns, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
			wantError: "database type is required",
		},
		{
			name:      "invalid_type",
			config:    &DatabaseConfig{Type: "mysql", DSN: "mysql://user:pass@host/db", MaxOpenConns: cryptoutilSharedMagic.PostgreSQLMaxIdleConns, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
			wantError: "database type must be 'postgres' or 'sqlite'",
		},
		{
			name:      "empty_dsn",
			config:    &DatabaseConfig{Type: cryptoutilSharedMagic.TestDatabaseSQLite, MaxOpenConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
			wantError: "database DSN is required",
		},
		{
			name:      "zero_max_open_conns",
			config:    &DatabaseConfig{Type: cryptoutilSharedMagic.TestDatabaseSQLite, DSN: "file::memory:", MaxOpenConns: 0, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
			wantError: "max open connections must be positive",
		},
		{
			name:      "zero_max_idle_conns",
			config:    &DatabaseConfig{Type: cryptoutilSharedMagic.TestDatabaseSQLite, DSN: "file::memory:", MaxOpenConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM, MaxIdleConns: 0},
			wantError: "max idle connections must be positive",
		},
		{
			name:      "idle_exceeds_open",
			config:    &DatabaseConfig{Type: cryptoutilSharedMagic.TestDatabaseSQLite, DSN: "file::memory:", MaxOpenConns: 3, MaxIdleConns: cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM},
			wantError: "max idle connections cannot exceed max open connections",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSessionConfig_Validate(t *testing.T) {
	t.Parallel()

	validBase := &SessionConfig{
		SessionLifetime: cryptoutilSharedMagic.DefaultSessionCleanupInterval,
		IdleTimeout:     cryptoutilSharedMagic.DefaultIdleTimeout,
		CookieName:      "session",
		CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
	}

	tests := []struct {
		name      string
		config    *SessionConfig
		wantError string
	}{
		{
			name:   "valid_strict",
			config: validBase,
		},
		{
			name: "valid_lax",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.DefaultSessionCleanupInterval,
				IdleTimeout:     cryptoutilSharedMagic.DefaultIdleTimeout,
				CookieName:      "session",
				CookieSameSite:  "Lax",
			},
		},
		{
			name: "valid_none",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.DefaultSessionCleanupInterval,
				IdleTimeout:     cryptoutilSharedMagic.DefaultIdleTimeout,
				CookieName:      "session",
				CookieSameSite:  "None",
			},
		},
		{
			name: "zero_session_lifetime",
			config: &SessionConfig{
				SessionLifetime: 0,
				IdleTimeout:     cryptoutilSharedMagic.DefaultIdleTimeout,
				CookieName:      "session",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			wantError: "session lifetime must be positive",
		},
		{
			name: "zero_idle_timeout",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.DefaultSessionCleanupInterval,
				IdleTimeout:     0,
				CookieName:      "session",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			wantError: "idle timeout must be positive",
		},
		{
			name: "empty_cookie_name",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.DefaultSessionCleanupInterval,
				IdleTimeout:     cryptoutilSharedMagic.DefaultIdleTimeout,
				CookieName:      "",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			wantError: "cookie name is required",
		},
		{
			name: "invalid_same_site",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.DefaultSessionCleanupInterval,
				IdleTimeout:     cryptoutilSharedMagic.DefaultIdleTimeout,
				CookieName:      "session",
				CookieSameSite:  "Invalid",
			},
			wantError: "cookie SameSite must be 'Strict', 'Lax', or 'None'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestObservabilityConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    *ObservabilityConfig
		wantError string
	}{
		{
			name:   "valid_debug",
			config: &ObservabilityConfig{LogLevel: "debug", LogFormat: "json"},
		},
		{
			name:   "valid_info_text",
			config: &ObservabilityConfig{LogLevel: "info", LogFormat: "text"},
		},
		{
			name:   "valid_warn",
			config: &ObservabilityConfig{LogLevel: "warn", LogFormat: "json"},
		},
		{
			name:   "valid_error",
			config: &ObservabilityConfig{LogLevel: cryptoutilSharedMagic.StringError, LogFormat: "json"},
		},
		{
			name:   "valid_metrics_enabled",
			config: &ObservabilityConfig{LogLevel: "info", LogFormat: "json", MetricsEnabled: true, MetricsPath: "/metrics"},
		},
		{
			name:   "valid_tracing_enabled",
			config: &ObservabilityConfig{LogLevel: "info", LogFormat: "json", TracingEnabled: true, TracingBackend: "otlp"},
		},
		{
			name:      "empty_log_level",
			config:    &ObservabilityConfig{LogLevel: "", LogFormat: "json"},
			wantError: "log level is required",
		},
		{
			name:      "invalid_log_level",
			config:    &ObservabilityConfig{LogLevel: "verbose", LogFormat: "json"},
			wantError: "log level must be 'debug', 'info', 'warn', or 'error'",
		},
		{
			name:      "invalid_log_format",
			config:    &ObservabilityConfig{LogLevel: "info", LogFormat: "xml"},
			wantError: "log format must be 'json' or 'text'",
		},
		{
			name:      "metrics_enabled_missing_path",
			config:    &ObservabilityConfig{LogLevel: "info", LogFormat: "json", MetricsEnabled: true, MetricsPath: ""},
			wantError: "metrics path is required when metrics are enabled",
		},
		{
			name:      "tracing_enabled_missing_backend",
			config:    &ObservabilityConfig{LogLevel: "info", LogFormat: "json", TracingEnabled: true, TracingBackend: ""},
			wantError: "tracing backend is required when tracing is enabled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
