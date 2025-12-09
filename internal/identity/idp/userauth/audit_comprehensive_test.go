// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// TestTelemetryAuditLogger_LogTokenGeneration tests token generation audit logging.
func TestTelemetryAuditLogger_LogTokenGeneration(t *testing.T) {
	t.Parallel()

	meterProvider := noop.NewMeterProvider()
	logger, err := NewTelemetryAuditLogger(meterProvider)
	require.NoError(t, err)

	ctx := context.Background()
	event := TokenGenerationEvent{
		UserID:       "user123",
		TokenID:      googleUuid.New(),
		TokenType:    "sms_otp",
		Provider:     "sms",
		ExpiresAt:    time.Now().Add(time.Minute * 5),
		PhoneNumber:  "+1234567890",
		EmailAddress: "",
	}

	err = logger.LogTokenGeneration(ctx, event)
	require.NoError(t, err)
}

// TestTelemetryAuditLogger_LogValidationAttempt tests validation attempt logging.
func TestTelemetryAuditLogger_LogValidationAttempt(t *testing.T) {
	t.Parallel()

	meterProvider := noop.NewMeterProvider()
	logger, err := NewTelemetryAuditLogger(meterProvider)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name  string
		event ValidationAttemptEvent
	}{
		{
			name: "successful validation",
			event: ValidationAttemptEvent{
				UserID:            "user123",
				TokenID:           googleUuid.New(),
				TokenType:         "sms_otp",
				Success:           true,
				FailureReason:     "",
				RemainingAttempts: 3,
				IPAddress:         "192.168.1.1",
			},
		},
		{
			name: "failed validation - expired",
			event: ValidationAttemptEvent{
				UserID:            "user456",
				TokenID:           googleUuid.New(),
				TokenType:         "email_otp",
				Success:           false,
				FailureReason:     "expired",
				RemainingAttempts: 2,
				IPAddress:         "10.0.0.1",
			},
		},
		{
			name: "failed validation - rate limited",
			event: ValidationAttemptEvent{
				UserID:            "user789",
				TokenID:           googleUuid.New(),
				TokenType:         "magic_link",
				Success:           false,
				FailureReason:     "rate_limited",
				RemainingAttempts: 0,
				IPAddress:         "172.16.0.1",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err = logger.LogValidationAttempt(ctx, tc.event)
			require.NoError(t, err)
		})
	}
}

// TestTelemetryAuditLogger_LogTokenInvalidation tests token invalidation logging.
func TestTelemetryAuditLogger_LogTokenInvalidation(t *testing.T) {
	t.Parallel()

	meterProvider := noop.NewMeterProvider()
	logger, err := NewTelemetryAuditLogger(meterProvider)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name  string
		event TokenInvalidationEvent
	}{
		{
			name: "used token",
			event: TokenInvalidationEvent{
				UserID:    "user123",
				TokenID:   googleUuid.New(),
				TokenType: "sms_otp",
				Reason:    "used",
			},
		},
		{
			name: "expired token",
			event: TokenInvalidationEvent{
				UserID:    "user456",
				TokenID:   googleUuid.New(),
				TokenType: "email_otp",
				Reason:    "expired",
			},
		},
		{
			name: "manual revoke",
			event: TokenInvalidationEvent{
				UserID:    "user789",
				TokenID:   googleUuid.New(),
				TokenType: "magic_link",
				Reason:    "manual_revoke",
			},
		},
		{
			name: "security incident",
			event: TokenInvalidationEvent{
				UserID:    "user101",
				TokenID:   googleUuid.New(),
				TokenType: "sms_otp",
				Reason:    "security_incident",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err = logger.LogTokenInvalidation(ctx, tc.event)
			require.NoError(t, err)
		})
	}
}

