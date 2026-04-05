// Copyright (c) 2025 Justin Cranford

// Package config provides identity-rs server configuration settings.
package config

import (
	"fmt"
	"os"
	"strings"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
)

// IdentityRSServerSettings contains identity-rs specific configuration.
type IdentityRSServerSettings struct {
	*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings

	// Token validation settings.
	AuthzServerURL   string // URL of the OAuth 2.1 authorization server for token validation.
	JWKSEndpoint     string // JWKS endpoint for token signature validation.
	IntrospectionURL string // Token introspection endpoint (optional).

	// Access control settings.
	RequiredScopes    []string // Required OAuth scopes for all endpoints.
	RequiredAudiences []string // Required audiences in access tokens.
	AllowBearerToken  bool     // Allow Bearer token authentication.
	AllowClientCert   bool     // Allow mTLS client certificate authentication.

	// Caching settings.
	JWKSCacheTTL       int  // JWKS cache TTL in seconds.
	TokenCacheTTL      int  // Validated token cache TTL in seconds.
	EnableTokenCaching bool // Enable token validation result caching.
}

// Identity-RS specific default values.
const (
	defaultRSAuthzServerURL   = "https://localhost:8200"       // Default authorization server URL.
	defaultJWKSEndpoint       = cryptoutilSharedMagic.PathJWKS // Standard JWKS endpoint path.
	defaultIntrospectionURL   = ""                             // No introspection by default.
	defaultAllowBearerToken   = true                           // Allow Bearer tokens by default.
	defaultAllowClientCert    = false                          // Disable client cert auth by default.
	defaultJWKSCacheTTL       = 3600                           // 1 hour JWKS cache.
	defaultTokenCacheTTL      = 300                            // 5 minute token cache.
	defaultEnableTokenCaching = true                           // Enable caching by default.
)

var (
	defaultRequiredScopes    = []string{} // No required scopes by default.
	defaultRequiredAudiences = []string{} // No required audiences by default.
)

var allIdentityRSServerRegisteredSettings []*cryptoutilAppsFrameworkServiceConfig.Setting //nolint:gochecknoglobals

// Identity-RS specific Setting objects for parameter attributes.
var (
	rsAuthzServerURLSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "authz-server-url",
		Shorthand:   "",
		Value:       defaultRSAuthzServerURL,
		Usage:       "URL of the OAuth 2.1 authorization server for token validation",
		Description: "AuthZ Server URL",
	})
	jwksEndpointSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "jwks-endpoint",
		Shorthand:   "",
		Value:       defaultJWKSEndpoint,
		Usage:       "JWKS endpoint path for token signature validation",
		Description: "JWKS Endpoint",
	})
	introspectionURLSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "introspection-url",
		Shorthand:   "",
		Value:       defaultIntrospectionURL,
		Usage:       "Token introspection endpoint URL (optional)",
		Description: "Introspection URL",
	})
	allowBearerTokenSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "allow-bearer-token",
		Shorthand:   "",
		Value:       defaultAllowBearerToken,
		Usage:       "Allow Bearer token authentication",
		Description: "Allow Bearer Token",
	})
	allowClientCertSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "allow-client-cert",
		Shorthand:   "",
		Value:       defaultAllowClientCert,
		Usage:       "Allow mTLS client certificate authentication",
		Description: "Allow Client Cert",
	})
	jwksCacheTTLSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "jwks-cache-ttl",
		Shorthand:   "",
		Value:       defaultJWKSCacheTTL,
		Usage:       "JWKS cache TTL in seconds",
		Description: "JWKS Cache TTL",
	})
	tokenCacheTTLSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "token-cache-ttl",
		Shorthand:   "",
		Value:       defaultTokenCacheTTL,
		Usage:       "Validated token cache TTL in seconds",
		Description: "Token Cache TTL",
	})
	enableTokenCachingSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityRSServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "enable-token-caching",
		Shorthand:   "",
		Value:       defaultEnableTokenCaching,
		Usage:       "Enable token validation result caching",
		Description: "Enable Token Caching",
	})
)

