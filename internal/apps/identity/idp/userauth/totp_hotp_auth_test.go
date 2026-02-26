// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
)

func TestTOTPAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)
	require.NotNil(t, auth, "NewTOTPAuthenticator should return non-nil authenticator")
}

func TestTOTPAuthenticator_Method(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)
	require.Equal(t, cryptoutilSharedMagic.MFATypeTOTP, auth.Method(), "Method should return 'totp'")
}

func TestTOTPAuthenticator_GenerateSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

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
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	secrets := make(map[string]bool)

	// Generate multiple secrets and ensure they're unique.
	for range cryptoutilSharedMagic.JoseJADefaultMaxMaterials {
		secret, err := auth.GenerateSecret(ctx)
		require.NoError(t, err, "GenerateSecret should succeed")
		require.False(t, secrets[secret], "Secret should be unique")
		secrets[secret] = true
	}
}

func TestTOTPAuthenticator_GenerateTOTPAndValidate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	// Generate a secret.
	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	// Generate a TOTP code.
	code, err := auth.GenerateTOTP(ctx, secret)
	require.NoError(t, err, "GenerateTOTP should succeed")
	require.Len(t, code, cryptoutilSharedMagic.DefaultEmailOTPLength, "TOTP code should be 6 digits")

	// Validate the code.
	valid := auth.ValidateTOTP(ctx, secret, code)
	require.True(t, valid, "ValidateTOTP should succeed with valid code")
}

func TestTOTPAuthenticator_ValidateTOTPInvalidCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

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
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	// Validate with invalid secret should fail gracefully.
	valid := auth.ValidateTOTP(ctx, "invalid-base32!", "123456")
	require.False(t, valid, "ValidateTOTP should fail with invalid secret")
}

// mockCounterStore implements CounterStore for testing HOTP.
type mockCounterStore struct {
	counters map[string]uint64
}

func newMockCounterStore() *mockCounterStore {
	return &mockCounterStore{
		counters: make(map[string]uint64),
	}
}

func (m *mockCounterStore) GetCounter(_ context.Context, userID string) (uint64, error) {
	counter, ok := m.counters[userID]
	if !ok {
		return 0, nil // Default to 0 for new users.
	}

	return counter, nil
}

func (m *mockCounterStore) IncrementCounter(_ context.Context, userID string) (uint64, error) {
	m.counters[userID]++

	return m.counters[userID], nil
}

func (m *mockCounterStore) SetCounter(_ context.Context, userID string, counter uint64) error {
	m.counters[userID] = counter

	return nil
}

// TestHOTPAuthenticator_NewAuthenticator tests NewHOTPAuthenticator.
func TestHOTPAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)
	require.NotNil(t, auth, "NewHOTPAuthenticator should return non-nil authenticator")
}

// TestHOTPAuthenticator_Method tests Method.
func TestHOTPAuthenticator_Method(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)
	require.Equal(t, "hotp", auth.Method(), "Method should return 'hotp'")
}

// TestHOTPAuthenticator_GenerateSecret tests GenerateSecret.
func TestHOTPAuthenticator_GenerateSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")
	require.NotEmpty(t, secret, "Secret should not be empty")

	// Secret should be base32 encoded.
	for _, c := range secret {
		isValidBase32 := (c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7')
		require.True(t, isValidBase32, "Secret should be valid base32: %s", secret)
	}
}

// TestHOTPAuthenticator_GenerateSecretUniqueness tests that secrets are unique.
func TestHOTPAuthenticator_GenerateSecretUniqueness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	secrets := make(map[string]bool)

	for range cryptoutilSharedMagic.JoseJADefaultMaxMaterials {
		secret, err := auth.GenerateSecret(ctx)
		require.NoError(t, err, "GenerateSecret should succeed")
		require.False(t, secrets[secret], "Secret should be unique")
		secrets[secret] = true
	}
}

// TestHOTPAuthenticator_GenerateHOTP tests GenerateHOTP.
func TestHOTPAuthenticator_GenerateHOTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	// Generate a secret.
	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	// Generate HOTP codes at different counters.
	code0, err := auth.GenerateHOTP(ctx, secret, 0)
	require.NoError(t, err, "GenerateHOTP should succeed for counter 0")
	require.Len(t, code0, cryptoutilSharedMagic.DefaultEmailOTPLength, "HOTP code should be 6 digits")

	code1, err := auth.GenerateHOTP(ctx, secret, 1)
	require.NoError(t, err, "GenerateHOTP should succeed for counter 1")
	require.Len(t, code1, cryptoutilSharedMagic.DefaultEmailOTPLength, "HOTP code should be 6 digits")

	// Codes at different counters should be different (with high probability).
	// Note: There's a tiny chance they could be the same, but it's astronomically unlikely.
	require.NotEqual(t, code0, code1, "HOTP codes at different counters should differ")
}

// TestHOTPAuthenticator_GenerateHOTPDeterministic tests that same secret+counter produces same code.
func TestHOTPAuthenticator_GenerateHOTPDeterministic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	// Use a fixed secret for deterministic testing.
	secret := "JBSWY3DPEHPK3PXP"

	code1, err := auth.GenerateHOTP(ctx, secret, 0)
	require.NoError(t, err, "GenerateHOTP should succeed")

	code2, err := auth.GenerateHOTP(ctx, secret, 0)
	require.NoError(t, err, "GenerateHOTP should succeed")

	require.Equal(t, code1, code2, "Same secret+counter should produce same HOTP code")
}

