// Copyright (c) 2025 Justin Cranford

// Package config provides identity-rp server configuration settings.
package config

import (
	"fmt"
	"os"
	"strings"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
)

// IdentityRPServerSettings contains identity-rp specific configuration.
type IdentityRPServerSettings struct {
	*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings

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
	defaultAuthzServerURL = "https://localhost:8200" // Default authorization server URL.
	defaultClientID       = ""                       // Must be configured.
	defaultClientSecret   = ""                       // Must be configured via Docker secret.
	defaultRedirectURI    = ""                       // Must be configured.
	defaultSPAOrigin      = "https://localhost:8600" // Default SPA origin.
	defaultSessionSecret  = ""                       // Must be configured via Docker secret.
)

var allIdentityRPServerRegisteredSettings []*cryptoutilAppsFrameworkServiceConfig.Setting //nolint:gochecknoglobals

// Identity-RP specific Setting objects for parameter attributes.
var (
	authzServerURLSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "authz-server-url",
		Shorthand:   "",
		Value:       defaultAuthzServerURL,
		Usage:       "URL of the OAuth 2.1 authorization server",
		Description: "AuthZ Server URL",
	})
	clientIDSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "client-id",
		Shorthand:   "",
		Value:       defaultClientID,
		Usage:       "OAuth 2.1 client ID for this relying party",
		Description: "Client ID",
	})
	clientSecretSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "client-secret",
		Shorthand:   "",
		Value:       defaultClientSecret,
		Usage:       "OAuth 2.1 client secret (use file:///run/secrets/client_secret)",
		Description: "Client Secret",
	})
	redirectURISetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "redirect-uri",
		Shorthand:   "",
		Value:       defaultRedirectURI,
		Usage:       "OAuth 2.1 redirect URI for this relying party",
		Description: "Redirect URI",
	})
	spaOriginSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "spa-origin",
		Shorthand:   "",
		Value:       defaultSPAOrigin,
		Usage:       "origin of the SPA frontend for CORS",
		Description: "SPA Origin",
	})
	sessionSecretSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "session-secret",
		Shorthand:   "",
		Value:       defaultSessionSecret,
		Usage:       "secret for encrypting session cookies (use file:///run/secrets/session_secret)",
		Description: "Session Secret",
	})
)

// ParseWithFlagSet parses command line arguments using provided FlagSet and returns identity-rp settings.
// This enables test isolation by allowing each test to use its own FlagSet.
func ParseWithFlagSet(fs *pflag.FlagSet, args []string, exitIfHelp bool) (*IdentityRPServerSettings, error) {
	// Register identity-rp specific flags on the provided FlagSet BEFORE parsing.
	fs.StringP(authzServerURLSetting.Name, authzServerURLSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(authzServerURLSetting), authzServerURLSetting.Description)
	fs.StringP(clientIDSetting.Name, clientIDSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(clientIDSetting), clientIDSetting.Description)
	fs.StringP(clientSecretSetting.Name, clientSecretSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(clientSecretSetting), clientSecretSetting.Description)
	fs.StringP(redirectURISetting.Name, redirectURISetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(redirectURISetting), redirectURISetting.Description)
	fs.StringP(spaOriginSetting.Name, spaOriginSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(spaOriginSetting), spaOriginSetting.Description)
	fs.StringP(sessionSecretSetting.Name, sessionSecretSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(sessionSecretSetting), sessionSecretSetting.Description)

	// Parse base template settings using the same FlagSet.
	baseSettings, err := cryptoutilAppsFrameworkServiceConfig.ParseWithFlagSet(fs, args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	authzServerURL, _ := fs.GetString(authzServerURLSetting.Name)
	clientID, _ := fs.GetString(clientIDSetting.Name)
	clientSecret, _ := fs.GetString(clientSecretSetting.Name)
	redirectURI, _ := fs.GetString(redirectURISetting.Name)
	spaOrigin, _ := fs.GetString(spaOriginSetting.Name)
	sessionSecret, _ := fs.GetString(sessionSecretSetting.Name)

	settings := &IdentityRPServerSettings{
		ServiceFrameworkServerSettings: baseSettings,
		AuthzServerURL:                 authzServerURL,
		ClientID:                       clientID,
		ClientSecret:                   clientSecret,
		RedirectURI:                    redirectURI,
		SPAOrigin:                      spaOrigin,
		SessionSecret:                  sessionSecret,
	}

	if !fs.Changed("bind-public-port") {
		settings.BindPublicPort = cryptoutilSharedMagic.IdentityRPServicePort
	}

	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityRP

	if err := validateIdentityRPSettings(settings); err != nil {
		return nil, fmt.Errorf("identity-rp settings validation failed: %w", err)
	}

	logIdentityRPSettings(settings)

	return settings, nil
}

// Parse parses command-line arguments and returns the identity-rp server settings.
func Parse(args []string, exitIfHelp bool) (*IdentityRPServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, args, exitIfHelp)
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
	baseConfig := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-rp specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityRP

	return &IdentityRPServerSettings{
		ServiceFrameworkServerSettings: baseConfig,
		AuthzServerURL:                 defaultAuthzServerURL,
		ClientID:                       "",
		ClientSecret:                   "",
		RedirectURI:                    "",
		SPAOrigin:                      defaultSPAOrigin,
		SessionSecret:                  "",
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentityRPServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
