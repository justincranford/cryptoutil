// Copyright (c) 2025 Justin Cranford

package email_test

import (
	"context"
	"testing"

	cryptoutilIdentityEmail "cryptoutil/internal/apps/identity/email"

	"github.com/stretchr/testify/require"
)

func TestMockEmailService_SendEmail(t *testing.T) {
	t.Parallel()

	mockService := cryptoutilIdentityEmail.NewMockEmailService()
	ctx := context.Background()

	err := mockService.SendEmail(ctx, "user@example.com", "Test Subject", "Test Body")
	require.NoError(t, err)

	require.Len(t, mockService.SentEmails, 1, "Should have 1 sent email")
	lastEmail := mockService.GetLastEmail()
	require.NotNil(t, lastEmail)
	require.Equal(t, "user@example.com", lastEmail.To)
	require.Equal(t, "Test Subject", lastEmail.Subject)
	require.Equal(t, "Test Body", lastEmail.Body)
}

func TestMockEmailService_GetLastEmail(t *testing.T) {
	t.Parallel()

	mockService := cryptoutilIdentityEmail.NewMockEmailService()

	// No emails sent yet.
	lastEmail := mockService.GetLastEmail()
	require.Nil(t, lastEmail, "Should return nil when no emails sent")

	// Send first email.
	ctx := context.Background()
	err := mockService.SendEmail(ctx, "user1@example.com", "Subject 1", "Body 1")
	require.NoError(t, err)

	lastEmail = mockService.GetLastEmail()
	require.NotNil(t, lastEmail)
	require.Equal(t, "user1@example.com", lastEmail.To)

	// Send second email.
	err = mockService.SendEmail(ctx, "user2@example.com", "Subject 2", "Body 2")
	require.NoError(t, err)

	lastEmail = mockService.GetLastEmail()
	require.NotNil(t, lastEmail)
	require.Equal(t, "user2@example.com", lastEmail.To, "Should return most recent email")
}

func TestMockEmailService_ContainsOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		wantOTP  string
		wantBool bool
	}{
		{
			name:     "valid_otp_in_body",
			body:     "Your OTP code is: 123456",
			wantOTP:  "123456",
			wantBool: true,
		},
		{
			name:     "otp_at_start",
			body:     "123456 is your verification code",
			wantOTP:  "123456",
			wantBool: true,
		},
		{
			name:     "no_otp",
			body:     "This email has no OTP",
			wantOTP:  "",
			wantBool: false,
		},
		{
			name:     "partial_otp",
			body:     "Code: 12345",
			wantOTP:  "",
			wantBool: false,
		},
		{
			name:     "non_numeric",
			body:     "Code: ABCDEF",
			wantOTP:  "",
			wantBool: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockService := cryptoutilIdentityEmail.NewMockEmailService()
			email := &cryptoutilIdentityEmail.SentEmail{
				To:      "user@example.com",
				Subject: "Test",
				Body:    tc.body,
			}

			otp, found := mockService.ContainsOTP(email)
			require.Equal(t, tc.wantOTP, otp)
			require.Equal(t, tc.wantBool, found)
		})
	}
}
