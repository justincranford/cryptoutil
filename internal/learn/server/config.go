// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the learn-im HTTPS server using the service template.
package server

import (
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	// Default message validation constraints.
	DefaultMessageMinLength   = 1
	DefaultMessageMaxLength   = 10000
	DefaultRecipientsMinCount = 1
	DefaultRecipientsMaxCount = 10

	// DefaultJWEAlgorithm is the default JWE algorithm (using magic constant).
	DefaultJWEAlgorithm = cryptoutilMagic.LearnJWEAlgorithm

	// Default JWT secret (MUST be changed in production).
	DefaultJWTSecret = "learn-im-dev-secret-change-in-production"
)

// AppConfig holds configuration for the learn-im server.
// Embeds ServerSettings for network/TLS configuration and adds learn-im-specific settings.
type AppConfig struct {
	// ServerSettings provides standard server configuration.
	// Includes: network binding, TLS, CORS, CSRF, rate limiting, database, OTLP telemetry.
	cryptoutilConfig.ServerSettings

	// Learn-IM-specific settings.
	JWEAlgorithm       string `mapstructure:"jwe_algorithm" yaml:"jwe_algorithm"`               // JWE algorithm for message encryption (default: dir+A256GCM).
	MessageMinLength   int    `mapstructure:"message_min_length" yaml:"message_min_length"`     // Minimum message length in characters (default: 1).
	MessageMaxLength   int    `mapstructure:"message_max_length" yaml:"message_max_length"`     // Maximum message length in characters (default: 10000).
	RecipientsMinCount int    `mapstructure:"recipients_min_count" yaml:"recipients_min_count"` // Minimum recipients per message (default: 1).
	RecipientsMaxCount int    `mapstructure:"recipients_max_count" yaml:"recipients_max_count"` // Maximum recipients per message (default: 10).

	// Authentication settings.
	JWTSecret string `mapstructure:"jwt_secret" yaml:"jwt_secret"` // JWT signing secret for session tokens.
}

// DefaultAppConfig returns the default learn-im application configuration.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ServerSettings:     cryptoutilConfig.ServerSettings{},
		JWEAlgorithm:       DefaultJWEAlgorithm,
		MessageMinLength:   DefaultMessageMinLength,
		MessageMaxLength:   DefaultMessageMaxLength,
		RecipientsMinCount: DefaultRecipientsMinCount,
		RecipientsMaxCount: DefaultRecipientsMaxCount,
		JWTSecret:          DefaultJWTSecret, // TODO: Load from configuration file in Phase 10.
	}
}
