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

func TestTokenConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *TokenConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_jws_tokens",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: false,
		},
		{
			name: "valid_jwe_access_token",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jwe",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: false,
		},
		{
			name: "valid_uuid_access_token",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "uuid",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: false,
		},
		{
			name: "invalid_access_token_lifetime_zero",
			config: &TokenConfig{
				AccessTokenLifetime:  0,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "access token lifetime must be positive",
		},
		{
			name: "invalid_access_token_lifetime_negative",
			config: &TokenConfig{
				AccessTokenLifetime:  -1 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "access token lifetime must be positive",
		},
		{
			name: "invalid_refresh_token_lifetime_zero",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 0,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "refresh token lifetime must be positive",
		},
		{
			name: "invalid_id_token_lifetime_zero",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      0,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "ID token lifetime must be positive",
		},
		{
			name: "invalid_access_token_format",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "invalid",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "access token format must be 'jws', 'jwe', or 'uuid'",
		},
		{
			name: "invalid_refresh_token_format",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "jws",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "refresh token format must be 'uuid'",
		},
		{
			name: "invalid_id_token_format",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "uuid",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "ID token format must be 'jws'",
		},
		{
			name: "missing_issuer",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "",
				SigningAlgorithm:     "RS256",
			},
			expectError: true,
			errorMsg:    "token issuer is required",
		},
		{
			name: "missing_signing_algorithm",
			config: &TokenConfig{
				AccessTokenLifetime:  3600 * time.Second,
				RefreshTokenLifetime: 86400 * time.Second,
				IDTokenLifetime:      3600 * time.Second,
				AccessTokenFormat:    "jws",
				RefreshTokenFormat:   "uuid",
				IDTokenFormat:        "jws",
				Issuer:               "https://example.com",
				SigningAlgorithm:     "",
			},
			expectError: true,
			errorMsg:    "signing algorithm is required",
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

func TestSessionConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *SessionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_strict_samesite",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Strict",
				CookieSecure:    true,
				CookieHTTPOnly:  true,
			},
			expectError: false,
		},
		{
			name: "valid_lax_samesite",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Lax",
				CookieSecure:    true,
				CookieHTTPOnly:  true,
			},
			expectError: false,
		},
		{
			name: "valid_none_samesite",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "None",
				CookieSecure:    true,
				CookieHTTPOnly:  true,
			},
			expectError: false,
		},
		{
			name: "invalid_session_lifetime_zero",
			config: &SessionConfig{
				SessionLifetime: 0,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Strict",
			},
			expectError: true,
			errorMsg:    "session lifetime must be positive",
		},
		{
			name: "invalid_session_lifetime_negative",
			config: &SessionConfig{
				SessionLifetime: -1 * time.Second,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Strict",
			},
			expectError: true,
			errorMsg:    "session lifetime must be positive",
		},
		{
			name: "invalid_idle_timeout_zero",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     0,
				CookieName:      "session_id",
				CookieSameSite:  "Strict",
			},
			expectError: true,
			errorMsg:    "idle timeout must be positive",
		},
		{
			name: "invalid_idle_timeout_negative",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     -1 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Strict",
			},
			expectError: true,
			errorMsg:    "idle timeout must be positive",
		},
		{
			name: "missing_cookie_name",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "",
				CookieSameSite:  "Strict",
			},
			expectError: true,
			errorMsg:    "cookie name is required",
		},
		{
			name: "invalid_cookie_samesite",
			config: &SessionConfig{
				SessionLifetime: 3600 * time.Second,
				IdleTimeout:     1800 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Invalid",
			},
			expectError: true,
			errorMsg:    "cookie SameSite must be 'Strict', 'Lax', or 'None'",
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

func TestSecurityConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *SecurityConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_s256_pkce",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "S256",
				RateLimitEnabled:    false,
			},
			expectError: false,
		},
		{
			name: "valid_plain_pkce",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "plain",
				RateLimitEnabled:    false,
			},
			expectError: false,
		},
		{
			name: "valid_with_rate_limiting",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "S256",
				RateLimitEnabled:    true,
				RateLimitRequests:   100,
				RateLimitWindow:     60 * time.Second,
			},
			expectError: false,
		},
		{
			name: "invalid_pkce_challenge_method",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "SHA256",
				RateLimitEnabled:    false,
			},
			expectError: true,
			errorMsg:    "pKCE challenge method must be 'S256' or 'plain'",
		},
		{
			name: "rate_limit_enabled_invalid_requests_zero",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "S256",
				RateLimitEnabled:    true,
				RateLimitRequests:   0,
				RateLimitWindow:     60 * time.Second,
			},
			expectError: true,
			errorMsg:    "rate limit requests must be positive",
		},
		{
			name: "rate_limit_enabled_invalid_requests_negative",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "S256",
				RateLimitEnabled:    true,
				RateLimitRequests:   -1,
				RateLimitWindow:     60 * time.Second,
			},
			expectError: true,
			errorMsg:    "rate limit requests must be positive",
		},
		{
			name: "rate_limit_enabled_invalid_window_zero",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "S256",
				RateLimitEnabled:    true,
				RateLimitRequests:   100,
				RateLimitWindow:     0,
			},
			expectError: true,
			errorMsg:    "rate limit window must be positive",
		},
		{
			name: "rate_limit_enabled_invalid_window_negative",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: "S256",
				RateLimitEnabled:    true,
				RateLimitRequests:   100,
				RateLimitWindow:     -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "rate limit window must be positive",
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

func TestObservabilityConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *ObservabilityConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_minimal",
			config: &ObservabilityConfig{
				LogLevel:       "info",
				LogFormat:      "json",
				MetricsEnabled: false,
				TracingEnabled: false,
			},
			expectError: false,
		},
		{
			name: "valid_with_metrics",
			config: &ObservabilityConfig{
				LogLevel:       "debug",
				LogFormat:      "text",
				MetricsEnabled: true,
				MetricsPath:    "/metrics",
				TracingEnabled: false,
			},
			expectError: false,
		},
		{
			name: "valid_with_tracing",
			config: &ObservabilityConfig{
				LogLevel:       "warn",
				LogFormat:      "json",
				MetricsEnabled: false,
				TracingEnabled: true,
				TracingBackend: "jaeger",
			},
			expectError: false,
		},
		{
			name: "valid_all_features",
			config: &ObservabilityConfig{
				LogLevel:       "error",
				LogFormat:      "json",
				MetricsEnabled: true,
				MetricsPath:    "/metrics",
				TracingEnabled: true,
				TracingBackend: "otlp",
			},
			expectError: false,
		},
		{
			name: "missing_log_level",
			config: &ObservabilityConfig{
				LogLevel:       "",
				LogFormat:      "json",
				MetricsEnabled: false,
				TracingEnabled: false,
			},
			expectError: true,
			errorMsg:    "log level is required",
		},
		{
			name: "invalid_log_level",
			config: &ObservabilityConfig{
				LogLevel:       "trace",
				LogFormat:      "json",
				MetricsEnabled: false,
				TracingEnabled: false,
			},
			expectError: true,
			errorMsg:    "log level must be 'debug', 'info', 'warn', or 'error'",
		},
		{
			name: "invalid_log_format",
			config: &ObservabilityConfig{
				LogLevel:       "info",
				LogFormat:      "xml",
				MetricsEnabled: false,
				TracingEnabled: false,
			},
			expectError: true,
			errorMsg:    "log format must be 'json' or 'text'",
		},
		{
			name: "metrics_enabled_missing_path",
			config: &ObservabilityConfig{
				LogLevel:       "info",
				LogFormat:      "json",
				MetricsEnabled: true,
				MetricsPath:    "",
				TracingEnabled: false,
			},
			expectError: true,
			errorMsg:    "metrics path is required when metrics are enabled",
		},
		{
			name: "tracing_enabled_missing_backend",
			config: &ObservabilityConfig{
				LogLevel:       "info",
				LogFormat:      "json",
				MetricsEnabled: false,
				TracingEnabled: true,
				TracingBackend: "",
			},
			expectError: true,
			errorMsg:    "tracing backend is required when tracing is enabled",
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

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_full_config",
			config: &Config{
				AuthZ: &ServerConfig{
					Name:        "authz-server",
					BindAddress: "127.0.0.1",
					Port:        8080,
				},
				IDP: &ServerConfig{
					Name:        "idp-server",
					BindAddress: "127.0.0.1",
					Port:        8081,
				},
				RS: &ServerConfig{
					Name:        "rs-server",
					BindAddress: "127.0.0.1",
					Port:        8082,
				},
				Database: &DatabaseConfig{
					Type:         "sqlite",
					DSN:          ":memory:",
					MaxOpenConns: 5,
					MaxIdleConns: 2,
				},
				Tokens: &TokenConfig{
					AccessTokenLifetime:  3600 * time.Second,
					RefreshTokenLifetime: 86400 * time.Second,
					IDTokenLifetime:      3600 * time.Second,
					AccessTokenFormat:    "jws",
					RefreshTokenFormat:   "uuid",
					IDTokenFormat:        "jws",
					Issuer:               "https://example.com",
					SigningAlgorithm:     "RS256",
				},
				Sessions: &SessionConfig{
					SessionLifetime: 3600 * time.Second,
					IdleTimeout:     1800 * time.Second,
					CookieName:      "session_id",
					CookieSameSite:  "Strict",
				},
				Security: &SecurityConfig{
					RequirePKCE:         true,
					PKCEChallengeMethod: "S256",
					RateLimitEnabled:    false,
				},
				Observability: &ObservabilityConfig{
					LogLevel:       "info",
					LogFormat:      "json",
					MetricsEnabled: false,
					TracingEnabled: false,
				},
			},
			expectError: false,
		},
		{
			name: "invalid_authz_config",
			config: &Config{
				AuthZ: &ServerConfig{
					Name:        "",
					BindAddress: "127.0.0.1",
					Port:        8080,
				},
			},
			expectError: true,
			errorMsg:    "authz config",
		},
		{
			name: "invalid_idp_config",
			config: &Config{
				IDP: &ServerConfig{
					Name:        "idp-server",
					BindAddress: "",
					Port:        8081,
				},
			},
			expectError: true,
			errorMsg:    "idp config",
		},
		{
			name: "invalid_rs_config",
			config: &Config{
				RS: &ServerConfig{
					Name:        "rs-server",
					BindAddress: "127.0.0.1",
					Port:        0,
				},
			},
			expectError: true,
			errorMsg:    "rs config",
		},
		{
			name: "invalid_database_config",
			config: &Config{
				Database: &DatabaseConfig{
					Type:         "",
					DSN:          ":memory:",
					MaxOpenConns: 5,
					MaxIdleConns: 2,
				},
			},
			expectError: true,
			errorMsg:    "database config",
		},
		{
			name: "invalid_tokens_config",
			config: &Config{
				Tokens: &TokenConfig{
					AccessTokenLifetime: 0,
					Issuer:              "https://example.com",
				},
			},
			expectError: true,
			errorMsg:    "tokens config",
		},
		{
			name: "invalid_sessions_config",
			config: &Config{
				Sessions: &SessionConfig{
					SessionLifetime: 0,
					CookieName:      "session_id",
				},
			},
			expectError: true,
			errorMsg:    "sessions config",
		},
		{
			name: "invalid_security_config",
			config: &Config{
				Security: &SecurityConfig{
					PKCEChallengeMethod: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "security config",
		},
		{
			name: "invalid_observability_config",
			config: &Config{
				Observability: &ObservabilityConfig{
					LogLevel:  "",
					LogFormat: "json",
				},
			},
			expectError: true,
			errorMsg:    "observability config",
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
