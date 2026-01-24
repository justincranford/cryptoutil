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

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/identity/idp/userauth"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// mockMagicLinkUserRepo implements UserRepository for magic link testing.
type mockMagicLinkUserRepo struct {
	users map[string]*cryptoutilIdentityDomain.User
}

func newMockMagicLinkUserRepo() *mockMagicLinkUserRepo {
	return &mockMagicLinkUserRepo{
		users: make(map[string]*cryptoutilIdentityDomain.User),
	}
}

func (m *mockMagicLinkUserRepo) GetBySub(_ context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.users[sub]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", sub)
	}

	return user, nil
}

func (m *mockMagicLinkUserRepo) Create(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockMagicLinkUserRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockMagicLinkUserRepo) GetByUsername(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockMagicLinkUserRepo) GetByEmail(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockMagicLinkUserRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockMagicLinkUserRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockMagicLinkUserRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockMagicLinkUserRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockMagicLinkUserRepo) AddUser(user *cryptoutilIdentityDomain.User) {
	m.users[user.Sub] = user
}

func TestMagicLinkAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(nil, nil, nil, nil, nil, "https://example.com")
	require.NotNil(t, auth, "NewMagicLinkAuthenticator should return non-nil authenticator")
}

func TestMagicLinkAuthenticator_Method(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(nil, nil, nil, nil, nil, "https://example.com")
	require.Equal(t, "magic_link", auth.Method(), "Method should return 'magic_link'")
}

func TestRiskBasedAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	thresholds := cryptoutilIdentityIdpUserauth.DefaultRiskThresholds()
	auth := cryptoutilIdentityIdpUserauth.NewRiskBasedAuthenticator(nil, nil, nil, thresholds, nil)
	require.NotNil(t, auth, "NewRiskBasedAuthenticator should return non-nil authenticator")
}

func TestRiskBasedAuthenticator_Method(t *testing.T) {
	t.Parallel()

	thresholds := cryptoutilIdentityIdpUserauth.DefaultRiskThresholds()
	auth := cryptoutilIdentityIdpUserauth.NewRiskBasedAuthenticator(nil, nil, nil, thresholds, nil)
	require.Equal(t, "risk_based", auth.Method(), "Method should return 'risk_based'")
}

func TestRiskBasedAuthenticator_DefaultThresholds(t *testing.T) {
	t.Parallel()

	thresholds := cryptoutilIdentityIdpUserauth.DefaultRiskThresholds()
	require.NotNil(t, thresholds, "DefaultRiskThresholds should return non-nil thresholds")
	require.Len(t, thresholds, 4, "Should have 4 risk levels: low, medium, high, critical")

	// Verify keys exist.
	require.Contains(t, thresholds, cryptoutilIdentityIdpUserauth.RiskLevelLow, "Should have low risk level")
	require.Contains(t, thresholds, cryptoutilIdentityIdpUserauth.RiskLevelMedium, "Should have medium risk level")
	require.Contains(t, thresholds, cryptoutilIdentityIdpUserauth.RiskLevelHigh, "Should have high risk level")
	require.Contains(t, thresholds, cryptoutilIdentityIdpUserauth.RiskLevelCritical, "Should have critical risk level")

	// Verify MinFactors increases with risk level.
	require.Less(t, thresholds[cryptoutilIdentityIdpUserauth.RiskLevelLow].MinFactors, thresholds[cryptoutilIdentityIdpUserauth.RiskLevelMedium].MinFactors,
		"Low risk should require fewer factors than medium")
}

// TestRiskBasedAuthenticator_InitiateAuth tests InitiateAuth.
func TestRiskBasedAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	thresholds := cryptoutilIdentityIdpUserauth.DefaultRiskThresholds()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewRiskBasedAuthenticator(nil, nil, challengeStore, thresholds, nil)

	userID := "test-user-risk"

	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID, challenge.UserID, "Challenge UserID should match")
	require.Equal(t, "risk_based", challenge.Method, "Challenge Method should match")
}

// TestRiskBasedAuthenticator_VerifyAuth tests VerifyAuth.
func TestRiskBasedAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	thresholds := cryptoutilIdentityIdpUserauth.DefaultRiskThresholds()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewRiskBasedAuthenticator(nil, nil, challengeStore, thresholds, nil)

	userID := "test-user-risk-verify"

	// Initiate auth first.
	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")

	// VerifyAuth with invalid challenge ID.
	_, err = auth.VerifyAuth(ctx, "invalid-uuid", "response")
	require.Error(t, err, "VerifyAuth should fail with invalid challenge ID")
	require.Contains(t, err.Error(), "invalid challenge ID", "Error should indicate invalid challenge ID")

	// VerifyAuth with valid challenge - should return error requiring context-specific verification.
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), "response")
	require.Error(t, err, "VerifyAuth should fail with context-specific verification required")
}

// TestRiskBasedAuthenticator_VerifyAuthChallengeNotFound tests VerifyAuth with non-existent challenge.
func TestRiskBasedAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	thresholds := cryptoutilIdentityIdpUserauth.DefaultRiskThresholds()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewRiskBasedAuthenticator(nil, nil, challengeStore, thresholds, nil)

	// Generate a valid UUID that doesn't exist as a challenge.
	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	_, err = auth.VerifyAuth(ctx, nonExistentID.String(), "response")
	require.Error(t, err, "VerifyAuth should fail with non-existent challenge")
	require.Contains(t, err.Error(), "challenge not found", "Error should indicate challenge not found")
}

func TestMockDeliveryService_NewService(t *testing.T) {
	t.Parallel()

	service := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
	require.NotNil(t, service, "NewMockDeliveryService should return non-nil service")
}

