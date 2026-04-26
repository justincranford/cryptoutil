// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ServerConfig represents HTTP server configuration shared across all PS-IDs.
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

// Validate validates server configuration.
func (sc *ServerConfig) Validate() error {
	if sc.Name == "" {
		return fmt.Errorf("server name is required")
	}

	if sc.BindAddress == "" {
		return fmt.Errorf("bind address is required")
	}

	if sc.Port <= 0 || sc.Port > int(cryptoutilSharedMagic.MaxPortNumber) {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if sc.TLSEnabled {
		if sc.TLSCertFile == "" {
			return fmt.Errorf("TLS cert file is required when TLS is enabled")
		}

		if sc.TLSKeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}

	if sc.AdminEnabled {
		if sc.AdminBindAddress == "" {
			return fmt.Errorf("admin bind address is required when admin is enabled")
		}

		if sc.AdminPort <= 0 || sc.AdminPort > int(cryptoutilSharedMagic.MaxPortNumber) {
			return fmt.Errorf("admin port must be between 1 and 65535")
		}
	}

	return nil
}

// DatabaseConfig represents database configuration shared across all PS-IDs.
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

// Validate validates database configuration.
func (dc *DatabaseConfig) Validate() error {
	if dc.Type == "" {
		return fmt.Errorf("database type is required")
	}

	if dc.Type != cryptoutilSharedMagic.DockerServicePostgres && dc.Type != "sqlite" {
		return fmt.Errorf("database type must be 'postgres' or 'sqlite'")
	}

	if dc.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	if dc.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}

	if dc.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be positive")
	}

	if dc.MaxIdleConns > dc.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}

	return nil
}

// SessionConfig represents session and cookie configuration shared across all PS-IDs.
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

// Validate validates session configuration.
func (sc *SessionConfig) Validate() error {
	if sc.SessionLifetime <= 0 {
		return fmt.Errorf("session lifetime must be positive")
	}

	if sc.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	if sc.CookieName == "" {
		return fmt.Errorf("cookie name is required")
	}

	if sc.CookieSameSite != cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict && sc.CookieSameSite != "Lax" && sc.CookieSameSite != "None" {
		return fmt.Errorf("cookie SameSite must be 'Strict', 'Lax', or 'None'")
	}

	return nil
}

// ObservabilityConfig represents observability configuration shared across all PS-IDs.
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

// Validate validates observability configuration.
func (oc *ObservabilityConfig) Validate() error {
	if oc.LogLevel == "" {
		return fmt.Errorf("log level is required")
	}

	validLogLevels := map[string]bool{
		"debug":                           true,
		"info":                            true,
		"warn":                            true,
		cryptoutilSharedMagic.StringError: true,
	}

	if !validLogLevels[oc.LogLevel] {
		return fmt.Errorf("log level must be 'debug', 'info', 'warn', or 'error'")
	}

	if oc.LogFormat != "json" && oc.LogFormat != "text" {
		return fmt.Errorf("log format must be 'json' or 'text'")
	}

	if oc.MetricsEnabled && oc.MetricsPath == "" {
		return fmt.Errorf("metrics path is required when metrics are enabled")
	}

	if oc.TracingEnabled && oc.TracingBackend == "" {
		return fmt.Errorf("tracing backend is required when tracing is enabled")
	}

	return nil
}
