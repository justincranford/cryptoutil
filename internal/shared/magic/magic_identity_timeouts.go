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
	DefaultDeviceCodeLifetime   = 1800 * time.Second  // Default device code lifetime (30 minutes) (RFC 8628).
	DefaultPARLifetime          = 90 * time.Second    // Default pushed authorization request lifetime (90 seconds) (RFC 9126).

	// Token expiry in seconds for OAuth 2.1 responses.
	AccessTokenExpirySeconds  = 3600  // Access token expiry in seconds (1 hour).
	RefreshTokenExpirySeconds = 86400 // Refresh token expiry in seconds (24 hours).
	IDTokenExpirySeconds      = 3600  // ID token expiry in seconds (1 hour).

	// Token cleanup interval.
	DefaultTokenCleanupInterval = 1 * time.Hour // Run cleanup every hour.
)

// Session lifetimes.
const (
	DefaultSessionLifetime = 3600 * time.Second // Default session lifetime (1 hour).
	DefaultIdleTimeout     = 900 * time.Second  // Default idle timeout (15 minutes).
)

// Device Authorization Grant polling (RFC 8628).
const (
	DefaultPollingInterval = 5 * time.Second // Minimum polling interval for device code grant.
)

// Logout timeouts.
const (
	BackChannelLogoutTimeout = 30 * time.Second // Timeout for back-channel logout HTTP requests.
)

// Server timeouts.
const (
	DefaultReadTimeout       = 30 * time.Second  // Default HTTP read timeout.
	DefaultWriteTimeout      = 30 * time.Second  // Default HTTP write timeout.
	DefaultIdleServerTimeout = 120 * time.Second // Default HTTP idle timeout.
)

// Demo/testing timeouts.
const (
	DemoStartupDelay  = 500 * time.Millisecond // Delay for server startup in demo.
	DemoTimeout       = 30 * time.Second       // Overall demo timeout.
	DemoRequestDelay  = 10 * time.Second       // Delay for request operations in demo.
	DemoAdminPort     = 9090                   // Admin port for demo server.
	DemoServerPort    = 8080                   // Server port for demo.
	DemoMinTokenChars = 20                     // Minimum token characters to display.
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
	DefaultDeviceCodeLength   = 32 // Default device code length (RFC 8628).
	DefaultUserCodeLength     = 8  // Default user code length (RFC 8628).
	DefaultRequestURILength   = 32 // Default request_uri identifier length (RFC 9126).
	AES256KeySize             = 32 // AES-256 key size in bytes.
	JWSPartCount              = 3  // JWT JWS part count (header.payload.signature).
	ByteShift                 = 8  // Bit shift for byte operations.

	// RSA key sizes (bits).
	RSA2048KeySize = 2048 // RSA-2048 key size (minimum for FIPS 140-3).
	RSA3072KeySize = 3072 // RSA-3072 key size.
	RSA4096KeySize = 4096 // RSA-4096 key size.

	// HMAC key sizes (bytes).
	HMACSHA256KeySize = 32 // HMAC-SHA256 key size (32 bytes = 256 bits).
	HMACSHA384KeySize = 48 // HMAC-SHA384 key size (48 bytes = 384 bits).
	HMACSHA512KeySize = 64 // HMAC-SHA512 key size (64 bytes = 512 bits).
)

// Password hashing.
const (
	// REMOVED: DefaultBcryptCost = 12 (non-FIPS). Use PBKDF2-HMAC-SHA256 (see internal/common/magic/magic_crypto.go: PBKDF2DefaultIterations).
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

// Push notification configuration.
const (
	DefaultPushNotificationTimeout     = 120 * time.Second // Default push notification timeout (2 minutes).
	DefaultPushNotificationTokenLength = 32                // Default push notification approval token length.
)

// Phone call OTP configuration.
const (
	DefaultPhoneCallOTPTimeout = 120 * time.Second // Default phone call OTP timeout (2 minutes).
	DefaultPhoneCallOTPRetries = 2                 // Default phone call OTP retry attempts.
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