func TestMockDeliveryService_SendSMSSuccess(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	service := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()

	err := service.SendSMS(ctx, "+1234567890", "Test message")
	require.NoError(t, err, "SendSMS should succeed")

	sent := service.GetSentSMS()
	require.Len(t, sent, 1, "Should have one sent SMS")
	require.Equal(t, "+1234567890", sent[0].PhoneNumber, "Phone number should match")
	require.Equal(t, "Test message", sent[0].Message, "Message should match")
}

func TestMockDeliveryService_SendEmailSuccess(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	service := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()

	err := service.SendEmail(ctx, "test@example.com", "Test Subject", "Test body")
	require.NoError(t, err, "SendEmail should succeed")

	sent := service.GetSentEmails()
	require.Len(t, sent, 1, "Should have one sent email")
	require.Equal(t, "test@example.com", sent[0].To, "Email address should match")
	require.Equal(t, "Test Subject", sent[0].Subject, "Subject should match")
	require.Equal(t, "Test body", sent[0].Body, "Body should match")
}

func TestMockDeliveryService_SetShouldFail(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	service := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()

	service.SetShouldFail(true)

	err := service.SendSMS(ctx, "+1234567890", "Test message")
	require.Error(t, err, "SendSMS should fail when SetShouldFail is true")

	err = service.SendEmail(ctx, "test@example.com", "Subject", "Body")
	require.Error(t, err, "SendEmail should fail when SetShouldFail is true")
}

func TestMockDeliveryService_Reset(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	service := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()

	// Send some messages.
	err := service.SendSMS(ctx, "+1234567890", "Test message")
	require.NoError(t, err, "SendSMS should succeed")

	err = service.SendEmail(ctx, "test@example.com", "Subject", "Body")
	require.NoError(t, err, "SendEmail should succeed")

	// Reset.
	service.Reset()

	// Verify reset.
	require.Empty(t, service.GetSentSMS(), "SMS list should be empty after reset")
	require.Empty(t, service.GetSentEmails(), "Email list should be empty after reset")
}

func TestStepUpAuthenticator_DefaultPolicies(t *testing.T) {
	t.Parallel()

	policies := cryptoutilIdentityIdpUserauth.DefaultStepUpPolicies()
	require.NotNil(t, policies, "DefaultStepUpPolicies should return non-nil policies")
	require.NotEmpty(t, policies, "Should have at least one default policy")
}

// TestStepUpAuthenticator_Method tests Method.
func TestStepUpAuthenticator_Method(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewStepUpAuthenticator(nil, nil, nil, nil, nil)
	require.Equal(t, "step_up", auth.Method(), "Method should return 'step_up'")
}

// TestStepUpAuthenticator_InitiateAuth tests InitiateAuth.
func TestStepUpAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewStepUpAuthenticator(nil, nil, nil, challengeStore, nil)

	userID := "test-user-stepup"

	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID, challenge.UserID, "Challenge UserID should match")
	require.Equal(t, "step_up", challenge.Method, "Challenge Method should match")
}

// TestStepUpAuthenticator_VerifyAuth tests VerifyAuth.
func TestStepUpAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewStepUpAuthenticator(nil, nil, nil, challengeStore, nil)

	userID := "test-user-stepup-verify"

	// Initiate auth first.
	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")

	// VerifyAuth with invalid challenge ID.
	_, err = auth.VerifyAuth(ctx, "invalid-uuid", "response")
	require.Error(t, err, "VerifyAuth should fail with invalid challenge ID")
	require.Contains(t, err.Error(), "invalid challenge ID", "Error should indicate invalid challenge ID")
}

// TestStepUpAuthenticator_VerifyAuthChallengeNotFound tests VerifyAuth with non-existent challenge.
func TestStepUpAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	auth := cryptoutilIdentityIdpUserauth.NewStepUpAuthenticator(nil, nil, nil, challengeStore, nil)

	// Generate a valid UUID that doesn't exist as a challenge.
	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	_, err = auth.VerifyAuth(ctx, nonExistentID.String(), "response")
	require.Error(t, err, "VerifyAuth should fail with non-existent challenge")
	require.Contains(t, err.Error(), "challenge not found", "Error should indicate challenge not found")
}

// TestMagicLinkAuthenticator_InitiateAuth tests InitiateAuth.
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
	token, err := generator.GenerateSecureToken(cryptoutilIdentityMagic.DefaultMagicLinkLength)
	require.NoError(t, err)

	hashedToken, err := cryptoutilIdentityIdpUserauth.HashToken(token)
	require.NoError(t, err)

	expiredChallenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID.String(),
		Method:    "magic_link",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago.
		Metadata:  map[string]any{"email": user.Email},
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
	token, err := generator.GenerateSecureToken(cryptoutilIdentityMagic.DefaultMagicLinkLength)
	require.NoError(t, err)

	hashedToken, err := cryptoutilIdentityIdpUserauth.HashToken(token)
	require.NoError(t, err)

	// Create challenge manually with known token.
	challenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID.String(),
		Method:    "magic_link",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Metadata:  map[string]any{"email": user.Email},
	}

	err = challengeStore.Store(ctx, challenge, hashedToken)
	require.NoError(t, err)

	// Verify with correct token.
	verifiedUser, err := auth.VerifyAuth(ctx, challenge.ID.String(), token)
	require.NoError(t, err, "VerifyAuth should succeed with correct token")
	require.NotNil(t, verifiedUser, "User should be returned")
	require.Equal(t, userID, verifiedUser.ID, "User ID should match")
}
