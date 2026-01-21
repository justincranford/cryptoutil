// Copyright (c) 2025 Justin Cranford
//
//

package mocks

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSMSProviderSuccess tests successful SMS sending.
func TestSMSProviderSuccess(t *testing.T) {
	t.Parallel()

	provider := NewSMSProvider()
	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	err := provider.SendSMS(ctx, "+15551234567", "Test message")
	require.NoError(t, err)

	messages := provider.GetSentMessages()
	require.Len(t, messages, 1)
	require.Equal(t, "+15551234567", messages[0].PhoneNumber)
	require.Equal(t, "Test message", messages[0].Message)
	require.Equal(t, int64(1234567890), messages[0].Timestamp)
	require.Equal(t, 1, provider.GetCallCount())
}

// TestSMSProviderFailure tests SMS provider failure mode.
func TestSMSProviderFailure(t *testing.T) {
	t.Parallel()

	provider := NewSMSProvider()
	provider.SetShouldFail(true, fmt.Errorf("network timeout"))

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	err := provider.SendSMS(ctx, "+15551234567", "Test message")
	require.Error(t, err)
	require.Contains(t, err.Error(), "network timeout")

	messages := provider.GetSentMessages()
	require.Len(t, messages, 0, "Failed sends should not record messages")
}

// TestSMSProviderNetworkErrorInjection tests network error at specific calls.
func TestSMSProviderNetworkErrorInjection(t *testing.T) {
	t.Parallel()

	provider := NewSMSProvider()
	provider.InjectNetworkError(2, fmt.Errorf("connection refused"))

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	// First call succeeds.
	err := provider.SendSMS(ctx, "+15551234567", "First message")
	require.NoError(t, err)

	// Second call fails with injected error.
	err = provider.SendSMS(ctx, "+15551234567", "Second message")
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection refused")

	// Third call succeeds.
	err = provider.SendSMS(ctx, "+15551234567", "Third message")
	require.NoError(t, err)

	messages := provider.GetSentMessages()
	require.Len(t, messages, 2, "Only successful sends should be recorded")
	require.Equal(t, 3, provider.GetCallCount())
}

// TestSMSProviderReset tests provider reset functionality.
func TestSMSProviderReset(t *testing.T) {
	t.Parallel()

	provider := NewSMSProvider()
	provider.SetShouldFail(true, nil)
	provider.InjectNetworkError(1, fmt.Errorf("test error"))

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	_ = provider.SendSMS(ctx, "+15551234567", "Test message") //nolint:errcheck // Test setup - error intentionally ignored to test reset functionality

	provider.Reset()

	// After reset, provider should work normally.
	err := provider.SendSMS(ctx, "+15551234567", "After reset")
	require.NoError(t, err)

	messages := provider.GetSentMessages()
	require.Len(t, messages, 1, "Reset should clear previous messages")
	require.Equal(t, 1, provider.GetCallCount(), "Reset should clear call count")
}

// TestEmailProviderSuccess tests successful email sending.
func TestEmailProviderSuccess(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	err := provider.SendEmail(ctx, "user@example.com", "Test Subject", "Test body")
	require.NoError(t, err)

	emails := provider.GetSentEmails()
	require.Len(t, emails, 1)
	require.Equal(t, "user@example.com", emails[0].To)
	require.Equal(t, "Test Subject", emails[0].Subject)
	require.Equal(t, "Test body", emails[0].Body)
	require.Equal(t, int64(1234567890), emails[0].Timestamp)
	require.Equal(t, 1, provider.GetCallCount())
}

// TestEmailProviderFailure tests email provider failure mode.
func TestEmailProviderFailure(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	provider.SetShouldFail(true, fmt.Errorf("SMTP error"))

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	err := provider.SendEmail(ctx, "user@example.com", "Test", "Body")
	require.Error(t, err)
	require.Contains(t, err.Error(), "SMTP error")

	emails := provider.GetSentEmails()
	require.Len(t, emails, 0, "Failed sends should not record emails")
}

// TestEmailProviderNetworkErrorInjection tests network error injection.
func TestEmailProviderNetworkErrorInjection(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	provider.InjectNetworkError(3, fmt.Errorf("DNS resolution failed"))

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	// First two calls succeed.
	err := provider.SendEmail(ctx, "user1@example.com", "Subject 1", "Body 1")
	require.NoError(t, err)

	err = provider.SendEmail(ctx, "user2@example.com", "Subject 2", "Body 2")
	require.NoError(t, err)

	// Third call fails.
	err = provider.SendEmail(ctx, "user3@example.com", "Subject 3", "Body 3")
	require.Error(t, err)
	require.Contains(t, err.Error(), "DNS resolution failed")

	emails := provider.GetSentEmails()
	require.Len(t, emails, 2)
	require.Equal(t, 3, provider.GetCallCount())
}

// TestDeliveryServiceInterfaceCompliance tests interface implementation.
func TestDeliveryServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	t.Run("SMS_Provider_SendEmail_Not_Supported", func(t *testing.T) {
		t.Parallel()

		provider := NewSMSProvider()

		err := provider.SendEmail(ctx, "user@example.com", "Subject", "Body")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not supported by SMS provider")
	})

	t.Run("Email_Provider_SendSMS_Not_Supported", func(t *testing.T) {
		t.Parallel()

		provider := NewEmailProvider()

		err := provider.SendSMS(ctx, "+15551234567", "Message")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not supported by email provider")
	})
}

// TestEmailProviderReset tests email provider reset functionality.
func TestEmailProviderReset(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	provider.SetShouldFail(true, nil)
	provider.InjectNetworkError(1, fmt.Errorf("test error"))

	ctx := context.WithValue(context.Background(), contextKeyTimestamp, int64(1234567890))

	_ = provider.SendEmail(ctx, "user@example.com", "Test", "Body") //nolint:errcheck // Test setup - error intentionally ignored to test reset functionality

	provider.Reset()

	// After reset, provider should work normally.
	err := provider.SendEmail(ctx, "user@example.com", "After reset", "Body")
	require.NoError(t, err)

	emails := provider.GetSentEmails()
	require.Len(t, emails, 1, "Reset should clear previous emails")
	require.Equal(t, 1, provider.GetCallCount(), "Reset should clear call count")
}
