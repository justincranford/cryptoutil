// Copyright (c) 2025 Justin Cranford
//
//

// Package config provides configuration management for identity services.
package config

import (
	"time"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
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

// ServerConfig is a shared framework HTTP server configuration type.
type ServerConfig = cryptoutilAppsFrameworkServiceConfig.ServerConfig

// DatabaseConfig is a shared framework database configuration type.
type DatabaseConfig = cryptoutilAppsFrameworkServiceConfig.DatabaseConfig

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

// SessionConfig is a shared framework session configuration type.
type SessionConfig = cryptoutilAppsFrameworkServiceConfig.SessionConfig

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

// ObservabilityConfig is a shared framework observability configuration type.
type ObservabilityConfig = cryptoutilAppsFrameworkServiceConfig.ObservabilityConfig
