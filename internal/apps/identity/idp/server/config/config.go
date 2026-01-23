// Copyright (c) 2025 Justin Cranford

// Package config provides identity-idp server configuration settings.
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

// IdentityIDPServerSettings contains identity-idp specific configuration.
type IdentityIDPServerSettings struct {
	*cryptoutilTemplateConfig.ServiceTemplateServerSettings

	// IdP (Identity Provider) settings.
	AuthzServerURL string // URL of the OAuth 2.1 authorization server to integrate with.
	LoginPagePath  string // Path to custom login page template.
	ConsentPagePath string // Path to custom consent page template.

	// MFA enrollment settings.
	EnableMFAEnrollment bool // Enable MFA enrollment during login.
	RequireMFA          bool // Require MFA for all logins.
	MFAMethods          []string // Supported MFA methods (totp, webauthn, push).

	// Session settings.
	LoginSessionTimeout   int // Login session timeout in seconds.
	ConsentSessionTimeout int // Consent session timeout in seconds.
}

// Identity-IDP specific default values.
const (
	defaultIDPAuthzServerURL      = "https://localhost:18000" // Default authorization server URL.
	defaultLoginPagePath          = ""                        // Use built-in login page.
	defaultConsentPagePath        = ""                        // Use built-in consent page.
	defaultEnableMFAEnrollment    = false                     // Disable MFA enrollment by default.
	defaultRequireMFA             = false                     // Don't require MFA by default.
	defaultLoginSessionTimeout    = 300                       // 5 minute login session timeout.
	defaultConsentSessionTimeout  = 300                       // 5 minute consent session timeout.
)

var defaultMFAMethods = []string{"totp"} // Default MFA methods.

var allIdentityIDPServerRegisteredSettings []*cryptoutilTemplateConfig.Setting //nolint:gochecknoglobals

// Identity-IDP specific Setting objects for parameter attributes.
var (
	idpAuthzServerURLSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "authz-server-url",
		Shorthand:   "",
		Value:       defaultIDPAuthzServerURL,
		Usage:       "URL of the OAuth 2.1 authorization server to integrate with",
		Description: "AuthZ Server URL",
	})
	loginPagePathSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "login-page-path",
		Shorthand:   "",
		Value:       defaultLoginPagePath,
		Usage:       "Path to custom login page template (empty for built-in)",
		Description: "Login Page Path",
	})
	consentPagePathSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "consent-page-path",
		Shorthand:   "",
		Value:       defaultConsentPagePath,
		Usage:       "Path to custom consent page template (empty for built-in)",
		Description: "Consent Page Path",
	})
	enableMFAEnrollmentSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "enable-mfa-enrollment",
		Shorthand:   "",
		Value:       defaultEnableMFAEnrollment,
		Usage:       "Enable MFA enrollment during login",
		Description: "Enable MFA Enrollment",
	})
	requireMFASetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "require-mfa",
		Shorthand:   "",
		Value:       defaultRequireMFA,
		Usage:       "Require MFA for all logins",
		Description: "Require MFA",
	})
	loginSessionTimeoutSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "login-session-timeout",
		Shorthand:   "",
		Value:       defaultLoginSessionTimeout,
		Usage:       "Login session timeout in seconds",
		Description: "Login Session Timeout",
	})
	consentSessionTimeoutSetting = cryptoutilTemplateConfig.SetEnvAndRegisterSetting(allIdentityIDPServerRegisteredSettings, &cryptoutilTemplateConfig.Setting{
		Name:        "consent-session-timeout",
		Shorthand:   "",
		Value:       defaultConsentSessionTimeout,
		Usage:       "Consent session timeout in seconds",
		Description: "Consent Session Timeout",
	})
)

// Parse parses command line arguments and returns identity-idp settings.
func Parse(args []string, exitIfHelp bool) (*IdentityIDPServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilTemplateConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register identity-idp specific flags.
	pflag.StringP(idpAuthzServerURLSetting.Name, idpAuthzServerURLSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(idpAuthzServerURLSetting), idpAuthzServerURLSetting.Description)
	pflag.StringP(loginPagePathSetting.Name, loginPagePathSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(loginPagePathSetting), loginPagePathSetting.Description)
	pflag.StringP(consentPagePathSetting.Name, consentPagePathSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsStringSetting(consentPagePathSetting), consentPagePathSetting.Description)
	pflag.BoolP(enableMFAEnrollmentSetting.Name, enableMFAEnrollmentSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsBoolSetting(enableMFAEnrollmentSetting), enableMFAEnrollmentSetting.Description)
	pflag.BoolP(requireMFASetting.Name, requireMFASetting.Shorthand, cryptoutilTemplateConfig.RegisterAsBoolSetting(requireMFASetting), requireMFASetting.Description)
	pflag.IntP(loginSessionTimeoutSetting.Name, loginSessionTimeoutSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsIntSetting(loginSessionTimeoutSetting), loginSessionTimeoutSetting.Description)
	pflag.IntP(consentSessionTimeoutSetting.Name, consentSessionTimeoutSetting.Shorthand, cryptoutilTemplateConfig.RegisterAsIntSetting(consentSessionTimeoutSetting), consentSessionTimeoutSetting.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create identity-idp settings.
	settings := &IdentityIDPServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		AuthzServerURL:                viper.GetString(idpAuthzServerURLSetting.Name),
		LoginPagePath:                 viper.GetString(loginPagePathSetting.Name),
		ConsentPagePath:               viper.GetString(consentPagePathSetting.Name),
		EnableMFAEnrollment:           viper.GetBool(enableMFAEnrollmentSetting.Name),
		RequireMFA:                    viper.GetBool(requireMFASetting.Name),
		MFAMethods:                    defaultMFAMethods,
		LoginSessionTimeout:           viper.GetInt(loginSessionTimeoutSetting.Name),
		ConsentSessionTimeout:         viper.GetInt(consentSessionTimeoutSetting.Name),
	}

	// Override template defaults with identity-idp specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.IdentityIDPServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityIDP

	// Validate identity-idp specific settings.
	if err := validateIdentityIDPSettings(settings); err != nil {
		return nil, fmt.Errorf("identity-idp settings validation failed: %w", err)
	}

	// Log identity-idp specific settings.
	logIdentityIDPSettings(settings)

	return settings, nil
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
	baseConfig := cryptoutilTemplateConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-idp specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentityIDP

	return &IdentityIDPServerSettings{
		ServiceTemplateServerSettings: baseConfig,
		AuthzServerURL:                defaultIDPAuthzServerURL,
		LoginPagePath:                 defaultLoginPagePath,
		ConsentPagePath:               defaultConsentPagePath,
		EnableMFAEnrollment:           defaultEnableMFAEnrollment,
		RequireMFA:                    defaultRequireMFA,
		MFAMethods:                    defaultMFAMethods,
		LoginSessionTimeout:           defaultLoginSessionTimeout,
		ConsentSessionTimeout:         defaultConsentSessionTimeout,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentityIDPServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
