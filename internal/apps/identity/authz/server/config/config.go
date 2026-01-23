// Copyright (c) 2025 Justin Cranford

// Package config provides identity-authz server configuration settings.
package config

import (
	"fmt"
	"os"
	"strings"

	cryptoutilTemplateConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// IdentityAuthzServerSettings contains identity-authz specific configuration.
type IdentityAuthzServerSettings struct {
	*cryptoutilTemplateConfig.ServiceTemplateServerSettings

	// OAuth 2.1 Authorization Server settings.
	Issuer               string // OAuth 2.1 issuer URL (used in tokens and discovery).
	TokenLifetime        int    // Access token lifetime in seconds.
	RefreshTokenLifetime int    // Refresh token lifetime in seconds.
	AuthorizationCodeTTL int    // Authorization code TTL in seconds.

	// OIDC Discovery settings.
	EnableDiscovery bool // Enable OIDC Discovery endpoint.

	// Client registration settings.
	EnableDynamicRegistration bool // Enable dynamic client registration.
}

// Identity-Authz specific default values.
const (
	defaultIssuer               = "https://localhost:18000" // Default issuer URL.
	defaultTokenLifetime        = 3600                      // 1 hour access token lifetime.
	defaultRefreshTokenLifetime = 86400                     // 24 hour refresh token lifetime.
	defaultAuthorizationCodeTTL = 600                       // 10 minute authorization code TTL.
	defaultEnableDiscovery      = true                      // Enable OIDC discovery by default.
	defaultEnableDynReg         = false                     // Disable dynamic registration by default.
)

var allIdentityAuthzServerRegisteredSettings []*cryptoutilTemplateConfig.Setting //nolint:gochecknoglobals

// Identity-Authz specific Setting objects for parameter attributes.
var (
	issuerSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityAuthzServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "issuer",
		Shorthand:   "",
		Value:       defaultIssuer,
		Usage:       "OAuth 2.1 issuer URL (used in tokens and discovery)",
		Description: "Issuer URL",
	})
	tokenLifetimeSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityAuthzServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "token-lifetime",
		Shorthand:   "",
		Value:       defaultTokenLifetime,
		Usage:       "Access token lifetime in seconds",
		Description: "Token Lifetime",
	})
	refreshTokenLifetimeSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityAuthzServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "refresh-token-lifetime",
		Shorthand:   "",
		Value:       defaultRefreshTokenLifetime,
		Usage:       "Refresh token lifetime in seconds",
		Description: "Refresh Token Lifetime",
	})
	authorizationCodeTTLSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityAuthzServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "authorization-code-ttl",
		Shorthand:   "",
		Value:       defaultAuthorizationCodeTTL,
		Usage:       "Authorization code TTL in seconds",
		Description: "Authorization Code TTL",
	})
	enableDiscoverySetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityAuthzServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "enable-discovery",
		Shorthand:   "",
		Value:       defaultEnableDiscovery,
		Usage:       "Enable OIDC Discovery endpoint",
		Description: "Enable Discovery",
	})
	enableDynamicRegistrationSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityAuthzServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "enable-dynamic-registration",
		Shorthand:   "",
		Value:       defaultEnableDynReg,
		Usage:       "Enable dynamic client registration",
		Description: "Enable Dynamic Registration",
	})
)

