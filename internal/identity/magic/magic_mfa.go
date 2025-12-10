// Copyright (c) 2025 Iwan van der Kleijn
// SPDX-License-Identifier: MIT

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
	DefaultRecoveryCodeLength = 16                                     // 16 characters per code.
	DefaultRecoveryCodeCount  = 10                                     // 10 codes per batch.
	RecoveryCodeCharset       = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"     // Exclude ambiguous chars (0/O, 1/I/L).

	// Recovery Code Lifecycle.
	DefaultRecoveryCodeLifetime = 90 * 24 * time.Hour // 90 days.
)
