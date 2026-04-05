// Copyright (c) 2025 Justin Cranford

// Package config provides identity-idp server configuration settings.
package config

import (
	"fmt"
	"os"
	"strings"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
)

// IdentityIDPServerSettings contains identity-idp specific configuration.
type IdentityIDPServerSettings struct {
	*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings

	// IdP (Identity Provider) settings.
	AuthzServerURL  string // URL of the OAuth 2.1 authorization server to integrate with.
	LoginPagePath   string // Path to custom login page template.
	ConsentPagePath string // Path to custom consent page template.

	// MFA enrollment settings.
	EnableMFAEnrollment bool     // Enable MFA enrollment during login.
	RequireMFA          bool     // Require MFA for all logins.
	MFAMethods          []string // Supported MFA methods (totp, webauthn, push).

	// Session settings.
	LoginSessionTimeout   int // Login session timeout in seconds.
	ConsentSessionTimeout int // Consent session timeout in seconds.
}

// Identity-IDP specific default values.
const (
	defaultIDPAuthzServerURL     = "https://localhost:8200" // Default authorization server URL.
	defaultLoginPagePath         = ""                       // Use built-in login page.
	defaultConsentPagePath       = ""                       // Use built-in consent page.
	defaultEnableMFAEnrollment   = false                    // Disable MFA enrollment by default.
	defaultRequireMFA            = false                    // Don't require MFA by default.
	defaultLoginSessionTimeout   = 300                      // 5 minute login session timeout.
	defaultConsentSessionTimeout = 300                      // 5 minute consent session timeout.
)

var defaultMFAMethods = []string{cryptoutilSharedMagic.MFATypeTOTP} // Default MFA methods.

var allIdentityIDPServerRegisteredSettings []*cryptoutilAppsFrameworkServiceConfig.Setting //nolint:gochecknoglobals

// Identity-IDP specific Setting objects for parameter attributes.
var (
	idpAuthzServerURLSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "authz-server-url",
		Shorthand:   "",
		Value:       defaultIDPAuthzServerURL,
		Usage:       "URL of the OAuth 2.1 authorization server to integrate with",
		Description: "AuthZ Server URL",
	})
	loginPagePathSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "login-page-path",
		Shorthand:   "",
		Value:       defaultLoginPagePath,
		Usage:       "Path to custom login page template (empty for built-in)",
		Description: "Login Page Path",
	})
	consentPagePathSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "consent-page-path",
		Shorthand:   "",
		Value:       defaultConsentPagePath,
		Usage:       "Path to custom consent page template (empty for built-in)",
		Description: "Consent Page Path",
	})
	enableMFAEnrollmentSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "enable-mfa-enrollment",
		Shorthand:   "",
		Value:       defaultEnableMFAEnrollment,
		Usage:       "Enable MFA enrollment during login",
		Description: "Enable MFA Enrollment",
	})
	requireMFASetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "require-mfa",
		Shorthand:   "",
		Value:       defaultRequireMFA,
		Usage:       "Require MFA for all logins",
		Description: "Require MFA",
	})
	loginSessionTimeoutSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "login-session-timeout",
		Shorthand:   "",
		Value:       defaultLoginSessionTimeout,
		Usage:       "Login session timeout in seconds",
		Description: "Login Session Timeout",
	})
	consentSessionTimeoutSetting = cryptoutilAppsFrameworkServiceConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilAppsFrameworkServiceConfig.Setting{
		Name:        "consent-session-timeout",
		Shorthand:   "",
		Value:       defaultConsentSessionTimeout,
		Usage:       "Consent session timeout in seconds",
		Description: "Consent Session Timeout",
	})
)

