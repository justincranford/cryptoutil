// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
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
