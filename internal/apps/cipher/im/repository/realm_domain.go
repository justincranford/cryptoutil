// Copyright (c) 2025 Justin Cranford

package repository

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// RealmType represents the authentication method type for a realm.
type RealmType string

const (
	// RealmTypeJWESessionCookie uses JWE (encrypted JWT) session cookies for browser clients.
	RealmTypeJWESessionCookie RealmType = "jwe-session-cookie"

	// RealmTypeJWSSessionCookie uses JWS (signed JWT) session cookies for browser clients.
	RealmTypeJWSSessionCookie RealmType = "jws-session-cookie"

	// RealmTypeOpaqueSessionCookie uses opaque (database-backed) session cookies for browser clients.
	RealmTypeOpaqueSessionCookie RealmType = "opaque-session-cookie"

	// RealmTypeBasicAuth uses HTTP Basic authentication (username/password for browser, client ID/secret for headless).
	RealmTypeBasicAuth RealmType = "basic-username-password"

	// RealmTypeBearerToken uses Bearer token authentication (API keys) for both browser and headless clients.
	RealmTypeBearerToken RealmType = "bearer-api-token"

	// RealmTypeHTTPSClientCert uses HTTPS client certificate (mTLS) authentication for both browser and headless clients.
	RealmTypeHTTPSClientCert RealmType = "https-client-cert"
)

// Realm represents an authentication realm configuration.
// Each realm defines an authentication method, security policies, and session settings.
// Supports both config file and database storage (Config > DB priority).
type Realm struct {
	ID       googleUuid.UUID `gorm:"type:text;primaryKey"`
	RealmID  googleUuid.UUID `gorm:"type:text;not null;uniqueIndex:idx_realms_realm_id"` // Unique realm identifier.
	Type     RealmType       `gorm:"type:text;not null;index"`                           // Realm type (authentication method).
	Name     string          `gorm:"type:text;not null;uniqueIndex:idx_realms_name"`     // Human-readable realm name.
	Config   string          `gorm:"type:text"`                                          // JSON configuration for realm-specific settings.
	Active   bool            `gorm:"not null;default:true;index"`                        // Active/inactive realm.
	Source   string          `gorm:"type:text;not null;default:db"`                      // Source: config (YAML) or db (database).
	Priority int             `gorm:"not null;default:0;index"`                           // Priority for realm selection (higher = preferred).

	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName returns the database table name for Realm.
func (Realm) TableName() string {
	return "cipher_im_realms"
}

// RealmConfig holds realm-specific validation and security configuration.
// Each realm can have different password complexity, session timeout, and MFA requirements.
type RealmConfig struct {
	// Password validation rules (for Basic Auth realms).
	PasswordMinLength        int  `json:"password_min_length" yaml:"password_min_length"`                 // Minimum password length (default: 12).
	PasswordRequireUppercase bool `json:"password_require_uppercase" yaml:"password_require_uppercase"`   // Require uppercase characters (default: true).
	PasswordRequireLowercase bool `json:"password_require_lowercase" yaml:"password_require_lowercase"`   // Require lowercase characters (default: true).
	PasswordRequireDigits    bool `json:"password_require_digits" yaml:"password_require_digits"`         // Require numeric digits (default: true).
	PasswordRequireSpecial   bool `json:"password_require_special" yaml:"password_require_special"`       // Require special characters (default: true).
	PasswordMinUniqueChars   int  `json:"password_min_unique_chars" yaml:"password_min_unique_chars"`     // Minimum unique characters (default: 8).
	PasswordMaxRepeatedChars int  `json:"password_max_repeated_chars" yaml:"password_max_repeated_chars"` // Maximum consecutive repeated characters (default: 3).

	// Session configuration (for session-based realms: JWE, JWS, OPAQUE).
	SessionTimeout        int  `json:"session_timeout" yaml:"session_timeout"`                 // Session timeout in seconds (default: 3600).
	SessionAbsoluteMax    int  `json:"session_absolute_max" yaml:"session_absolute_max"`       // Absolute maximum session duration regardless of activity (default: 86400).
	SessionRefreshEnabled bool `json:"session_refresh_enabled" yaml:"session_refresh_enabled"` // Enable session refresh on activity (default: true).

	// Token configuration (for token-based realms: Bearer, HTTPS Client Cert).
	TokenExpiry int `json:"token_expiry" yaml:"token_expiry"` // Token expiry in seconds (default: 3600 for Bearer tokens).

	// Multi-factor authentication (applicable to all realms).
	MFARequired bool     `json:"mfa_required" yaml:"mfa_required"` // Require MFA for all users (default: false).
	MFAMethods  []string `json:"mfa_methods" yaml:"mfa_methods"`   // Allowed MFA methods (e.g., TOTP, WebAuthn, SMS) (default: empty).

	// Rate limiting overrides (per realm).
	LoginRateLimit   int `json:"login_rate_limit" yaml:"login_rate_limit"`     // Login attempts per minute (default: 5).
	MessageRateLimit int `json:"message_rate_limit" yaml:"message_rate_limit"` // Messages sent per minute (default: 10).

	// TLS configuration (for HTTPS Client Cert realm).
	RequireClientCert bool     `json:"require_client_cert" yaml:"require_client_cert"` // Require client certificate (default: true for HTTPS Client Cert realm).
	TrustedCAs        []string `json:"trusted_cas" yaml:"trusted_cas"`                 // List of trusted CA certificates (PEM format).
}
