// Copyright (c) 2025 Justin Cranford

// Package config provides identity-rp server configuration settings.
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

// IdentityRPServerSettings contains identity-rp specific configuration.
type IdentityRPServerSettings struct {
	*cryptoutilTemplateConfig.ServiceTemplateServerSettings

	// OAuth 2.1 Provider settings.
	AuthzServerURL string // URL of the OAuth 2.1 authorization server.
	ClientID       string // OAuth 2.1 client ID for this RP.
	ClientSecret   string // OAuth 2.1 client secret (loaded from Docker secret).
	RedirectURI    string // OAuth 2.1 redirect URI for this RP.

	// BFF (Backend-for-Frontend) settings.
	SPAOrigin     string // Origin of the SPA frontend (for CORS).
	SessionSecret string // Secret for encrypting session cookies.
}

// Identity-RP specific default values.
const (
	defaultAuthzServerURL = "https://localhost:18000" // Default authorization server URL.
	defaultClientID       = ""                        // Must be configured.
	defaultClientSecret   = ""                        // Must be configured via Docker secret.
	defaultRedirectURI    = ""                        // Must be configured.
	defaultSPAOrigin      = "https://localhost:18400" // Default SPA origin.
	defaultSessionSecret  = ""                        // Must be configured via Docker secret.
)

var allIdentityRPServerRegisteredSettings []*cryptoutilTemplateConfig.Setting //nolint:gochecknoglobals

// Identity-RP specific Setting objects for parameter attributes.
var (
	authzServerURLSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "authz-server-url",
		Shorthand:   "",
		Value:       defaultAuthzServerURL,
		Usage:       "URL of the OAuth 2.1 authorization server",
		Description: "AuthZ Server URL",
	})
	clientIDSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "client-id",
		Shorthand:   "",
		Value:       defaultClientID,
		Usage:       "OAuth 2.1 client ID for this relying party",
		Description: "Client ID",
	})
	clientSecretSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "client-secret",
		Shorthand:   "",
		Value:       defaultClientSecret,
		Usage:       "OAuth 2.1 client secret (use file:///run/secrets/client_secret)",
		Description: "Client Secret",
	})
	redirectURISetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "redirect-uri",
		Shorthand:   "",
		Value:       defaultRedirectURI,
		Usage:       "OAuth 2.1 redirect URI for this relying party",
		Description: "Redirect URI",
	})
	spaOriginSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "spa-origin",
		Shorthand:   "",
		Value:       defaultSPAOrigin,
		Usage:       "origin of the SPA frontend for CORS",
		Description: "SPA Origin",
	})
	sessionSecretSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "session-secret",
		Shorthand:   "",
		Value:       defaultSessionSecret,
		Usage:       "secret for encrypting session cookies (use file:///run/secrets/session_secret)",
		Description: "Session Secret",
	})
)

// Parse parses command line arguments and returns identity-rp settings.
func Parse(args []string, exitIfHelp bool) (*IdentityRPServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilTemplateConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register identity-rp specific flags.
	pflag.StringP(authzServerURLSetting.Name, authzServerURLSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(authzServerURLSetting), authzServerURLSetting.Description)
	pflag.StringP(clientIDSetting.Name, clientIDSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(clientIDSetting), clientIDSetting.Description)
	pflag.StringP(clientSecretSetting.Name, clientSecretSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(clientSecretSetting), clientSecretSetting.Description)
	pflag.StringP(redirectURISetting.Name, redirectURISetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(redirectURISetting), redirectURISetting.Description)
	pflag.StringP(spaOriginSetting.Name, spaOriginSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(spaOriginSetting), spaOriginSetting.Description)
	pflag.StringP(sessionSecretSetting.Name, sessionSecretSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(sessionSecretSetting), sessionSecretSetting.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create identity-rp settings.
	settings := &IdentityRPServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		AuthzServerURL:                viper.GetString(authzServerURLSetting.Name),
		ClientID:                      viper.GetString(clientIDSetting.Name),
		ClientSecret:                  viper.GetString(clientSecretSetting.Name),
		RedirectURI:                   viper.GetString(redirectURISetting.Name),
		SPAOrigin:                     viper.GetString(spaOriginSetting.Name),
		SessionSecret:                 viper.GetString(sessionSecretSetting.Name),
	}

	// Override template defaults with identity-rp specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.IdentityRPServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityRP

	// Validate identity-rp specific settings.
	if err := validateIdentityRPSettings(settings); err != nil {
		return nil, fmt.Errorf("identity-rp settings validation failed: %w", err)
	}

	// Log identity-rp specific settings.
	logIdentityRPSettings(settings)

	return settings, nil
}

// validateIdentityRPSettings validates identity-rp specific configuration.
func validateIdentityRPSettings(s *IdentityRPServerSettings) error {
	var validationErrors []string

	// Validate AuthZ server URL if specified (required for production, optional for dev mode).
	if s.AuthzServerURL == "" && !s.DevMode {
		validationErrors = append(validationErrors, "authz-server-url is required in production mode")
	}

	// Validate SPA origin format if specified.
	if s.SPAOrigin != "" && !strings.HasPrefix(s.SPAOrigin, "http://") && !strings.HasPrefix(s.SPAOrigin, "https://") {
		validationErrors = append(validationErrors, "spa-origin must start with http:// or https://")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logIdentityRPSettings logs identity-rp specific configuration to stderr.
func logIdentityRPSettings(s *IdentityRPServerSettings) {
	fmt.Fprintf(os.Stderr, "Identity-RP Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  AuthZ Server URL: %s\n", s.AuthzServerURL)
	fmt.Fprintf(os.Stderr, "  Client ID: %s\n", s.ClientID)
	fmt.Fprintf(os.Stderr, "  Redirect URI: %s\n", s.RedirectURI)
	fmt.Fprintf(os.Stderr, "  SPA Origin: %s\n", s.SPAOrigin)
	fmt.Fprintf(os.Stderr, "  Session Secret: %s\n", maskSecret(s.SessionSecret))
}

// Secret masking constants.
const (
	secretMaskMinLength = 8 // Minimum length before showing partial secret.
	secretMaskPrefixLen = 4 // Number of characters to show at start.
)

// maskSecret masks a secret for logging (shows first 4 chars if long enough).
func maskSecret(secret string) string {
	if len(secret) == 0 {
		return "(not set)"
	} else if len(secret) <= secretMaskMinLength {
		return "****"
	}

	return secret[:secretMaskPrefixLen] + "****"
}

// NewTestConfig creates an IdentityRPServerSettings instance for testing without calling Parse().
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
// Returns directly populated IdentityRPServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *IdentityRPServerSettings {
	// Get base template config.
	baseConfig := cryptoutilTemplateConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-rp specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityRP

	return &IdentityRPServerSettings{
		ServiceTemplateServerSettings: baseConfig,
		AuthzServerURL:                defaultAuthzServerURL,
		ClientID:                      "",
		ClientSecret:                  "",
		RedirectURI:                   "",
		SPAOrigin:                     defaultSPAOrigin,
		SessionSecret:                 "",
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentityRPServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
