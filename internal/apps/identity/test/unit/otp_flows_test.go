// Copyright (c) 2025 Justin Cranford

package unit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
	cryptoutilIdentityIdpUserauthMocks "cryptoutil/internal/apps/identity/idp/userauth/mocks"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestSMSOTPCompleteFlow tests the complete SMS OTP flow: generate → send → validate.
// NOTE: This test uses local mocks only (no HTTP servers required).
func TestSMSOTPCompleteFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup mock delivery service.
	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()

	// Setup mock challenge store.
	challengeStore := newMockChallengeStore()

	// Setup mock rate limiter (permissive for E2E test).
	rateLimiter := newMockRateLimiter(cryptoutilSharedMagic.JoseJAMaxMaterials, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute) // 100 attempts per 5 minutes

	// Setup mock user repository with test user.
	userRepo := newMockUserRepository()
	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		Email:       "test@example.com",
		PhoneNumber: "+15551234567",
		Name:        "Test User",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	// Create SMS OTP authenticator.
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator,
		mockSMS,
		challengeStore,
		rateLimiter,
		userRepo,
	)

	// Step 1: Initiate authentication (generate OTP, send SMS).
	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge)
	require.Equal(t, cryptoutilSharedMagic.AuthMethodSMSOTP, challenge.Method)
	require.Equal(t, testUser.Sub, challenge.UserID)
	require.True(t, time.Now().UTC().Before(challenge.ExpiresAt), "Challenge should not be expired")

	// Verify SMS sent.
	require.Equal(t, 1, mockSMS.GetCallCount(), "SMS should be sent")
	messages := mockSMS.GetSentMessages()
	require.Len(t, messages, 1)
	require.Contains(t, messages[0].Message, "Your verification code is:", "SMS should contain OTP")

	// Extract OTP from mock SMS (format: "Your verification code is: 123456 ...").
	var otp string

	_, err = fmt.Sscanf(messages[0].Message, "Your verification code is: %s", &otp)
	require.NoError(t, err, "OTP should be extractable from SMS")
	require.Len(t, otp, cryptoutilSharedMagic.DefaultOTPLength, "OTP should be 6 digits")

	// Step 2: Verify authentication with correct OTP.
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), otp)
	require.NoError(t, err, "VerifyAuth should succeed with correct OTP")
	require.NotNil(t, user)
	require.Equal(t, testUser.Sub, user.Sub)
	require.Equal(t, testUser.PhoneNumber, user.PhoneNumber)

	// Step 3: Verify challenge deleted (single-use).
	_, _, err = challengeStore.Retrieve(ctx, challenge.ID)
	require.Error(t, err, "Challenge should be deleted after successful verification")
}

// TestSMSOTPInvalidToken tests OTP validation with incorrect token.
func TestSMSOTPInvalidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(cryptoutilSharedMagic.JoseJAMaxMaterials, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		PhoneNumber: "+15551234567",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator, mockSMS, challengeStore, rateLimiter, userRepo,
	)

	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err)

	// Attempt verification with WRONG OTP.
	wrongOTP := "000000"
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), wrongOTP)

	require.Error(t, err, "VerifyAuth should fail with wrong OTP")
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid OTP", "Error should indicate invalid OTP")
}

// TestSMSOTPExpiredChallenge tests OTP validation after expiration.
func TestSMSOTPExpiredChallenge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(cryptoutilSharedMagic.JoseJAMaxMaterials, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		PhoneNumber: "+15551234567",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator, mockSMS, challengeStore, rateLimiter, userRepo,
	)

	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err)

	// Manually expire challenge.
	challenge.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute)
	challengeStore.challenges[challenge.ID] = challengeEntry{
		challenge: challenge,
		token:     challengeStore.challenges[challenge.ID].token,
	}

	// Attempt verification with expired challenge.
	otp := "123456" // Doesn't matter, challenge expired
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), otp)

	require.Error(t, err, "VerifyAuth should fail with expired challenge")
	require.Nil(t, user)
	require.Contains(t, err.Error(), "expired", "Error should indicate expiration")
}

// TestSMSOTPRateLimitEnforcement tests per-user rate limiting enforcement.
func TestSMSOTPRateLimitEnforcement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(3, 1*time.Minute) // Only 3 attempts per minute
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		PhoneNumber: "+15551234567",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator, mockSMS, challengeStore, rateLimiter, userRepo,
	)

	// First 3 attempts should succeed.
	for i := 0; i < 3; i++ {
		_, err := authenticator.InitiateAuth(ctx, testUser.Sub)
		require.NoError(t, err, "Attempt %d should succeed", i+1)
	}

	// 4th attempt should fail due to rate limiting.
	_, err = authenticator.InitiateAuth(ctx, testUser.Sub)
	require.Error(t, err, "4th attempt should fail due to rate limit")
	require.Contains(t, err.Error(), "rate limit", "Error should indicate rate limiting")
}

// TestEmailMagicLinkCompleteFlow tests the complete magic link flow: generate → send → validate.
func TestEmailMagicLinkCompleteFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockEmail := cryptoutilIdentityIdpUserauthMocks.NewEmailProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(cryptoutilSharedMagic.JoseJAMaxMaterials, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(
		generator, mockEmail, challengeStore, rateLimiter, userRepo,
		"https://example.com", // base URL
	)

	// Step 1: Initiate authentication (generate token, send email).
	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge)
	require.Equal(t, "magic_link", challenge.Method)
	require.Equal(t, testUser.Sub, challenge.UserID)

	// Verify email sent.
	require.Equal(t, 1, mockEmail.GetCallCount(), "Email should be sent")
	emails := mockEmail.GetSentEmails()
	require.Len(t, emails, 1)
	require.Contains(t, emails[0].Body, "https://example.com/auth/magic-link/verify", "Email should contain magic link")

	// Extract token from email body (format: "...?token=abc123&challenge=uuid").
	var token string

	lines := splitLines(emails[0].Body)
	for _, line := range lines {
		if strings.Contains(line, "https://example.com/auth/magic-link/verify?token=") {
			parts := strings.Split(line, "token=")
			if len(parts) >= 2 {
				tokenPart := strings.Split(parts[1], "&")[0]
				token = tokenPart

				break
			}
		}
	}

	require.NotEmpty(t, token, "Token should be extractable from email")

	// Step 2: Verify authentication with correct token.
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), token)
	require.NoError(t, err, "VerifyAuth should succeed with correct token")
	require.NotNil(t, user)
	require.Equal(t, testUser.Sub, user.Sub)
	require.Equal(t, testUser.Email, user.Email)

	// Step 3: Verify challenge deleted (single-use).
	_, _, err = challengeStore.Retrieve(ctx, challenge.ID)
	require.Error(t, err, "Challenge should be deleted after successful verification")
}

// TestMagicLinkInvalidToken tests magic link validation with incorrect token.
