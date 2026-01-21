// Copyright (c) 2025 Justin Cranford

// Package magic defines MFA-related constants for the identity service.
package magic

import "time"

// MFA Factor Type constants.
const (
	// MFATypeTOTP is the TOTP (time-based one-time password) MFA factor type.
	MFATypeTOTP = "totp"
	// MFATypeRecoveryCode is the recovery code MFA factor type.
	MFATypeRecoveryCode = "recovery_code"
	// MFATypeWebAuthn is the WebAuthn MFA factor type.
	MFATypeWebAuthn = "webauthn"
	// MFATypeEmail is the email-based MFA factor type.
	MFATypeEmail = "email"
	// MFATypeSMS is the SMS-based MFA factor type.
	MFATypeSMS = "sms"
)

// Recovery Code Generation constants.
const (
	// DefaultRecoveryCodeLength is 16 characters per code.
	DefaultRecoveryCodeLength = 16
	// DefaultRecoveryCodeCount is 10 codes per batch.
	DefaultRecoveryCodeCount = 10
	// RecoveryCodeCharset excludes ambiguous chars (0/O, 1/I/L).
	RecoveryCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

// Recovery Code Lifecycle constants.
const (
	// DefaultRecoveryCodeLifetime is 90 days.
	DefaultRecoveryCodeLifetime = 90 * 24 * time.Hour
)

// Email OTP Generation constants.
const (
	// DefaultEmailOTPLength is 6-digit numeric OTP.
	DefaultEmailOTPLength = 6
	// DefaultEmailOTPLifetime is 10 minutes.
	DefaultEmailOTPLifetime = 10 * time.Minute
	// EmailOTPCharset is numeric digits only.
	EmailOTPCharset = "0123456789"
)

// Email OTP Rate Limiting constants.
const (
	// DefaultEmailOTPRateLimit is 5 OTP requests per window.
	DefaultEmailOTPRateLimit = 5
	// DefaultEmailOTPRateLimitWindow is 1-hour rate limit window.
	DefaultEmailOTPRateLimitWindow = 1 * time.Hour
)
