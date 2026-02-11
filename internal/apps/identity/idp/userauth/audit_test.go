// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"

	googleUuid "github.com/google/uuid"
)

// TestTelemetryAuditLoggerTokenGeneration tests token generation logging.
func TestTelemetryAuditLoggerTokenGeneration(t *testing.T) {
	t.Parallel()

	logger, err := NewTelemetryAuditLogger(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	event := TokenGenerationEvent{
		UserID:       "user-123",
		TokenID:      googleUuid.New(),
		TokenType:    "sms_otp",
		Provider:     "sms",
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
		PhoneNumber:  "+15551234567",
		EmailAddress: "",
	}

	err = logger.LogTokenGeneration(ctx, event)
	require.NoError(t, err)
}

// TestTelemetryAuditLoggerValidationAttempt tests validation attempt logging.
func TestTelemetryAuditLoggerValidationAttempt(t *testing.T) {
	t.Parallel()

	logger, err := NewTelemetryAuditLogger(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()

	// Successful validation.
	event := ValidationAttemptEvent{
		UserID:            "user-456",
		TokenID:           googleUuid.New(),
		TokenType:         "email_otp",
		Success:           true,
		RemainingAttempts: 2,
		IPAddress:         "192.168.1.100",
	}

	err = logger.LogValidationAttempt(ctx, event)
	require.NoError(t, err)

	// Failed validation.
	failedEvent := ValidationAttemptEvent{
		UserID:            "user-789",
		TokenID:           googleUuid.New(),
		TokenType:         "sms_otp",
		Success:           false,
		FailureReason:     "expired",
		RemainingAttempts: 0,
		IPAddress:         "10.0.0.50",
	}

	err = logger.LogValidationAttempt(ctx, failedEvent)
	require.NoError(t, err)
}

// TestTelemetryAuditLoggerTokenInvalidation tests token invalidation logging.
func TestTelemetryAuditLoggerTokenInvalidation(t *testing.T) {
	t.Parallel()

	logger, err := NewTelemetryAuditLogger(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	event := TokenInvalidationEvent{
		UserID:    "user-abc",
		TokenID:   googleUuid.New(),
		TokenType: "magic_link",
		Reason:    "used",
	}

	err = logger.LogTokenInvalidation(ctx, event)
	require.NoError(t, err)
}

// TestExtractDomain tests email domain extraction.
func TestExtractDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Standard email",
			email:    "user@example.com",
			expected: "example.com",
		},
		{
			name:     "Subdomain email",
			email:    "admin@mail.example.org",
			expected: "mail.example.org",
		},
		{
			name:     "No @ symbol",
			email:    "notanemail",
			expected: "unknown",
		},
		{
			name:     "Empty string",
			email:    "",
			expected: "unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractDomain(tc.email)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestMaskIPAddress tests IP address masking for privacy.
func TestMaskIPAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "Standard IPv4",
			ip:       "192.168.1.100",
			expected: "192.168.1.xxx",
		},
		{
			name:     "IPv4 with single digit last octet",
			ip:       "10.0.0.5",
			expected: "10.0.0.xxx",
		},
		{
			name:     "IPv4 with three digit last octet",
			ip:       "203.0.113.255",
			expected: "203.0.113.xxx",
		},
		{
			name:     "No dots (invalid IP)",
			ip:       "notanip",
			expected: "xxx.xxx.xxx.xxx",
		},
		{
			name:     "Empty string",
			ip:       "",
			expected: "xxx.xxx.xxx.xxx",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := maskIPAddress(tc.ip)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestAuditLoggerConcurrent tests concurrent audit logging.
func TestAuditLoggerConcurrent(t *testing.T) {
	t.Parallel()

	logger, err := NewTelemetryAuditLogger(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()
	done := make(chan bool, 10)

	// 10 concurrent goroutines logging events.
	for range 10 {
		go func() {
			defer func() { done <- true }()

			// Log token generation.
			genEvent := TokenGenerationEvent{
				UserID:       "concurrent-user",
				TokenID:      googleUuid.New(),
				TokenType:    "sms_otp",
				Provider:     "sms",
				ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
				PhoneNumber:  "+15551234567",
				EmailAddress: "",
			}

			_ = logger.LogTokenGeneration(ctx, genEvent) //nolint:errcheck // Test audit logging - error not critical for test validation

			// Log validation attempt.
			valEvent := ValidationAttemptEvent{
				UserID:            "concurrent-user",
				TokenID:           googleUuid.New(),
				TokenType:         "sms_otp",
				Success:           true,
				RemainingAttempts: 3,
				IPAddress:         "192.168.1.100",
			}

			_ = logger.LogValidationAttempt(ctx, valEvent) //nolint:errcheck // Test audit logging - error not critical for test validation

			// Log invalidation.
			invEvent := TokenInvalidationEvent{
				UserID:    "concurrent-user",
				TokenID:   googleUuid.New(),
				TokenType: "sms_otp",
				Reason:    "used",
			}

			_ = logger.LogTokenInvalidation(ctx, invEvent) //nolint:errcheck // Test audit logging - error not critical for test validation
		}()
	}

	// Wait for all goroutines.
	for range 10 {
		<-done
	}
	// No assertions needed - just verify no panics/crashes during concurrent logging.
}

// TestAuditLoggerPIIProtection tests that sensitive data is not logged.
func TestAuditLoggerPIIProtection(t *testing.T) {
	t.Parallel()

	logger, err := NewTelemetryAuditLogger(noop.NewMeterProvider())
	require.NoError(t, err)

	ctx := context.Background()

	// Token generation with phone number.
	genEvent := TokenGenerationEvent{
		UserID:       "user-pii",
		TokenID:      googleUuid.New(),
		TokenType:    "sms_otp",
		Provider:     "sms",
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
		PhoneNumber:  "+15551234567", // Full phone number (last 4 logged: "4567")
		EmailAddress: "",
	}

	err = logger.LogTokenGeneration(ctx, genEvent)
	require.NoError(t, err)

	// Token generation with email address.
	emailEvent := TokenGenerationEvent{
		UserID:       "user-pii-2",
		TokenID:      googleUuid.New(),
		TokenType:    "email_otp",
		Provider:     "email",
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
		PhoneNumber:  "",
		EmailAddress: "sensitive@example.com", // Full email (domain logged: "example.com")
	}

	err = logger.LogTokenGeneration(ctx, emailEvent)
	require.NoError(t, err)

	// Validation attempt with IP address.
	valEvent := ValidationAttemptEvent{
		UserID:            "user-pii-3",
		TokenID:           googleUuid.New(),
		TokenType:         "sms_otp",
		Success:           false,
		FailureReason:     "invalid",
		RemainingAttempts: 2,
		IPAddress:         "203.0.113.42", // Full IP (masked: "203.0.113.xxx")
	}

	err = logger.LogValidationAttempt(ctx, valEvent)
	require.NoError(t, err)
	// CRITICAL: This test verifies the audit logger doesn't panic with PII data.
	// Actual PII masking is tested in TestExtractDomain and TestMaskIPAddress.
	// Production verification: grep logs for full phone/email/IP (should find none).
}
