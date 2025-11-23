// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// Token lifetimes.
const (
	DefaultAccessTokenLifetime  = 3600 * time.Second  // Default access token lifetime (1 hour).
	DefaultRefreshTokenLifetime = 86400 * time.Second // Default refresh token lifetime (24 hours).
	DefaultIDTokenLifetime      = 3600 * time.Second  // Default ID token lifetime (1 hour).
	DefaultCodeLifetime         = 300 * time.Second   // Default authorization code lifetime (5 minutes).

	// Token expiry in seconds for OAuth 2.1 responses.
	AccessTokenExpirySeconds  = 3600  // Access token expiry in seconds (1 hour).
	RefreshTokenExpirySeconds = 86400 // Refresh token expiry in seconds (24 hours).
	IDTokenExpirySeconds      = 3600  // ID token expiry in seconds (1 hour).
)

// Session lifetimes.
const (
	DefaultSessionLifetime = 3600 * time.Second // Default session lifetime (1 hour).
	DefaultIdleTimeout     = 900 * time.Second  // Default idle timeout (15 minutes).
)

// Server timeouts.
const (
	DefaultReadTimeout       = 30 * time.Second  // Default HTTP read timeout.
	DefaultWriteTimeout      = 30 * time.Second  // Default HTTP write timeout.
	DefaultIdleServerTimeout = 120 * time.Second // Default HTTP idle timeout.
)

// Rate limiting.
const (
	DefaultRateLimitRequests = 100              // Default rate limit requests per window.
	DefaultRateLimitWindow   = 60 * time.Second // Default rate limit window (1 minute).
)

// Database connection pool.
const (
	DefaultMaxOpenConns    = 25               // Default maximum open connections.
	DefaultMaxIdleConns    = 5                // Default maximum idle connections.
	DefaultConnMaxLifetime = 5 * time.Minute  // Default connection max lifetime.
	DefaultConnMaxIdleTime = 10 * time.Minute // Default connection max idle time.
)

// PKCE code challenge.
const (
	DefaultCodeChallengeLength = 43  // Default PKCE code challenge length (43-128 characters).
	MinCodeChallengeLength     = 43  // Minimum PKCE code challenge length.
	MaxCodeChallengeLength     = 128 // Maximum PKCE code challenge length.
)

// Token sizes.
const (
	DefaultStateLength        = 32 // Default state parameter length.
	DefaultNonceLength        = 32 // Default nonce parameter length.
	DefaultAuthCodeLength     = 32 // Default authorization code length.
	DefaultRefreshTokenLength = 64 // Default refresh token length.
	AES256KeySize             = 32 // AES-256 key size in bytes.
	JWSPartCount              = 3  // JWT JWS part count (header.payload.signature).
	ByteShift                 = 8  // Bit shift for byte operations.
)

// Password hashing.
const (
	// REMOVED: DefaultBcryptCost = 12  // bcrypt is NOT FIPS-140-3 approved. Use PBKDF2-HMAC-SHA256 (see internal/common/magic/magic_crypto.go: PBKDF2DefaultIterations).
	MinPasswordLength = 8   // Minimum password length.
	MaxPasswordLength = 128 // Maximum password length.
)

// TOTP/HOTP configuration.
const (
	DefaultTOTPDigits    = 6                // Default TOTP digits.
	DefaultTOTPPeriod    = 30 * time.Second // Default TOTP period (30 seconds).
	DefaultTOTPAlgorithm = "SHA1"           // Default TOTP algorithm.
	DefaultHOTPDigits    = 6                // Default HOTP digits.
)

// OTP configuration.
const (
	DefaultOTPLength   = 6                 // Default OTP length.
	DefaultOTPLifetime = 300 * time.Second // Default OTP lifetime (5 minutes).
	MaxOTPAttempts     = 3                 // Maximum OTP verification attempts.
	DefaultOTPLockout  = 15 * time.Minute  // Default OTP lockout duration (15 minutes).
	DecimalRadix       = 10                // Decimal radix for numeric OTP generation.
)

// Magic link configuration.
const (
	DefaultMagicLinkLifetime = 900 * time.Second // Default magic link lifetime (15 minutes).
	DefaultMagicLinkLength   = 64                // Default magic link token length.
)

// Retry configuration.
const (
	DefaultMaxRetries    = 3                // Default maximum retries.
	DefaultRetryDelay    = 1 * time.Second  // Default retry delay.
	DefaultRetryMaxDelay = 30 * time.Second // Default maximum retry delay.
)

// Cleanup intervals.
const (
	ChallengeCleanupInterval = 5 * time.Minute // Challenge cleanup interval for expired entries.
)

// Certificate validation.
const (
	DefaultCertificateMaxAgeDays = 365 // Default maximum certificate age in days (1 year).
	StrictCertificateMaxAgeDays  = 90  // Strict maximum certificate age in days (90 days).
)

// Key rotation configuration.
const (
	// Default policy (balanced security).
	DefaultKeyRotationInterval = 30 * 24 * time.Hour // Default key rotation interval (30 days).
	DefaultKeyGracePeriod      = 7 * 24 * time.Hour  // Default key grace period (7 days).
	DefaultMaxActiveKeys       = 3                   // Default maximum active keys.

	// Strict policy (production).
	StrictKeyRotationInterval = 7 * 24 * time.Hour // Strict key rotation interval (7 days).
	StrictKeyGracePeriod      = 24 * time.Hour     // Strict key grace period (1 day).
	StrictMaxActiveKeys       = 2                  // Strict maximum active keys.

	// Development policy (relaxed).
	DevelopmentKeyRotationInterval = 365 * 24 * time.Hour // Development key rotation interval (1 year).
	DevelopmentKeyGracePeriod      = 30 * 24 * time.Hour  // Development key grace period (30 days).
	DevelopmentMaxActiveKeys       = 5                    // Development maximum active keys.
)

// Authentication method identifiers.
const (
	AuthMethodTOTP             = "totp"              // TOTP authentication method.
	AuthMethodSMSOTP           = "sms_otp"           // SMS OTP authentication method.
	AuthMethodHardwareKey      = "hardware_key"      // Hardware key authentication method.
	AuthMethodBiometric        = "biometric"         // Biometric authentication method.
	AuthMethodUsernamePassword = "username_password" // Username/password authentication method.
)

// Certificate revocation checking timeouts.
const (
	DefaultOCSPTimeout       = 5 * time.Second  // Default OCSP request timeout.
	DefaultCRLTimeout        = 10 * time.Second // Default CRL download timeout.
	DefaultCRLCacheMaxAge    = 1 * time.Hour    // Default CRL cache max age.
	DefaultRevocationTimeout = 10 * time.Second // Default revocation check timeout (certificate validator).
)
