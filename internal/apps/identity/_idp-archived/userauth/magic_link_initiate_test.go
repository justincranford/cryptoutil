// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestMagicLinkAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Add user with email.
	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "test@example.com",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID.String(), challenge.UserID, "Challenge UserID should match")
	require.Equal(t, "magic_link", challenge.Method, "Challenge Method should match")

	// Verify email was sent.
	sent := delivery.GetSentEmails()
	require.Len(t, sent, 1, "One email should have been sent")
	require.Equal(t, "test@example.com", sent[0].To, "Email recipient should match")
}

// TestMagicLinkAuthenticator_InitiateAuthUserNotFound tests InitiateAuth with non-existent user.
func TestMagicLinkAuthenticator_InitiateAuthUserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	challenge, err := auth.InitiateAuth(ctx, nonExistentID.String())
	require.Error(t, err, "InitiateAuth should fail for non-existent user")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "user not found", "Error should indicate user not found")
}

// TestMagicLinkAuthenticator_InitiateAuthNoEmail tests InitiateAuth without email.
func TestMagicLinkAuthenticator_InitiateAuthNoEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Add user without email.
	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.Error(t, err, "InitiateAuth should fail without email")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "no email", "Error should indicate missing email")
}

// TestMagicLinkAuthenticator_InitiateAuthTokenGenerationFailure tests token generation failure.
func TestMagicLinkAuthenticator_InitiateAuthTokenGenerationFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	// Mock OTPGenerator that fails.
	generator := &mockFailingOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "test@example.com",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.Error(t, err, "InitiateAuth should fail when token generation fails")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "failed to generate token", "Error should indicate token generation failure")
}

// mockFailingOTPGenerator always fails to generate tokens.
type mockFailingOTPGenerator struct{}

func (m *mockFailingOTPGenerator) GenerateOTP(_ int) (string, error) {
	return "", fmt.Errorf("random number generator failed")
}

func (m *mockFailingOTPGenerator) GenerateSecureToken(_ int) (string, error) {
	return "", fmt.Errorf("random number generator failed")
}

// TestMagicLinkAuthenticator_InitiateAuthEmailDeliveryFailure tests email delivery failure.
func TestMagicLinkAuthenticator_InitiateAuthEmailDeliveryFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	delivery.SetShouldFail(true) // Make delivery fail.

	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "test@example.com",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.Error(t, err, "InitiateAuth should fail when email delivery fails")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "failed to send email", "Error should indicate email delivery failure")
}

// TestMagicLinkAuthenticator_VerifyAuth tests VerifyAuth.
func TestMagicLinkAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Add user with email.
	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "test@example.com",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	// Initiate auth first.
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")

	// VerifyAuth with invalid challenge ID.
	_, err = auth.VerifyAuth(ctx, "invalid-uuid", "some-token")
	require.Error(t, err, "VerifyAuth should fail with invalid challenge ID")
	require.Contains(t, err.Error(), "invalid challenge ID", "Error should indicate invalid challenge ID")

	// VerifyAuth with wrong token (challenge exists but token is wrong).
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), "wrong-token")
	require.Error(t, err, "VerifyAuth should fail with wrong token")
}

// TestMagicLinkAuthenticator_VerifyAuthChallengeNotFound tests VerifyAuth with non-existent challenge.
func TestMagicLinkAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	// Generate a valid UUID that doesn't exist as a challenge.
	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	_, err = auth.VerifyAuth(ctx, nonExistentID.String(), "some-token")
	require.Error(t, err, "VerifyAuth should fail with non-existent challenge")
	require.Contains(t, err.Error(), "challenge not found", "Error should indicate challenge not found")
}

// TestMagicLinkAuthenticator_VerifyAuthExpired tests VerifyAuth with expired challenge.
func TestMagicLinkAuthenticator_VerifyAuthExpired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "test@example.com",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	// Create expired challenge manually.
	token, err := generator.GenerateSecureToken(cryptoutilSharedMagic.DefaultMagicLinkLength)
	require.NoError(t, err)

	hashedToken, err := cryptoutilIdentityIdpUserauth.HashToken(token)
	require.NoError(t, err)

	expiredChallenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID.String(),
		Method:    "magic_link",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // Expired 1 hour ago.
		Metadata:  map[string]any{cryptoutilSharedMagic.ClaimEmail: user.Email},
	}

	err = challengeStore.Store(ctx, expiredChallenge, hashedToken)
	require.NoError(t, err)

	// Verify with expired challenge.
	_, err = auth.VerifyAuth(ctx, expiredChallenge.ID.String(), token)
	require.Error(t, err, "VerifyAuth should fail with expired challenge")
	require.Contains(t, err.Error(), "expired", "Error should indicate expiration")
}

// TestMagicLinkAuthenticator_VerifyAuthSuccess tests successful magic link verification.
func TestMagicLinkAuthenticator_VerifyAuthSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockMagicLinkUserRepo()
	delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	user := &cryptoutilIdentityDomain.User{
		ID:    userID,
		Sub:   userID.String(),
		Email: "test@example.com",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.com")

	// Generate token before initiating auth so we can capture it.
	token, err := generator.GenerateSecureToken(cryptoutilSharedMagic.DefaultMagicLinkLength)
	require.NoError(t, err)

	hashedToken, err := cryptoutilIdentityIdpUserauth.HashToken(token)
	require.NoError(t, err)

	// Create challenge manually with known token.
	challenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID.String(),
		Method:    "magic_link",
		ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
		Metadata:  map[string]any{cryptoutilSharedMagic.ClaimEmail: user.Email},
	}

	err = challengeStore.Store(ctx, challenge, hashedToken)
	require.NoError(t, err)

	// Verify with correct token.
	verifiedUser, err := auth.VerifyAuth(ctx, challenge.ID.String(), token)
	require.NoError(t, err, "VerifyAuth should succeed with correct token")
	require.NotNil(t, verifiedUser, "User should be returned")
	require.Equal(t, userID, verifiedUser.ID, "User ID should match")
}
