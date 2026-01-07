// Copyright (c) 2025 Justin Cranford
//
//

// Package config implements the cipher-im HTTPS server using the service template.
package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateServerRealms "cryptoutil/internal/apps/template/service/server/realms"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Cipher-IM specific default values.
const (
	defaultJWEAlgorithm       = cryptoutilSharedMagic.CipherJWEAlgorithm
	defaultMessageMinLength   = cryptoutilSharedMagic.CipherMessageMinLength
	defaultMessageMaxLength   = cryptoutilSharedMagic.CipherMessageMaxLength
	defaultRecipientsMinCount = cryptoutilSharedMagic.CipherRecipientsMinCount
	defaultRecipientsMaxCount = cryptoutilSharedMagic.CipherRecipientsMaxCount
	defaultJWTSecret          = "" // MUST be provided at runtime (no default secret).
)

// CipherImServerSettings holds configuration for the cipher-im server.
// Embeds ServiceTemplateServerSettings for network/TLS configuration and adds cipher-im-specific settings.
type CipherImServerSettings struct {
	// ServiceTemplateServerSettings provides standard server configuration.
	// Includes: network binding, TLS, CORS, CSRF, rate limiting, database, OTLP telemetry.
	cryptoutilConfig.ServiceTemplateServerSettings

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

// Parse parses cipher-im configuration from config files, command-line parameters, and environment variables.
// This function extends the parent Parse() from ServiceTemplateServerSettings and adds cipher-im specific parameters.
//
// Parameters:
//   - parameters: Command-line arguments (flags)
//   - validateSubcommand: Whether to validate and extract the subcommand from parameters
//
// Returns:
//   - *CipherImServerSettings: Parsed configuration with cipher-im specific and template settings
//   - error: Any parsing or validation error
//
// Configuration Precedence (highest to lowest):
//  1. Command-line flags (--flag=value)
//  2. Environment variables (CRYPTOUTIL_FLAG_NAME)
//  3. Config files (YAML)
//  4. Default values
//
// Example usage:
//
//	settings, err := config.Parse(os.Args[1:], true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	server, err := server.NewFromConfig(ctx, settings)
func Parse(parameters []string, validateSubcommand bool) (*CipherImServerSettings, error) {
	// Parse base template settings first (network, TLS, database, OTLP, etc.).
	baseSettings, err := cryptoutilConfig.Parse(parameters, validateSubcommand)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template settings: %w", err)
	}

	// Register cipher-im specific flags with pflag.
	// These flags extend the base template settings with cipher-im specific configuration.
	pflag.String("jwe-algorithm", defaultJWEAlgorithm, "JWE algorithm for message encryption (e.g., dir+A256GCM)")
	pflag.Int("message-min-length", defaultMessageMinLength, "minimum message length in characters")
	pflag.Int("message-max-length", defaultMessageMaxLength, "maximum message length in characters")
	pflag.Int("recipients-min-count", defaultRecipientsMinCount, "minimum recipients per message")
	pflag.Int("recipients-max-count", defaultRecipientsMaxCount, "maximum recipients per message")
	pflag.String("jwt-secret", defaultJWTSecret, "JWT signing secret for session tokens (REQUIRED)")

	// Parse cipher-im specific flags.
	// Note: pflag has already parsed base flags in cryptoutilConfig.Parse(), so we just bind to viper.
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("failed to bind cipher-im flags: %w", err)
	}

	// Build CipherImServerSettings by embedding base settings and adding cipher-im specific values.
	settings := &CipherImServerSettings{
		ServiceTemplateServerSettings: *baseSettings,
		JWEAlgorithm:                  viper.GetString("jwe-algorithm"),
		MessageMinLength:              viper.GetInt("message-min-length"),
		MessageMaxLength:              viper.GetInt("message-max-length"),
		RecipientsMinCount:            viper.GetInt("recipients-min-count"),
		RecipientsMaxCount:            viper.GetInt("recipients-max-count"),
		JWTSecret:                     viper.GetString("jwt-secret"),
		Realms: map[string]*cryptoutilTemplateServerRealms.RealmConfig{
			"default":    cryptoutilTemplateServerRealms.DefaultRealm(),
			"enterprise": cryptoutilTemplateServerRealms.EnterpriseRealm(),
		},
	}

	// Override defaults for cipher-im service.
	// These overrides are applied only if the user didn't provide explicit values.
	if settings.BindPublicPort == cryptoutilSharedMagic.DefaultPublicPortCryptoutil {
		settings.BindPublicPort = cryptoutilSharedMagic.DefaultPublicPortCipherIM
	}

	if settings.BindPrivatePort == cryptoutilSharedMagic.DefaultPrivatePortCryptoutil {
		settings.BindPrivatePort = cryptoutilSharedMagic.DefaultPrivatePortCipherIM
	}

	if settings.OTLPService == cryptoutilSharedMagic.DefaultOTLPServiceDefault {
		settings.OTLPService = "cipher-im"
	}

	return settings, nil
}

// DefaultTestConfig returns a default CipherImServerSettings for testing purposes.
// This function should ONLY be used in tests, NOT in production code.
//
// For production code, ALWAYS use Parse() to properly parse configuration from
// config files, command-line parameters, and environment variables.
func DefaultTestConfig() *CipherImServerSettings {
	return &CipherImServerSettings{
		ServiceTemplateServerSettings: cryptoutilConfig.ServiceTemplateServerSettings{
			BindPublicPort:  cryptoutilSharedMagic.DefaultPublicPortCipherIM,
			BindPrivatePort: cryptoutilSharedMagic.DefaultPrivatePortCipherIM,
			TLSPublicMode:   cryptoutilConfig.TLSModeAuto, // Auto-generate TLS for tests.
			TLSPrivateMode:  cryptoutilConfig.TLSModeAuto, // Auto-generate TLS for tests.
			OTLPService:     "cipher-im",
			OTLPEnabled:     false,
		},
		JWEAlgorithm:       defaultJWEAlgorithm,
		MessageMinLength:   defaultMessageMinLength,
		MessageMaxLength:   defaultMessageMaxLength,
		RecipientsMinCount: defaultRecipientsMinCount,
		RecipientsMaxCount: defaultRecipientsMaxCount,
		JWTSecret:          "test-secret", // Test-only secret.
		Realms: map[string]*cryptoutilTemplateServerRealms.RealmConfig{
			"default":    cryptoutilTemplateServerRealms.DefaultRealm(),
			"enterprise": cryptoutilTemplateServerRealms.EnterpriseRealm(),
		},
	}
}

// GetRealmConfig retrieves a realm configuration by name with fallback to default.
// If realmName is empty or not found, returns the "default" realm.
func (cfg *CipherImServerSettings) GetRealmConfig(realmName string) *cryptoutilTemplateServerRealms.RealmConfig {
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
