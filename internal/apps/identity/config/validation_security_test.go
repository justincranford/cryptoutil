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
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
				RateLimitEnabled:    false,
			},
			expectError: false,
		},
		{
			name: "valid_plain_pkce",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodPlain,
				RateLimitEnabled:    false,
			},
			expectError: false,
		},
		{
			name: "valid_with_rate_limiting",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
				RateLimitEnabled:    true,
				RateLimitRequests:   cryptoutilSharedMagic.JoseJAMaxMaterials,
				RateLimitWindow:     cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Second,
			},
			expectError: false,
		},
		{
			name: "invalid_pkce_challenge_method",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.SHA256,
				RateLimitEnabled:    false,
			},
			expectError: true,
			errorMsg:    "pKCE challenge method must be 'S256' or 'plain'",
		},
		{
			name: "rate_limit_enabled_invalid_requests_zero",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
				RateLimitEnabled:    true,
				RateLimitRequests:   0,
				RateLimitWindow:     cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Second,
			},
			expectError: true,
			errorMsg:    "rate limit requests must be positive",
		},
		{
			name: "rate_limit_enabled_invalid_requests_negative",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
				RateLimitEnabled:    true,
				RateLimitRequests:   -1,
				RateLimitWindow:     cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Second,
			},
			expectError: true,
			errorMsg:    "rate limit requests must be positive",
		},
		{
			name: "rate_limit_enabled_invalid_window_zero",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
				RateLimitEnabled:    true,
				RateLimitRequests:   cryptoutilSharedMagic.JoseJAMaxMaterials,
				RateLimitWindow:     0,
			},
			expectError: true,
			errorMsg:    "rate limit window must be positive",
		},
		{
			name: "rate_limit_enabled_invalid_window_negative",
			config: &SecurityConfig{
				RequirePKCE:         true,
				PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
				RateLimitEnabled:    true,
				RateLimitRequests:   cryptoutilSharedMagic.JoseJAMaxMaterials,
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
				LogLevel:       cryptoutilSharedMagic.StringError,
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
					BindAddress: cryptoutilSharedMagic.IPv4Loopback,
					Port:        cryptoutilSharedMagic.DemoServerPort,
				},
				IDP: &ServerConfig{
					Name:        "idp-server",
					BindAddress: cryptoutilSharedMagic.IPv4Loopback,
					Port:        cryptoutilSharedMagic.TestIDPServerPort,
				},
				RS: &ServerConfig{
					Name:        "rs-server",
					BindAddress: cryptoutilSharedMagic.IPv4Loopback,
					Port:        cryptoutilSharedMagic.TestResourceServerPort,
				},
				Database: &DatabaseConfig{
					Type:         cryptoutilSharedMagic.TestDatabaseSQLite,
					DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
					MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
					MaxIdleConns: 2,
				},
				Tokens: &TokenConfig{
					AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
					RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
					IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
					AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
					RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
					IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
					Issuer:               "https://example.com",
					SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
				},
				Sessions: &SessionConfig{
					SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
					IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
					CookieName:      "session_id",
					CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
				},
				Security: &SecurityConfig{
					RequirePKCE:         true,
					PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
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
					BindAddress: cryptoutilSharedMagic.IPv4Loopback,
					Port:        cryptoutilSharedMagic.DemoServerPort,
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
					Port:        cryptoutilSharedMagic.TestIDPServerPort,
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
					BindAddress: cryptoutilSharedMagic.IPv4Loopback,
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
					DSN:          cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
					MaxOpenConns: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
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