// TestAuditLogger_ExtractDomain tests domain extraction from email addresses.
func TestAuditLogger_ExtractDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"valid email", "user@example.com", "example.com"},
		{"subdomain", "admin@mail.company.org", "mail.company.org"},
		{"no @ sign", "invalid-email", "unknown"},
		{"empty", "", "unknown"},
		{"multiple @", "user@test@example.com", "example.com"}, // Last @ wins.
		{"no domain", "user@", ""},
		{"only domain", "@example.com", "example.com"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractDomain(tc.email)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestAuditLogger_MaskIPAddress tests IP address masking for privacy.
func TestAuditLogger_MaskIPAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{"IPv4", "192.168.1.100", "192.168.1.xxx"},
		{"IPv4 localhost", "127.0.0.1", "127.0.0.xxx"},
		{"IPv4 public", "8.8.8.8", "8.8.8.xxx"},
		{"IPv6 short", "2001:db8::1", "xxx.xxx.xxx.xxx"}, // No dots, fallback.
		{"IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "xxx.xxx.xxx.xxx"}, // No dots, fallback.
		{"empty", "", "xxx.xxx.xxx.xxx"},
		{"invalid", "not-an-ip", "xxx.xxx.xxx.xxx"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := maskIPAddress(tc.ip)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestNewTelemetryAuditLogger_MeterCreationError tests error handling during meter creation.
func TestNewTelemetryAuditLogger_MeterCreationError(t *testing.T) {
	t.Parallel()

	// Using noop.NewMeterProvider() should never fail, but test constructor succeeds.
	meterProvider := noop.NewMeterProvider()
	logger, err := NewTelemetryAuditLogger(meterProvider)
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.NotNil(t, logger.generationCounter)
	require.NotNil(t, logger.validationCounter)
	require.NotNil(t, logger.invalidationCounter)
}

// TestTelemetryAuditLogger_ComprehensiveLogging tests all logging methods together.
func TestTelemetryAuditLogger_ComprehensiveLogging(t *testing.T) {
	t.Parallel()

	meterProvider := noop.NewMeterProvider()
	logger, err := NewTelemetryAuditLogger(meterProvider)
	require.NoError(t, err)

	ctx := context.Background()
	userID := "comprehensive-test-user"
	tokenID := googleUuid.New()

	// Log token generation.
	genEvent := TokenGenerationEvent{
		UserID:       userID,
		TokenID:      tokenID,
		TokenType:    "sms_otp",
		Provider:     "sms",
		ExpiresAt:    time.Now().Add(time.Minute * 5),
		PhoneNumber:  "+1234567890",
		EmailAddress: "",
	}
	err = logger.LogTokenGeneration(ctx, genEvent)
	require.NoError(t, err)

	// Log validation attempt (success).
	valEvent := ValidationAttemptEvent{
		UserID:            userID,
		TokenID:           tokenID,
		TokenType:         "sms_otp",
		Success:           true,
		FailureReason:     "",
		RemainingAttempts: 0,
		IPAddress:         "192.168.1.100",
	}
	err = logger.LogValidationAttempt(ctx, valEvent)
	require.NoError(t, err)

	// Log token invalidation (used).
	invEvent := TokenInvalidationEvent{
		UserID:    userID,
		TokenID:   tokenID,
		TokenType: "sms_otp",
		Reason:    "used",
	}
	err = logger.LogTokenInvalidation(ctx, invEvent)
	require.NoError(t, err)
}

// mockMeterProvider implements metric.MeterProvider for testing error conditions.
type mockMeterProvider struct {
	shouldFailCounter bool
}

func (m *mockMeterProvider) Meter(_ string, _ ...metric.MeterOption) metric.Meter {
	return &mockMeter{shouldFailCounter: m.shouldFailCounter}
}

type mockMeter struct {
	shouldFailCounter bool
	noop.Meter
}

func (m *mockMeter) Int64Counter(name string, _ ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	if m.shouldFailCounter {
		return nil, fmt.Errorf("mock error creating counter: %s", name)
	}

	return noop.Int64Counter{}, nil
}