// TestHOTPAuthenticator_ValidateHOTP tests ValidateHOTP.
func TestHOTPAuthenticator_ValidateHOTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	// Generate a secret.
	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	userID := "test-user-hotp"

	// Generate HOTP code at counter 0.
	code, err := auth.GenerateHOTP(ctx, secret, 0)
	require.NoError(t, err, "GenerateHOTP should succeed")

	// Validate the code.
	valid, err := auth.ValidateHOTP(ctx, userID, secret, code)
	require.NoError(t, err, "ValidateHOTP should succeed")
	require.True(t, valid, "ValidateHOTP should return true for valid code")

	// Counter should have been incremented.
	counter, err := counterStore.GetCounter(ctx, userID)
	require.NoError(t, err, "GetCounter should succeed")
	require.Equal(t, uint64(1), counter, "Counter should be incremented after validation")
}

// TestHOTPAuthenticator_ValidateHOTPInvalid tests ValidateHOTP with invalid code.
func TestHOTPAuthenticator_ValidateHOTPInvalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	userID := "test-user-hotp-invalid"

	// Validate with invalid code.
	valid, err := auth.ValidateHOTP(ctx, userID, secret, "000000")
	require.NoError(t, err, "ValidateHOTP should not error for invalid code")
	require.False(t, valid, "ValidateHOTP should return false for invalid code")
}

// TestHOTPAuthenticator_ValidateHOTPLookahead tests lookahead window.
func TestHOTPAuthenticator_ValidateHOTPLookahead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	secret, err := auth.GenerateSecret(ctx)
	require.NoError(t, err, "GenerateSecret should succeed")

	userID := "test-user-hotp-lookahead"

	// Generate HOTP code at counter 5 (within lookahead window of 10).
	code, err := auth.GenerateHOTP(ctx, secret, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.NoError(t, err, "GenerateHOTP should succeed")

	// Validate - should succeed due to lookahead.
	valid, err := auth.ValidateHOTP(ctx, userID, secret, code)
	require.NoError(t, err, "ValidateHOTP should succeed")
	require.True(t, valid, "ValidateHOTP should succeed within lookahead window")

	// Counter should be set to 6 after validation.
	counter, err := counterStore.GetCounter(ctx, userID)
	require.NoError(t, err, "GetCounter should succeed")
	require.Equal(t, uint64(cryptoutilSharedMagic.DefaultEmailOTPLength), counter, "Counter should be set to next value after validation")
}

// TestHOTPAuthenticator_InitiateAuth tests InitiateAuth.
func TestHOTPAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	userID := "test-user-initiate"

	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID, challenge.UserID, "Challenge UserID should match")
	require.Equal(t, "hotp", challenge.Method, "Challenge Method should be 'hotp'")
}

// TestTOTPAuthenticator_InitiateAuth tests TOTP InitiateAuth.
func TestTOTPAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	userID := "test-user-totp-initiate"

	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID, challenge.UserID, "Challenge UserID should match")
	require.Equal(t, cryptoutilSharedMagic.MFATypeTOTP, challenge.Method, "Challenge Method should be 'totp'")
}

// TestTOTPAuthenticator_VerifyAuth tests VerifyAuth.
func TestTOTPAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	userID := "test-user-totp-verify"

	// Initiate auth first.
	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")

	// VerifyAuth with invalid challenge ID.
	_, err = auth.VerifyAuth(ctx, "invalid-uuid", "123456")
	require.Error(t, err, "VerifyAuth should fail with invalid challenge ID")
	require.Contains(t, err.Error(), "invalid challenge ID", "Error should indicate invalid challenge ID")

	// VerifyAuth with wrong TOTP code (challenge exists but code is wrong).
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), "000000")
	require.Error(t, err, "VerifyAuth should fail with wrong TOTP code")
}

// TestTOTPAuthenticator_VerifyAuthChallengeNotFound tests VerifyAuth with non-existent challenge.
func TestTOTPAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	// Generate a valid UUID that doesn't exist as a challenge.
	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	_, err = auth.VerifyAuth(ctx, nonExistentID.String(), "123456")
	require.Error(t, err, "VerifyAuth should fail with non-existent challenge")
	require.Contains(t, err.Error(), "challenge not found", "Error should indicate challenge not found")
}

// TestHOTPAuthenticator_VerifyAuth tests HOTP VerifyAuth.
func TestHOTPAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	userID := "test-user-hotp-verify"

	// Initiate auth first.
	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")

	// VerifyAuth with invalid challenge ID.
	_, err = auth.VerifyAuth(ctx, "invalid-uuid", "123456")
	require.Error(t, err, "VerifyAuth should fail with invalid challenge ID")
	require.Contains(t, err.Error(), "invalid challenge ID", "Error should indicate invalid challenge ID")
}

// TestHOTPAuthenticator_VerifyAuthChallengeNotFound tests VerifyAuth with non-existent challenge.
func TestHOTPAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	counterStore := newMockCounterStore()
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, counterStore)

	// Generate a valid UUID that doesn't exist as a challenge.
	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	_, err = auth.VerifyAuth(ctx, nonExistentID.String(), "123456")
	require.Error(t, err, "VerifyAuth should fail with non-existent challenge")
	require.Contains(t, err.Error(), "challenge not found", "Error should indicate challenge not found")
}