// Parse parses command line arguments and returns identity-authz settings.
func Parse(args []string, exitIfHelp bool) (*IdentityAuthzServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilTemplateConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register identity-authz specific flags.
	pflag.StringP(issuerSetting.Name, issuerSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(issuerSetting), issuerSetting.Description)
	pflag.IntP(tokenLifetimeSetting.Name, tokenLifetimeSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsIntSetting(tokenLifetimeSetting), tokenLifetimeSetting.Description)
	pflag.IntP(refreshTokenLifetimeSetting.Name, refreshTokenLifetimeSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsIntSetting(refreshTokenLifetimeSetting), refreshTokenLifetimeSetting.Description)
	pflag.IntP(authorizationCodeTTLSetting.Name, authorizationCodeTTLSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsIntSetting(authorizationCodeTTLSetting), authorizationCodeTTLSetting.Description)
	pflag.BoolP(enableDiscoverySetting.Name, enableDiscoverySetting.Shorthand, cryptoutilTemplateConfig.RegisterAsBoolSetting(enableDiscoverySetting), enableDiscoverySetting.Description)
	pflag.BoolP(enableDynamicRegistrationSetting.Name, enableDynamicRegistrationSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsBoolSetting(enableDynamicRegistrationSetting), enableDynamicRegistrationSetting.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create identity-authz settings.
	settings := &IdentityAuthzServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		Issuer:                        viper.GetString(issuerSetting.Name),
		TokenLifetime:                 viper.GetInt(tokenLifetimeSetting.Name),
		RefreshTokenLifetime:          viper.GetInt(refreshTokenLifetimeSetting.Name),
		AuthorizationCodeTTL:          viper.GetInt(authorizationCodeTTLSetting.Name),
		EnableDiscovery:               viper.GetBool(enableDiscoverySetting.Name),
		EnableDynamicRegistration:     viper.GetBool(enableDynamicRegistrationSetting.Name),
	}

	// Override template defaults with identity-authz specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.IdentityAuthzServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityAuthz

	// Validate identity-authz specific settings.
	if err := validateIdentityAuthzSettings(settings); err != nil {
		return nil, fmt.Errorf("identity-authz settings validation failed: %w", err)
	}

	// Log identity-authz specific settings.
	logIdentityAuthzSettings(settings)

	return settings, nil
}

// validateIdentityAuthzSettings validates identity-authz specific configuration.
func validateIdentityAuthzSettings(s *IdentityAuthzServerSettings) error {
	var validationErrors []string

	// Validate issuer URL format.
	if s.Issuer == "" {
		validationErrors = append(validationErrors, "issuer URL is required")
	} else if !strings.HasPrefix(s.Issuer, "http://") && !strings.HasPrefix(s.Issuer, "https://") {
		validationErrors = append(validationErrors, "issuer URL must start with http:// or https://")
	}

	// Validate token lifetimes.
	if s.TokenLifetime <= 0 {
		validationErrors = append(validationErrors, "token-lifetime must be positive")
	}

	if s.RefreshTokenLifetime <= 0 {
		validationErrors = append(validationErrors, "refresh-token-lifetime must be positive")
	}

	if s.AuthorizationCodeTTL <= 0 {
		validationErrors = append(validationErrors, "authorization-code-ttl must be positive")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logIdentityAuthzSettings logs identity-authz specific configuration to stderr.
func logIdentityAuthzSettings(s *IdentityAuthzServerSettings) {
	fmt.Fprintf(os.Stderr, "Identity-Authz Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Issuer: %s\n", s.Issuer)
	fmt.Fprintf(os.Stderr, "  Token Lifetime: %ds\n", s.TokenLifetime)
	fmt.Fprintf(os.Stderr, "  Refresh Token Lifetime: %ds\n", s.RefreshTokenLifetime)
	fmt.Fprintf(os.Stderr, "  Authorization Code TTL: %ds\n", s.AuthorizationCodeTTL)
	fmt.Fprintf(os.Stderr, "  OIDC Discovery: %t\n", s.EnableDiscovery)
	fmt.Fprintf(os.Stderr, "  Dynamic Registration: %t\n", s.EnableDynamicRegistration)
}

// NewTestConfig creates an IdentityAuthzServerSettings instance for testing without calling Parse().
// This bypasses pflag's global FlagSet to allow multiple config creations in tests.
//
// Use this in tests instead of Parse() to avoid "flag redefined" panics
// when creating multiple server instances.
//
// Parameters:
//   - bindAddr: public bind address (typically cryptoutilSharedMagic.IPv4Loopback).
//   - bindPort: public bind port (use 0 for dynamic allocation).
//   - devMode: enable development mode (in-memory SQLite, relaxed security).
//
// Returns directly populated IdentityAuthzServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *IdentityAuthzServerSettings {
	// Get base template config.
	baseConfig := cryptoutilTemplateConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-authz specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityAuthz

	return &IdentityAuthzServerSettings{
		ServiceTemplateServerSettings: baseConfig,
		Issuer:                        defaultIssuer,
		TokenLifetime:                 defaultTokenLifetime,
		RefreshTokenLifetime:          defaultRefreshTokenLifetime,
		AuthorizationCodeTTL:          defaultAuthorizationCodeTTL,
		EnableDiscovery:               defaultEnableDiscovery,
		EnableDynamicRegistration:     defaultEnableDynReg,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentityAuthzServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
