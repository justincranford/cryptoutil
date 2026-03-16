// Copyright (c) 2025 Justin Cranford

package mocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSMSProvider_EmptyPhone(t *testing.T) {
	t.Parallel()

	provider := NewSMSProvider()
	err := provider.SendSMS(context.Background(), "", "message")
	require.Error(t, err)
	require.ErrorContains(t, err, "phone number cannot be empty")
}

func TestSMSProvider_EmptyMessage(t *testing.T) {
	t.Parallel()

	provider := NewSMSProvider()
	err := provider.SendSMS(context.Background(), "+15551234567", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "message cannot be empty")
}

func TestEmailProvider_EmptyRecipient(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	err := provider.SendEmail(context.Background(), "", "Subject", "Body")
	require.Error(t, err)
	require.ErrorContains(t, err, "recipient cannot be empty")
}

func TestEmailProvider_EmptySubject(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	err := provider.SendEmail(context.Background(), "user@example.com", "", "Body")
	require.Error(t, err)
	require.ErrorContains(t, err, "subject cannot be empty")
}

func TestEmailProvider_EmptyBody(t *testing.T) {
	t.Parallel()

	provider := NewEmailProvider()
	err := provider.SendEmail(context.Background(), "user@example.com", "Subject", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "body cannot be empty")
}
