// Copyright (c) 2025 Justin Cranford

package email_test

import (
	"context"
	"testing"

	cryptoutilIdentityEmail "cryptoutil/internal/apps/identity/email"

	"github.com/stretchr/testify/require"
)

// TestNewSMTPEmailService tests SMTP email service construction.
func TestNewSMTPEmailService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config cryptoutilIdentityEmail.SMTPConfig
	}{
		{
			name: "valid_config",
			config: cryptoutilIdentityEmail.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "password123",
				From:     "noreply@example.com",
			},
		},
		{
			name: "empty_config",
			config: cryptoutilIdentityEmail.SMTPConfig{
				Host:     "",
				Port:     0,
				Username: "",
				Password: "",
				From:     "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := cryptoutilIdentityEmail.NewSMTPEmailService(tc.config)
			require.NotNil(t, service, "NewSMTPEmailService should return non-nil service")
		})
	}
}

// TestSMTPEmailService_SendEmail_ErrorPaths tests SMTP SendEmail error handling.
//
// NOTE: Success path tests removed - require live SMTP server or complex mock of smtp.SendMail.
// smtp.SendMail is a standard library function that directly opens network connections.
// Mocking requires either:
// 1. Live SMTP server (external dependency, not suitable for unit tests)
// 2. Interface wrapper + mock (adds complexity for minimal coverage value)
// 3. Network interception (fragile, platform-specific)
//
// Coverage: NewSMTPEmailService 100% (constructor always succeeds), SendEmail 0% (requires network).
func TestSMTPEmailService_SendEmail_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config cryptoutilIdentityEmail.SMTPConfig
		to     string
	}{
		{
			name: "invalid_smtp_host",
			config: cryptoutilIdentityEmail.SMTPConfig{
				Host:     "invalid.smtp.host.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "password123",
				From:     "noreply@example.com",
			},
			to: "recipient@example.com",
		},
		{
			name: "invalid_smtp_port",
			config: cryptoutilIdentityEmail.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     99999, // Invalid port
				Username: "user@example.com",
				Password: "password123",
				From:     "noreply@example.com",
			},
			to: "recipient@example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := cryptoutilIdentityEmail.NewSMTPEmailService(tc.config)
			ctx := context.Background()

			err := service.SendEmail(ctx, tc.to, "Test Subject", "Test Body")

			// Expect error due to invalid SMTP configuration.
			// Network errors may vary (dial, connection refused, timeout), so just check error exists.
			require.Error(t, err, "SendEmail should fail with invalid SMTP config")
		})
	}
}

// TestMockEmailService_ContainsOTP_NilEmail tests ContainsOTP with nil email.
func TestMockEmailService_ContainsOTP_NilEmail(t *testing.T) {
	t.Parallel()

	mockService := cryptoutilIdentityEmail.NewMockEmailService()

	otp, found := mockService.ContainsOTP(nil)
	require.Empty(t, otp, "OTP should be empty for nil email")
	require.False(t, found, "found should be false for nil email")
}

// TestMockEmailService_ContainsOTP_EdgeCases tests ContainsOTP with additional edge cases.
func TestMockEmailService_ContainsOTP_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		wantOTP  string
		wantBool bool
	}{
		{
			name:     "empty_body",
			body:     "",
			wantOTP:  "",
			wantBool: false,
		},
		{
			name:     "seven_digit_number",
			body:     "Code: 1234567",
			wantOTP:  "",
			wantBool: false,
		},
		{
			name:     "five_digit_number",
			body:     "Code: 12345",
			wantOTP:  "",
			wantBool: false,
		},
		{
			name:     "alphanumeric_six_chars",
			body:     "Code: 12A45B",
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
			require.Equal(t, tc.wantOTP, otp, "OTP mismatch")
			require.Equal(t, tc.wantBool, found, "found mismatch")
		})
	}
}
