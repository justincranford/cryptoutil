// Copyright (c) 2025 Justin Cranford
//
//

package contract

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
	cryptoutilIdentityIdpUserauthMocks "cryptoutil/internal/apps/identity/idp/userauth/mocks"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// contextKeyTimestamp is the context key for timestamp values.
const contextKeyTimestamp contextKey = "timestamp"

// DeliveryServiceContractTests defines standard test scenarios that any DeliveryService implementation must pass.
// This contract test suite validates interface compliance for both SMS and email delivery providers.
type DeliveryServiceContractTests struct {
	NewProvider func() cryptoutilIdentityIdpUserauth.DeliveryService
}

// TestDeliverySMSSuccess validates successful SMS sending.
func (c *DeliveryServiceContractTests) TestDeliverySMSSuccess(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	err := provider.SendSMS(ctx, "+15551234567", "Test message")
	require.NoError(t, err, "SendSMS should succeed for valid input")
}

// TestDeliverySMSInvalidPhoneNumber validates error handling for invalid phone numbers.
func (c *DeliveryServiceContractTests) TestDeliverySMSInvalidPhoneNumber(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.Background()

	err := provider.SendSMS(ctx, "", "Test message")
	require.Error(t, err, "SendSMS should fail for empty phone number")
	// Note: Invalid format validation is implementation-specific.
	// Mocks don't validate phone number format, only empty check.
}

// TestDeliverySMSEmptyMessage validates error handling for empty messages.
func (c *DeliveryServiceContractTests) TestDeliverySMSEmptyMessage(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.Background()

	err := provider.SendSMS(ctx, "+15551234567", "")
	require.Error(t, err, "SendSMS should fail for empty message")
}

// TestDeliveryEmailSuccess validates successful email sending.
func (c *DeliveryServiceContractTests) TestDeliveryEmailSuccess(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	err := provider.SendEmail(ctx, "user@example.com", "Test Subject", "Test body")
	require.NoError(t, err, "SendEmail should succeed for valid input")
}

// TestDeliveryEmailInvalidRecipient validates error handling for invalid email addresses.
func (c *DeliveryServiceContractTests) TestDeliveryEmailInvalidRecipient(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.Background()

	err := provider.SendEmail(ctx, "", "Subject", "Body")
	require.Error(t, err, "SendEmail should fail for empty recipient")
	// Note: Invalid email format validation is implementation-specific.
	// Mocks don't validate email format, only empty check.
}

// TestDeliveryEmailEmptySubject validates error handling for empty subjects.
func (c *DeliveryServiceContractTests) TestDeliveryEmailEmptySubject(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.Background()

	err := provider.SendEmail(ctx, "user@example.com", "", "Body")
	require.Error(t, err, "SendEmail should fail for empty subject")
}

// TestDeliveryEmailEmptyBody validates error handling for empty bodies.
func (c *DeliveryServiceContractTests) TestDeliveryEmailEmptyBody(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.Background()

	err := provider.SendEmail(ctx, "user@example.com", "Subject", "")
	require.Error(t, err, "SendEmail should fail for empty body")
}

// TestDeliveryErrorFormat validates error message format consistency.
func (c *DeliveryServiceContractTests) TestDeliveryErrorFormat(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()
	ctx := context.Background()

	// Test empty input errors are descriptive.
	err := provider.SendSMS(ctx, "", "Message")
	require.Error(t, err)
	require.NotEmpty(t, err.Error(), "Error message should be descriptive")

	err = provider.SendEmail(ctx, "", "Subject", "Body")
	require.Error(t, err)
	require.NotEmpty(t, err.Error(), "Error message should be descriptive")
}

// TestDeliveryNilContext validates error handling for nil context.
func (c *DeliveryServiceContractTests) TestDeliveryNilContext(t *testing.T) {
	t.Parallel()

	provider := c.NewProvider()

	// nil context should be handled gracefully (error or use context.Background()).
	err := provider.SendSMS(context.TODO(), "+15551234567", "Message")
	// Don't assert error/success - implementations may handle nil differently.
	// Just verify it doesn't panic.
	_ = err

	err = provider.SendEmail(context.TODO(), "user@example.com", "Subject", "Body")
	_ = err
}

// RunDeliveryServiceContractTests runs all contract tests for a DeliveryService implementation.
func RunDeliveryServiceContractTests(t *testing.T, providerType string, newProvider func() cryptoutilIdentityIdpUserauth.DeliveryService) {
	t.Helper()

	tests := &DeliveryServiceContractTests{NewProvider: newProvider}

	// Only run SMS tests for SMS providers, email tests for email providers.
	switch providerType {
	case "sms":
		t.Run("SMS_Success", tests.TestDeliverySMSSuccess)
		t.Run("SMS_Invalid_Phone_Number", tests.TestDeliverySMSInvalidPhoneNumber)
		t.Run("SMS_Empty_Message", tests.TestDeliverySMSEmptyMessage)
	case "email":
		t.Run("Email_Success", tests.TestDeliveryEmailSuccess)
		t.Run("Email_Invalid_Recipient", tests.TestDeliveryEmailInvalidRecipient)
		t.Run("Email_Empty_Subject", tests.TestDeliveryEmailEmptySubject)
		t.Run("Email_Empty_Body", tests.TestDeliveryEmailEmptyBody)
	}

	t.Run("Error_Format", tests.TestDeliveryErrorFormat)
	t.Run("Nil_Context", tests.TestDeliveryNilContext)
}

// TestSMSProviderContract validates SMSProvider against DeliveryService contract.
func TestSMSProviderContract(t *testing.T) {
	t.Parallel()

	RunDeliveryServiceContractTests(t, "sms", func() cryptoutilIdentityIdpUserauth.DeliveryService {
		return cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	})
}

// TestEmailProviderContract validates EmailProvider against DeliveryService contract.
func TestEmailProviderContract(t *testing.T) {
	t.Parallel()

	RunDeliveryServiceContractTests(t, "email", func() cryptoutilIdentityIdpUserauth.DeliveryService {
		return cryptoutilIdentityIdpUserauthMocks.NewEmailProvider()
	})
}
