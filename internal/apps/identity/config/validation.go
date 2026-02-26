// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.AuthZ != nil {
		if err := c.AuthZ.Validate(); err != nil {
			return fmt.Errorf("authz config: %w", err)
		}
	}

	if c.IDP != nil {
		if err := c.IDP.Validate(); err != nil {
			return fmt.Errorf("idp config: %w", err)
		}
	}

	if c.RS != nil {
		if err := c.RS.Validate(); err != nil {
			return fmt.Errorf("rs config: %w", err)
		}
	}

	if c.Database != nil {
		if err := c.Database.Validate(); err != nil {
			return fmt.Errorf("database config: %w", err)
		}
	}

	if c.Tokens != nil {
		if err := c.Tokens.Validate(); err != nil {
			return fmt.Errorf("tokens config: %w", err)
		}
	}

	if c.Sessions != nil {
		if err := c.Sessions.Validate(); err != nil {
			return fmt.Errorf("sessions config: %w", err)
		}
	}

	if c.Security != nil {
		if err := c.Security.Validate(); err != nil {
			return fmt.Errorf("security config: %w", err)
		}
	}

	if c.Observability != nil {
		if err := c.Observability.Validate(); err != nil {
			return fmt.Errorf("observability config: %w", err)
		}
	}

	return nil
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

// Validate validates token configuration.
func (tc *TokenConfig) Validate() error {
	if tc.AccessTokenLifetime <= 0 {
		return fmt.Errorf("access token lifetime must be positive")
	}

	if tc.RefreshTokenLifetime <= 0 {
		return fmt.Errorf("refresh token lifetime must be positive")
	}

	if tc.IDTokenLifetime <= 0 {
		return fmt.Errorf("ID token lifetime must be positive")
	}

	if tc.AccessTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatJWS && tc.AccessTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatJWE && tc.AccessTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatUUID {
		return fmt.Errorf("access token format must be 'jws', 'jwe', or 'uuid'")
	}

	if tc.RefreshTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatUUID {
		return fmt.Errorf("refresh token format must be 'uuid'")
	}

	if tc.IDTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatJWS {
		return fmt.Errorf("ID token format must be 'jws'")
	}

	if tc.Issuer == "" {
		return fmt.Errorf("token issuer is required")
	}

	if tc.SigningAlgorithm == "" {
		return fmt.Errorf("signing algorithm is required")
	}

	return nil
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

// Validate validates security configuration.
func (sc *SecurityConfig) Validate() error {
	if sc.PKCEChallengeMethod != cryptoutilSharedMagic.PKCEMethodS256 && sc.PKCEChallengeMethod != cryptoutilSharedMagic.PKCEMethodPlain {
		return fmt.Errorf("pKCE challenge method must be 'S256' or 'plain'")
	}

	if sc.RateLimitEnabled {
		if sc.RateLimitRequests <= 0 {
			return fmt.Errorf("rate limit requests must be positive")
		}

		if sc.RateLimitWindow <= 0 {
			return fmt.Errorf("rate limit window must be positive")
		}
	}

	return nil
}

// Validate validates observability configuration.
func (oc *ObservabilityConfig) Validate() error {
	if oc.LogLevel == "" {
		return fmt.Errorf("log level is required")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
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
