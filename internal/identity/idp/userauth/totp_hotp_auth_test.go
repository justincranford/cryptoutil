// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/idp/userauth"
)

func TestTOTPAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)
	require.NotNil(t, auth, "NewTOTPAuthenticator should return non-nil authenticator")
}

func TestTOTPAuthenticator_Method(t *testing.T) {
	t.Parallel()

	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)
	require.Equal(t, "totp", auth.Method(), "Method should return 'totp'")
}

func TestTOTPAuthenticator_GenerateSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)

	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")
	require.NotEmpty(t, secret, "Secret should not be empty")

	// Secret should be base32 encoded (uppercase letters A-Z and digits 2-7).
	for _, c := range secret {
		isValidBase32 := (c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')
		require.True(t, isValidBase32, "Secret should be valid base32: %s", secret)
	}
}

func TestTOTPAuthenticator_GenerateSecretUniqueness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)

	secrets := make(map[string]bool)

	// Generate multiple secrets and ensure they're unique.
	for range 10 {
		secret, err := auth.GenerateSecret(ctx)
		require.NoError(t, err, "GenerateSecret should succeed")
		require.False(t, secrets[secret], "Secret should be unique")
		secrets[secret] = true
	}
}

func TestTOTPAuthenticator_GenerateTOTPAndValidate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)

	// Generate a secret.
	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	// Generate a TOTP code.
	code, err := auth.GenerateTOTP(ctx, secret)
	require.NoError(t, err, "GenerateTOTP should succeed")
	require.Len(t, code, 6, "TOTP code should be 6 digits")

	// Validate the code.
	valid := auth.ValidateTOTP(ctx, secret, code)
	require.True(t, valid, "ValidateTOTP should succeed with valid code")
}

func TestTOTPAuthenticator_ValidateTOTPInvalidCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)

	// Generate a secret.
	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	// Validate with wrong code.
	valid := auth.ValidateTOTP(ctx, secret, "000000")
	require.False(t, valid, "ValidateTOTP should fail with invalid code")
}

func TestTOTPAuthenticator_ValidateTOTPInvalidSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := userauth.NewInMemoryChallengeStore()
	auth := userauth.NewTOTPAuthenticator("test-issuer", store, nil)

	// Validate with invalid secret should fail gracefully.
	valid := auth.ValidateTOTP(ctx, "invalid-base32!", "123456")
	require.False(t, valid, "ValidateTOTP should fail with invalid secret")
}
