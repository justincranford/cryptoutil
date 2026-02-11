// Copyright (c) 2025 Justin Cranford
//
//

// Package config provides configuration management for identity services.
package config

import (
	"time"
)

// Config represents the identity module configuration.
type Config struct {
	// Server configuration.
	AuthZ *ServerConfig `yaml:"authz" json:"authz"` // Authorization server configuration.
	IDP   *ServerConfig `yaml:"idp" json:"idp"`     // Identity provider configuration.
	RS    *ServerConfig `yaml:"rs" json:"rs"`       // Resource server configuration.

	// Database configuration.
	Database *DatabaseConfig `yaml:"database" json:"database"` // Database configuration.

	// Token configuration.
	Tokens *TokenConfig `yaml:"tokens" json:"tokens"` // Token configuration.

	// Session configuration.
	Sessions *SessionConfig `yaml:"sessions" json:"sessions"` // Session configuration.

	// Security configuration.
	Security *SecurityConfig `yaml:"security" json:"security"` // Security configuration.

	// Observability configuration.
	Observability *ObservabilityConfig `yaml:"observability" json:"observability"` // Observability configuration.
}

// ServerConfig represents HTTP server configuration.
type ServerConfig struct {
	// Server identification.
	Name string `yaml:"name" json:"name"` // Server name.

	// Bind configuration.
	BindAddress string `yaml:"bind_address" json:"bind_address"` // Server bind address.
	Port        int    `yaml:"port" json:"port"`                 // Server port.

	// TLS configuration.
	TLSEnabled  bool   `yaml:"tls_enabled" json:"tls_enabled"`     // Enable TLS.
	TLSCertFile string `yaml:"tls_cert_file" json:"tls_cert_file"` // TLS certificate file.
	TLSKeyFile  string `yaml:"tls_key_file" json:"tls_key_file"`   // TLS private key file.

	// Timeouts.
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`   // Read timeout.
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"` // Write timeout.
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`   // Idle timeout.

	// Admin API configuration.
	AdminEnabled     bool   `yaml:"admin_enabled" json:"admin_enabled"`           // Enable admin API.
	AdminBindAddress string `yaml:"admin_bind_address" json:"admin_bind_address"` // Admin API bind address.
	AdminPort        int    `yaml:"admin_port" json:"admin_port"`                 // Admin API port.
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	// Database type.
	Type string `yaml:"type" json:"type"` // Database type (postgres, sqlite).

	// Connection string.
	DSN string `yaml:"dsn" json:"dsn"` // Database DSN.

	// Connection pool configuration.
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`         // Maximum open connections.
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`         // Maximum idle connections.
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`   // Connection max lifetime.
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"` // Connection max idle time.

	// Migration configuration.
	AutoMigrate bool `yaml:"auto_migrate" json:"auto_migrate"` // Enable auto-migration.
}

// TokenConfig represents token configuration.
type TokenConfig struct {
	// Token lifetimes.
	AccessTokenLifetime  time.Duration `yaml:"access_token_lifetime" json:"access_token_lifetime"`   // Access token lifetime.
	RefreshTokenLifetime time.Duration `yaml:"refresh_token_lifetime" json:"refresh_token_lifetime"` // Refresh token lifetime.
	IDTokenLifetime      time.Duration `yaml:"id_token_lifetime" json:"id_token_lifetime"`           // ID token lifetime.

	// Token formats.
	AccessTokenFormat  string `yaml:"access_token_format" json:"access_token_format"`   // Access token format (jws, jwe, uuid).
	RefreshTokenFormat string `yaml:"refresh_token_format" json:"refresh_token_format"` // Refresh token format (uuid).
	IDTokenFormat      string `yaml:"id_token_format" json:"id_token_format"`           // ID token format (jws).

	// JWT configuration.
	Issuer            string `yaml:"issuer" json:"issuer"`                         // Token issuer.
	SigningAlgorithm  string `yaml:"signing_algorithm" json:"signing_algorithm"`   // JWT signing algorithm.
	SigningKeyID      string `yaml:"signing_key_id" json:"signing_key_id"`         // JWT signing key ID.
	EncryptionEnabled bool   `yaml:"encryption_enabled" json:"encryption_enabled"` // Enable JWE encryption.
}

// SessionConfig represents session configuration.
type SessionConfig struct {
	// Session lifetime.
	SessionLifetime time.Duration `yaml:"session_lifetime" json:"session_lifetime"` // Session lifetime.
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`         // Session idle timeout.

	// Cookie configuration.
	CookieName     string `yaml:"cookie_name" json:"cookie_name"`           // Session cookie name.
	CookieDomain   string `yaml:"cookie_domain" json:"cookie_domain"`       // Session cookie domain.
	CookiePath     string `yaml:"cookie_path" json:"cookie_path"`           // Session cookie path.
	CookieSecure   bool   `yaml:"cookie_secure" json:"cookie_secure"`       // Session cookie secure flag.
	CookieHTTPOnly bool   `yaml:"cookie_http_only" json:"cookie_http_only"` // Session cookie HTTP-only flag.
	CookieSameSite string `yaml:"cookie_same_site" json:"cookie_same_site"` // Session cookie SameSite attribute.
}

// SecurityConfig represents security configuration.
type SecurityConfig struct {
	// PKCE configuration.
	RequirePKCE         bool   `yaml:"require_pkce" json:"require_pkce"`                   // Require PKCE for authorization code flow.
	PKCEChallengeMethod string `yaml:"pkce_challenge_method" json:"pkce_challenge_method"` // PKCE challenge method.

	// State parameter configuration.
	RequireState bool `yaml:"require_state" json:"require_state"` // Require state parameter.

	// Rate limiting.
	RateLimitEnabled  bool          `yaml:"rate_limit_enabled" json:"rate_limit_enabled"`   // Enable rate limiting.
	RateLimitRequests int           `yaml:"rate_limit_requests" json:"rate_limit_requests"` // Rate limit requests per window.
	RateLimitWindow   time.Duration `yaml:"rate_limit_window" json:"rate_limit_window"`     // Rate limit window.

	// CORS configuration.
	CORSEnabled        bool     `yaml:"cors_enabled" json:"cors_enabled"`                 // Enable CORS.
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins" json:"cors_allowed_origins"` // CORS allowed origins.

	// CSRF configuration.
	CSRFEnabled bool `yaml:"csrf_enabled" json:"csrf_enabled"` // Enable CSRF protection.
}

// ObservabilityConfig represents observability configuration.
type ObservabilityConfig struct {
	// Logging configuration.
	LogLevel  string `yaml:"log_level" json:"log_level"`   // Log level.
	LogFormat string `yaml:"log_format" json:"log_format"` // Log format (json, text).

	// Metrics configuration.
	MetricsEnabled bool   `yaml:"metrics_enabled" json:"metrics_enabled"` // Enable metrics.
	MetricsPath    string `yaml:"metrics_path" json:"metrics_path"`       // Metrics endpoint path.

	// Tracing configuration.
	TracingEnabled bool   `yaml:"tracing_enabled" json:"tracing_enabled"` // Enable tracing.
	TracingBackend string `yaml:"tracing_backend" json:"tracing_backend"` // Tracing backend.
}