// ParseWithFlagSet parses command line arguments using provided FlagSet and returns identity-idp settings.
// This enables test isolation by allowing each test to use its own FlagSet.
func ParseWithFlagSet(fs *pflag.FlagSet, args []string, exitIfHelp bool) (*IdentityIDPServerSettings, error) {
	// Register identity-idp specific flags on the provided FlagSet BEFORE parsing.
	// This must happen before calling template ParseWithFlagSet since it will call fs.Parse().
	fs.StringP(idpAuthzServerURLSetting.Name, idpAuthzServerURLSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(idpAuthzServerURLSetting), idpAuthzServerURLSetting.Description)
	fs.StringP(loginPagePathSetting.Name, loginPagePathSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(loginPagePathSetting), loginPagePathSetting.Description)
	fs.StringP(consentPagePathSetting.Name, consentPagePathSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsStringSetting(consentPagePathSetting), consentPagePathSetting.Description)
	fs.BoolP(enableMFAEnrollmentSetting.Name, enableMFAEnrollmentSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsBoolSetting(enableMFAEnrollmentSetting), enableMFAEnrollmentSetting.Description)
	fs.BoolP(requireMFASetting.Name, requireMFASetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsBoolSetting(requireMFASetting), requireMFASetting.Description)
	fs.IntP(loginSessionTimeoutSetting.Name, loginSessionTimeoutSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsIntSetting(loginSessionTimeoutSetting), loginSessionTimeoutSetting.Description)
	fs.IntP(consentSessionTimeoutSetting.Name, consentSessionTimeoutSetting.Shorthand, cryptoutilAppsFrameworkServiceConfig.RegisterAsIntSetting(consentSessionTimeoutSetting), consentSessionTimeoutSetting.Description)

	// Parse base template settings using the same FlagSet.
	// This will register template flags and call fs.Parse() + viper.BindPFlags().
	baseSettings, err := cryptoutilAppsFrameworkServiceConfig.ParseWithFlagSet(fs, args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Create identity-idp settings using values from the FlagSet (avoids global viper dependency).
	authzServerURL, _ := fs.GetString(idpAuthzServerURLSetting.Name)
	loginPagePath, _ := fs.GetString(loginPagePathSetting.Name)
	consentPagePath, _ := fs.GetString(consentPagePathSetting.Name)
	enableMFAEnrollment, _ := fs.GetBool(enableMFAEnrollmentSetting.Name)
	requireMFA, _ := fs.GetBool(requireMFASetting.Name)
	loginSessionTimeout, _ := fs.GetInt(loginSessionTimeoutSetting.Name)
	consentSessionTimeout, _ := fs.GetInt(consentSessionTimeoutSetting.Name)

	settings := &IdentityIDPServerSettings{
		ServiceFrameworkServerSettings: baseSettings,
		AuthzServerURL:                 authzServerURL,
		LoginPagePath:                  loginPagePath,
		ConsentPagePath:                consentPagePath,
		EnableMFAEnrollment:            enableMFAEnrollment,
		RequireMFA:                     requireMFA,
		MFAMethods:                     defaultMFAMethods,
		LoginSessionTimeout:            loginSessionTimeout,
		ConsentSessionTimeout:          consentSessionTimeout,
	}

	// Override template defaults with identity-idp specific values.
	// Only override public port if user didn't explicitly specify one via CLI flag.
	if !fs.Changed("bind-public-port") {
		settings.BindPublicPort = cryptoutilSharedMagic.IdentityIDPServicePort
	}

	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityIDP

	// Validate identity-idp specific settings.
	if err := validateIdentityIDPSettings(settings); err != nil {
		return nil, fmt.Errorf("identity-idp settings validation failed: %w", err)
	}

	// Log identity-idp specific settings.
	logIdentityIDPSettings(settings)

	return settings, nil
}

// Parse parses command-line arguments and returns the identity-idp server settings.
func Parse(args []string, exitIfHelp bool) (*IdentityIDPServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, args, exitIfHelp)
}

// validateIdentityIDPSettings validates identity-idp specific configuration.
func validateIdentityIDPSettings(s *IdentityIDPServerSettings) error {
	var validationErrors []string

	// Validate AuthZ server URL format if specified.
	if s.AuthzServerURL == "" && !s.DevMode {
		validationErrors = append(validationErrors, "authz-server-url is required in production mode")
	} else if s.AuthzServerURL != "" && !strings.HasPrefix(s.AuthzServerURL, "http://") && !strings.HasPrefix(s.AuthzServerURL, "https://") {
		validationErrors = append(validationErrors, "authz-server-url must start with http:// or https://")
	}

	// Validate session timeouts.
	if s.LoginSessionTimeout <= 0 {
		validationErrors = append(validationErrors, "login-session-timeout must be positive")
	}

	if s.ConsentSessionTimeout <= 0 {
		validationErrors = append(validationErrors, "consent-session-timeout must be positive")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logIdentityIDPSettings logs identity-idp specific configuration to stderr.
func logIdentityIDPSettings(s *IdentityIDPServerSettings) {
	fmt.Fprintf(os.Stderr, "Identity-IDP Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  AuthZ Server URL: %s\n", s.AuthzServerURL)
	fmt.Fprintf(os.Stderr, "  Login Page Path: %s\n", maskEmpty(s.LoginPagePath, "(built-in)"))
	fmt.Fprintf(os.Stderr, "  Consent Page Path: %s\n", maskEmpty(s.ConsentPagePath, "(built-in)"))
	fmt.Fprintf(os.Stderr, "  MFA Enrollment: %t\n", s.EnableMFAEnrollment)
	fmt.Fprintf(os.Stderr, "  Require MFA: %t\n", s.RequireMFA)
	fmt.Fprintf(os.Stderr, "  MFA Methods: %v\n", s.MFAMethods)
	fmt.Fprintf(os.Stderr, "  Login Session Timeout: %ds\n", s.LoginSessionTimeout)
	fmt.Fprintf(os.Stderr, "  Consent Session Timeout: %ds\n", s.ConsentSessionTimeout)
}

// maskEmpty returns a default value if the string is empty.
func maskEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

// NewTestConfig creates an IdentityIDPServerSettings instance for testing without calling Parse().
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
// Returns directly populated IdentityIDPServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *IdentityIDPServerSettings {
	// Get base template config.
	baseConfig := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-idp specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityIDP

	return &IdentityIDPServerSettings{
		ServiceFrameworkServerSettings: baseConfig,
		AuthzServerURL:                 defaultIDPAuthzServerURL,
		LoginPagePath:                  defaultLoginPagePath,
		ConsentPagePath:                defaultConsentPagePath,
		EnableMFAEnrollment:            defaultEnableMFAEnrollment,
		RequireMFA:                     defaultRequireMFA,
		MFAMethods:                     defaultMFAMethods,
		LoginSessionTimeout:            defaultLoginSessionTimeout,
		ConsentSessionTimeout:          defaultConsentSessionTimeout,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentityIDPServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
