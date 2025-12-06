// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/idp/userauth"
)

func TestMagicLinkAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	auth := userauth.NewMagicLinkAuthenticator(nil, nil, nil, nil, nil, "https://example.com")
	require.NotNil(t, auth, "NewMagicLinkAuthenticator should return non-nil authenticator")
}

func TestMagicLinkAuthenticator_Method(t *testing.T) {
	t.Parallel()

	auth := userauth.NewMagicLinkAuthenticator(nil, nil, nil, nil, nil, "https://example.com")
	require.Equal(t, "magic_link", auth.Method(), "Method should return 'magic_link'")
}

func TestRiskBasedAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	thresholds := userauth.DefaultRiskThresholds()
	auth := userauth.NewRiskBasedAuthenticator(nil, nil, nil, thresholds, nil)
	require.NotNil(t, auth, "NewRiskBasedAuthenticator should return non-nil authenticator")
}

func TestRiskBasedAuthenticator_Method(t *testing.T) {
	t.Parallel()

	thresholds := userauth.DefaultRiskThresholds()
	auth := userauth.NewRiskBasedAuthenticator(nil, nil, nil, thresholds, nil)
	require.Equal(t, "risk_based", auth.Method(), "Method should return 'risk_based'")
}

func TestRiskBasedAuthenticator_DefaultThresholds(t *testing.T) {
	t.Parallel()

	thresholds := userauth.DefaultRiskThresholds()
	require.NotNil(t, thresholds, "DefaultRiskThresholds should return non-nil thresholds")
	require.Len(t, thresholds, 4, "Should have 4 risk levels: low, medium, high, critical")

	// Verify keys exist.
	require.Contains(t, thresholds, userauth.RiskLevelLow, "Should have low risk level")
	require.Contains(t, thresholds, userauth.RiskLevelMedium, "Should have medium risk level")
	require.Contains(t, thresholds, userauth.RiskLevelHigh, "Should have high risk level")
	require.Contains(t, thresholds, userauth.RiskLevelCritical, "Should have critical risk level")

	// Verify MinFactors increases with risk level.
	require.Less(t, thresholds[userauth.RiskLevelLow].MinFactors, thresholds[userauth.RiskLevelMedium].MinFactors,
		"Low risk should require fewer factors than medium")
}

func TestMockDeliveryService_NewService(t *testing.T) {
	t.Parallel()

	service := userauth.NewMockDeliveryService()
	require.NotNil(t, service, "NewMockDeliveryService should return non-nil service")
}

func TestMockDeliveryService_SendSMSSuccess(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	service := userauth.NewMockDeliveryService()

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
	service := userauth.NewMockDeliveryService()

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
	service := userauth.NewMockDeliveryService()

	service.SetShouldFail(true)

	err := service.SendSMS(ctx, "+1234567890", "Test message")
	require.Error(t, err, "SendSMS should fail when SetShouldFail is true")

	err = service.SendEmail(ctx, "test@example.com", "Subject", "Body")
	require.Error(t, err, "SendEmail should fail when SetShouldFail is true")
}

func TestMockDeliveryService_Reset(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	service := userauth.NewMockDeliveryService()

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

	policies := userauth.DefaultStepUpPolicies()
	require.NotNil(t, policies, "DefaultStepUpPolicies should return non-nil policies")
	require.NotEmpty(t, policies, "Should have at least one default policy")
}
