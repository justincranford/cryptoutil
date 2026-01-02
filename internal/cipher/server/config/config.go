// Copyright (c) 2025 Justin Cranford
//
//

// Package config implements the cipher-im HTTPS server using the service template.
package config

import (
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilTemplateServerRealms "cryptoutil/internal/template/server/realms"
)

// AppConfig holds configuration for the cipher-im server.
// Embeds ServerSettings for network/TLS configuration and adds cipher-im-specific settings.
type AppConfig struct {
	// ServerSettings provides standard server configuration.
	// Includes: network binding, TLS, CORS, CSRF, rate limiting, database, OTLP telemetry.
	cryptoutilConfig.ServerSettings

	// Cipher-IM-specific settings.
	JWEAlgorithm       string `mapstructure:"jwe_algorithm" yaml:"jwe_algorithm"`               // JWE algorithm for message encryption (default: dir+A256GCM).
	MessageMinLength   int    `mapstructure:"message_min_length" yaml:"message_min_length"`     // Minimum message length in characters (default: 1).
	MessageMaxLength   int    `mapstructure:"message_max_length" yaml:"message_max_length"`     // Maximum message length in characters (default: 10000).
	RecipientsMinCount int    `mapstructure:"recipients_min_count" yaml:"recipients_min_count"` // Minimum recipients per message (default: 1).
	RecipientsMaxCount int    `mapstructure:"recipients_max_count" yaml:"recipients_max_count"` // Maximum recipients per message (default: 10).

	// Authentication settings.
	JWTSecret string `mapstructure:"jwt_secret" yaml:"jwt_secret"` // JWT signing secret for session tokens.

	// Realm-based validation configuration (Phase 12).
	// Maps realm names to RealmConfig instances for multi-tenant validation rules.
	Realms map[string]*cryptoutilTemplateServerRealms.RealmConfig `mapstructure:"realms" yaml:"realms"` // Realm-specific validation and security configuration.
}

// DefaultAppConfig returns the default cipher-im application configuration.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ServerSettings:     cryptoutilConfig.ServerSettings{},
		JWEAlgorithm:       cryptoutilSharedMagic.CipherJWEAlgorithm,
		MessageMinLength:   cryptoutilSharedMagic.CipherMessageMinLength,
		MessageMaxLength:   cryptoutilSharedMagic.CipherMessageMaxLength,
		RecipientsMinCount: cryptoutilSharedMagic.CipherRecipientsMinCount,
		RecipientsMaxCount: cryptoutilSharedMagic.CipherRecipientsMaxCount,
		JWTSecret:          "", // MUST be provided at runtime (no default secret).
		Realms: map[string]*cryptoutilTemplateServerRealms.RealmConfig{
			"default":    cryptoutilTemplateServerRealms.DefaultRealm(),
			"enterprise": cryptoutilTemplateServerRealms.EnterpriseRealm(),
		},
	}
}

// GetRealmConfig retrieves a realm configuration by name with fallback to default.
// If realmName is empty or not found, returns the "default" realm.
func (cfg *AppConfig) GetRealmConfig(realmName string) *cryptoutilTemplateServerRealms.RealmConfig {
	if cfg.Realms == nil {
		return cryptoutilTemplateServerRealms.DefaultRealm()
	}

	// Try requested realm first.
	if realmName != "" {
		if realm, exists := cfg.Realms[realmName]; exists {
			return realm
		}
	}

	// Fall back to "default" realm.
	if defaultRealm, exists := cfg.Realms["default"]; exists {
		return defaultRealm
	}

	// Ultimate fallback to hardcoded default.
	return cryptoutilTemplateServerRealms.DefaultRealm()
}
