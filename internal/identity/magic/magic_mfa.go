// Copyright (c) 2025 Justin Cranford

// Package magic defines MFA-related constants for the identity service.
package magic

import "time"

const (
	// MFA Factor Types.
	MFATypeTOTP         = "totp"
	MFATypeRecoveryCode = "recovery_code"
	MFATypeWebAuthn     = "webauthn"
	MFATypeEmail        = "email"
	MFATypeSMS          = "sms"

	// Recovery Code Generation.
	DefaultRecoveryCodeLength = 16                                 // 16 characters per code.
	DefaultRecoveryCodeCount  = 10                                 // 10 codes per batch.
	RecoveryCodeCharset       = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude ambiguous chars (0/O, 1/I/L).

	// Recovery Code Lifecycle.
	DefaultRecoveryCodeLifetime = 90 * 24 * time.Hour // 90 days.

	// Email OTP Generation.
	DefaultEmailOTPLength   = 6                // 6-digit numeric OTP.
	DefaultEmailOTPLifetime = 10 * time.Minute // 10 minutes.
	EmailOTPCharset         = "0123456789"     // Numeric digits only.

	// Email OTP Rate Limiting.
	DefaultEmailOTPRateLimit       = 5             // 5 OTP requests per window.
	DefaultEmailOTPRateLimitWindow = 1 * time.Hour // 1-hour rate limit window.
)