// ParseWithFlagSet parses command line arguments using provided FlagSet and returns identity-rs settings.
// This enables test isolation by allowing each test to use its own FlagSet.
func ParseWithFlagSet(fs *pflag.FlagSet, args []string, exitIfHelp bool) (*IdentityRSServerSettings, error) {
	// Register identity-rs specific flags on the provided FlagSet BEFORE parsing.
	fs.StringP(rsAuthzServerURLSetting.Name, rsAuthzServerURLSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(rsAuthzServerURLSetting), rsAuthzServerURLSetting.Description)
	fs.StringP(jwksEndpointSetting.Name, jwksEndpointSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(jwksEndpointSetting), jwksEndpointSetting.Description)
	fs.StringP(introspectionURLSetting.Name, introspectionURLSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(introspectionURLSetting), introspectionURLSetting.Description)
	fs.BoolP(allowBearerTokenSetting.Name, allowBearerTokenSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsBoolSetting(allowBearerTokenSetting), allowBearerTokenSetting.Description)
	fs.BoolP(allowClientCertSetting.Name, allowClientCertSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsBoolSetting(allowClientCertSetting), allowClientCertSetting.Description)
	fs.IntP(jwksCacheTTLSetting.Name, jwksCacheTTLSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsIntSetting(jwksCacheTTLSetting), jwksCacheTTLSetting.Description)
	fs.IntP(tokenCacheTTLSetting.Name, tokenCacheTTLSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsIntSetting(tokenCacheTTLSetting), tokenCacheTTLSetting.Description)
	fs.BoolP(enableTokenCachingSetting.Name, enableTokenCachingSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsBoolSetting(enableTokenCachingSetting), enableTokenCachingSetting.Description)

	// Parse base template settings using the same FlagSet.
	baseSettings, err := cryptoutilAppsFrameworkServiceConfig.ParseWithFlagSet(fs, args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	authzServerURL, _ := fs.GetString(rsAuthzServerURLSetting.Name)
	jwksEndpoint, _ := fs.GetString(jwksEndpointSetting.Name)
	introspectionURL, _ := fs.GetString(introspectionURLSetting.Name)
	allowBearerToken, _ := fs.GetBool(allowBearerTokenSetting.Name)
	allowClientCert, _ := fs.GetBool(allowClientCertSetting.Name)
	jwksCacheTTL, _ := fs.GetInt(jwksCacheTTLSetting.Name)
	tokenCacheTTL, _ := fs.GetInt(tokenCacheTTLSetting.Name)
	enableTokenCaching, _ := fs.GetBool(enableTokenCachingSetting.Name)

	settings := &IdentityRSServerSettings{
		ServiceFrameworkServerSettings: baseSettings,
		AuthzServerURL:                 authzServerURL,
		JWKSEndpoint:                   jwksEndpoint,
		IntrospectionURL:               introspectionURL,
		RequiredScopes:                 defaultRequiredScopes,
		RequiredAudiences:              defaultRequiredAudiences,
		AllowBearerToken:               allowBearerToken,
		AllowClientCert:                allowClientCert,
		JWKSCacheTTL:                   jwksCacheTTL,
		TokenCacheTTL:                  tokenCacheTTL,
		EnableTokenCaching:             enableTokenCaching,
	}

	if !fs.Changed("bind-public-port") {
		settings.BindPublicPort = cryptoutilSharedMagic.IdentityRSServicePort
	}

	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityRS

	if err := validateIdentityRSSettings(settings); err != nil {
		return nil, fmt.Errorf("identity-rs settings validation failed: %w", err)
	}

	logIdentityRSSettings(settings)

	return settings, nil
}

// Parse parses command-line arguments and returns the identity-rs server settings.
func Parse(args []string, exitIfHelp bool) (*IdentityRSServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, args, exitIfHelp)
}

// validateIdentityRSSettings validates identity-rs specific configuration.
func validateIdentityRSSettings(s *IdentityRSServerSettings) error {
	var validationErrors []string

	// Validate AuthZ server URL format if specified.
	if s.AuthzServerURL == "" && !s.DevMode {
		validationErrors = append(validationErrors, "authz-server-url is required in production mode")
	} else if s.AuthzServerURL != "" && !strings.HasPrefix(s.AuthzServerURL, "http://") && !strings.HasPrefix(s.AuthzServerURL, "https://") {
		validationErrors = append(validationErrors, "authz-server-url must start with http:// or https://")
	}

	// Validate at least one auth method is enabled.
	if !s.AllowBearerToken && !s.AllowClientCert {
		validationErrors = append(validationErrors, "at least one authentication method must be enabled (bearer-token or client-cert)")
	}

	// Validate cache TTLs.
	if s.JWKSCacheTTL < 0 {
		validationErrors = append(validationErrors, "jwks-cache-ttl must be non-negative")
	}

	if s.TokenCacheTTL < 0 {
		validationErrors = append(validationErrors, "token-cache-ttl must be non-negative")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logIdentityRSSettings logs identity-rs specific configuration to stderr.
func logIdentityRSSettings(s *IdentityRSServerSettings) {
	fmt.Fprintf(os.Stderr, "Identity-RS Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  AuthZ Server URL: %s\n", s.AuthzServerURL)
	fmt.Fprintf(os.Stderr, "  JWKS Endpoint: %s\n", s.JWKSEndpoint)
	fmt.Fprintf(os.Stderr, "  Introspection URL: %s\n", maskEmpty(s.IntrospectionURL, "(not configured)"))
	fmt.Fprintf(os.Stderr, "  Required Scopes: %v\n", s.RequiredScopes)
	fmt.Fprintf(os.Stderr, "  Required Audiences: %v\n", s.RequiredAudiences)
	fmt.Fprintf(os.Stderr, "  Allow Bearer Token: %t\n", s.AllowBearerToken)
	fmt.Fprintf(os.Stderr, "  Allow Client Cert: %t\n", s.AllowClientCert)
	fmt.Fprintf(os.Stderr, "  JWKS Cache TTL: %ds\n", s.JWKSCacheTTL)
	fmt.Fprintf(os.Stderr, "  Token Cache TTL: %ds\n", s.TokenCacheTTL)
	fmt.Fprintf(os.Stderr, "  Token Caching Enabled: %t\n", s.EnableTokenCaching)
}

// maskEmpty returns a default value if the string is empty.
func maskEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

// NewTestConfig creates an IdentityRSServerSettings instance for testing without calling Parse().
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
// Returns directly populated IdentityRSServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *IdentityRSServerSettings {
	// Get base template config.
	baseConfig := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-rs specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityRS

	return &IdentityRSServerSettings{
		ServiceFrameworkServerSettings: baseConfig,
		AuthzServerURL:                 defaultRSAuthzServerURL,
		JWKSEndpoint:                   defaultJWKSEndpoint,
		IntrospectionURL:               defaultIntrospectionURL,
		RequiredScopes:                 defaultRequiredScopes,
		RequiredAudiences:              defaultRequiredAudiences,
		AllowBearerToken:               defaultAllowBearerToken,
		AllowClientCert:                defaultAllowClientCert,
		JWKSCacheTTL:                   defaultJWKSCacheTTL,
		TokenCacheTTL:                  defaultTokenCacheTTL,
		EnableTokenCaching:             defaultEnableTokenCaching,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentityRSServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
