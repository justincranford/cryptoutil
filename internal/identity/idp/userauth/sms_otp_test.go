// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/idp/userauth"
)

func TestDefaultOTPGenerator_GenerateOTP(t *testing.T) {
	t.Parallel()

	generator := &userauth.DefaultOTPGenerator{}

	otp, err := generator.GenerateOTP(6)
	require.NoError(t, err, "GenerateOTP should succeed")
	require.Len(t, otp, 6, "OTP should be 6 digits")

	// Verify all characters are digits.
	for _, c := range otp {
		require.True(t, c >= '0' && c <= '9', "OTP should contain only digits: %s", otp)
	}
}

func TestDefaultOTPGenerator_GenerateOTPInvalidLength(t *testing.T) {
	t.Parallel()

	generator := &userauth.DefaultOTPGenerator{}

	_, err := generator.GenerateOTP(0)
	require.Error(t, err, "GenerateOTP should fail with zero length")

	_, err = generator.GenerateOTP(-1)
	require.Error(t, err, "GenerateOTP should fail with negative length")
}

func TestDefaultOTPGenerator_GenerateOTPUniqueness(t *testing.T) {
	t.Parallel()

	generator := &userauth.DefaultOTPGenerator{}
	otps := make(map[string]bool)

	// Generate multiple OTPs - with 6 digits, duplicates are statistically very unlikely.
	for range 20 {
		otp, err := generator.GenerateOTP(8) // Use 8 digits to reduce collision chance.
		require.NoError(t, err, "GenerateOTP should succeed")

		otps[otp] = true
	}

	// Expect at least 18 unique values (allowing for some statistical collision).
	require.GreaterOrEqual(t, len(otps), 18, "Most OTPs should be unique")
}

func TestDefaultOTPGenerator_GenerateSecureToken(t *testing.T) {
	t.Parallel()

	generator := &userauth.DefaultOTPGenerator{}

	token, err := generator.GenerateSecureToken(16)
	require.NoError(t, err, "GenerateSecureToken should succeed")
	require.Len(t, token, 32, "Token should be double the byte length (hex encoded)")

	// Verify all characters are hex.
	for _, c := range token {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
		require.True(t, isHex, "Token should be valid hex: %s", token)
	}
}

func TestDefaultOTPGenerator_GenerateSecureTokenInvalidLength(t *testing.T) {
	t.Parallel()

	generator := &userauth.DefaultOTPGenerator{}

	_, err := generator.GenerateSecureToken(0)
	require.Error(t, err, "GenerateSecureToken should fail with zero length")

	_, err = generator.GenerateSecureToken(-1)
	require.Error(t, err, "GenerateSecureToken should fail with negative length")
}

func TestDefaultOTPGenerator_GenerateSecureTokenUniqueness(t *testing.T) {
	t.Parallel()

	generator := &userauth.DefaultOTPGenerator{}
	tokens := make(map[string]bool)

	// Generate multiple tokens and ensure they're unique.
	for range 10 {
		token, err := generator.GenerateSecureToken(16)
		require.NoError(t, err, "GenerateSecureToken should succeed")
		require.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
	}
}

func TestSMSOTPAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	auth := userauth.NewSMSOTPAuthenticator(nil, nil, nil, nil, nil)
	require.NotNil(t, auth, "NewSMSOTPAuthenticator should return non-nil authenticator")
}

func TestSMSOTPAuthenticator_Method(t *testing.T) {
	t.Parallel()

	auth := userauth.NewSMSOTPAuthenticator(nil, nil, nil, nil, nil)
	require.Equal(t, "sms_otp", auth.Method(), "Method should return 'sms_otp'")
}
