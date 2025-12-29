// Copyright (c) 2025 Justin Cranford
//
//

package server

const (
	// Default realm password constraints.
	DefaultPasswordMinLength        = 12
	DefaultPasswordMinUniqueChars   = 8
	DefaultPasswordMaxRepeatedChars = 3

	// Default realm session constraints (in seconds).
	DefaultSessionTimeout      = 3600  // 1 hour.
	DefaultSessionAbsoluteMax  = 86400 // 24 hours.

	// Default realm rate limits (per minute).
	DefaultLoginRateLimit   = 5
	DefaultMessageRateLimit = 10

	// Enterprise realm password constraints.
	EnterprisePasswordMinLength        = 16
	EnterprisePasswordMinUniqueChars   = 12
	EnterprisePasswordMaxRepeatedChars = 2

	// Enterprise realm session constraints (in seconds).
	EnterpriseSessionTimeout     = 1800  // 30 minutes.
	EnterpriseSessionAbsoluteMax = 28800 // 8 hours.

	// Enterprise realm rate limits (per minute).
	EnterpriseLoginRateLimit   = 3
	EnterpriseMessageRateLimit = 5
)

// RealmConfig holds realm-specific validation and security configuration.
// Enterprise deployments can configure different realms with varying password complexity,
// session timeout, and MFA requirements.
type RealmConfig struct {
	// Password validation rules.
	PasswordMinLength        int  `mapstructure:"password_min_length" yaml:"password_min_length"`                 // Minimum password length (default: 12).
	PasswordRequireUppercase bool `mapstructure:"password_require_uppercase" yaml:"password_require_uppercase"`   // Require uppercase characters (default: true).
	PasswordRequireLowercase bool `mapstructure:"password_require_lowercase" yaml:"password_require_lowercase"`   // Require lowercase characters (default: true).
	PasswordRequireDigits    bool `mapstructure:"password_require_digits" yaml:"password_require_digits"`         // Require numeric digits (default: true).
	PasswordRequireSpecial   bool `mapstructure:"password_require_special" yaml:"password_require_special"`       // Require special characters (default: true).
	PasswordMinUniqueChars   int  `mapstructure:"password_min_unique_chars" yaml:"password_min_unique_chars"`     // Minimum unique characters (default: 8).
	PasswordMaxRepeatedChars int  `mapstructure:"password_max_repeated_chars" yaml:"password_max_repeated_chars"` // Maximum consecutive repeated characters (default: 3).

	// Session configuration.
	SessionTimeout        int  `mapstructure:"session_timeout" yaml:"session_timeout"`                 // Session timeout in seconds (default: 3600).
	SessionAbsoluteMax    int  `mapstructure:"session_absolute_max" yaml:"session_absolute_max"`       // Absolute maximum session duration regardless of activity (default: 86400).
	SessionRefreshEnabled bool `mapstructure:"session_refresh_enabled" yaml:"session_refresh_enabled"` // Enable session refresh on activity (default: true).

	// Multi-factor authentication.
	MFARequired bool `mapstructure:"mfa_required" yaml:"mfa_required"` // Require MFA for all users (default: false).
	MFAMethods  []string `mapstructure:"mfa_methods" yaml:"mfa_methods"` // Allowed MFA methods (e.g., TOTP, WebAuthn, SMS) (default: empty).

	// Rate limiting overrides (per realm).
	LoginRateLimit   int `mapstructure:"login_rate_limit" yaml:"login_rate_limit"`     // Login attempts per minute (default: 5).
	MessageRateLimit int `mapstructure:"message_rate_limit" yaml:"message_rate_limit"` // Messages sent per minute (default: 10).
}

// DefaultRealm returns the default realm configuration.
// Used when no specific realm is configured or as fallback.
func DefaultRealm() *RealmConfig {
	return &RealmConfig{
		PasswordMinLength:        DefaultPasswordMinLength,
		PasswordRequireUppercase: true,
		PasswordRequireLowercase: true,
		PasswordRequireDigits:    true,
		PasswordRequireSpecial:   true,
		PasswordMinUniqueChars:   DefaultPasswordMinUniqueChars,
		PasswordMaxRepeatedChars: DefaultPasswordMaxRepeatedChars,
		SessionTimeout:           DefaultSessionTimeout,
		SessionAbsoluteMax:       DefaultSessionAbsoluteMax,
		SessionRefreshEnabled:    true,
		MFARequired:              false,
		MFAMethods:               []string{},
		LoginRateLimit:           DefaultLoginRateLimit,
		MessageRateLimit:         DefaultMessageRateLimit,
	}
}

// EnterpriseRealm returns a more restrictive realm configuration for enterprise deployments.
func EnterpriseRealm() *RealmConfig {
	return &RealmConfig{
		PasswordMinLength:        EnterprisePasswordMinLength,
		PasswordRequireUppercase: true,
		PasswordRequireLowercase: true,
		PasswordRequireDigits:    true,
		PasswordRequireSpecial:   true,
		PasswordMinUniqueChars:   EnterprisePasswordMinUniqueChars,
		PasswordMaxRepeatedChars: EnterprisePasswordMaxRepeatedChars,
		SessionTimeout:           EnterpriseSessionTimeout,
		SessionAbsoluteMax:       EnterpriseSessionAbsoluteMax,
		SessionRefreshEnabled:    true,
		MFARequired:              true,
		MFAMethods:               []string{"totp", "webauthn"},
		LoginRateLimit:           EnterpriseLoginRateLimit,
		MessageRateLimit:         EnterpriseMessageRateLimit,
	}
}
