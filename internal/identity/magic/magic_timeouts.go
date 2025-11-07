package magic

import "time"

// Token lifetimes.
const (
	DefaultAccessTokenLifetime  = 3600 * time.Second  // Default access token lifetime (1 hour).
	DefaultRefreshTokenLifetime = 86400 * time.Second // Default refresh token lifetime (24 hours).
	DefaultIDTokenLifetime      = 3600 * time.Second  // Default ID token lifetime (1 hour).
	DefaultCodeLifetime         = 300 * time.Second   // Default authorization code lifetime (5 minutes).
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
)

// Password hashing.
const (
	DefaultBcryptCost = 12  // Default bcrypt cost factor.
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
