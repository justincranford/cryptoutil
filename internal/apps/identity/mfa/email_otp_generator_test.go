// Copyright (c) 2025 Justin Cranford

package mfa_test

import (
	"testing"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"

	"github.com/stretchr/testify/require"
)

func TestGenerateEmailOTP_Format(t *testing.T) {
	t.Parallel()

	otp, err := cryptoutilIdentityMfa.GenerateEmailOTP()
	require.NoError(t, err)
	require.Len(t, otp, cryptoutilIdentityMagic.DefaultEmailOTPLength, "OTP should be 6 digits")
	require.Regexp(t, `^\d{6}$`, otp, "OTP should be 6 numeric digits")
}

func TestGenerateEmailOTP_Uniqueness(t *testing.T) {
	t.Parallel()

	const samples = 1000

	seen := make(map[string]bool)

	for i := 0; i < samples; i++ {
		otp, err := cryptoutilIdentityMfa.GenerateEmailOTP()
		require.NoError(t, err)

		seen[otp] = true
	}

	// With 6-digit OTPs (1,000,000 possibilities), 1000 samples should have >900 unique values.
	require.Greater(t, len(seen), 900, "Should generate mostly unique OTPs")
}

func TestGenerateEmailOTP_AllNumeric(t *testing.T) {
	t.Parallel()

	const samples = 100

	for i := 0; i < samples; i++ {
		otp, err := cryptoutilIdentityMfa.GenerateEmailOTP()
		require.NoError(t, err)

		// Verify all characters are digits.
		for _, char := range otp {
			require.True(t, char >= '0' && char <= '9', "All characters should be digits (0-9)")
		}
	}
}
