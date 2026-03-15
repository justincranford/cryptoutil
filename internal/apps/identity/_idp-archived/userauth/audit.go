// Copyright (c) 2025 Justin Cranford
//
//

// Package userauth provides user authentication and authorization services.
package userauth

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// AuditLogger defines interface for authentication audit logging.
type AuditLogger interface {
	// LogTokenGeneration logs OTP token generation events.
	LogTokenGeneration(ctx context.Context, event TokenGenerationEvent) error

	// LogValidationAttempt logs token validation attempts.
	LogValidationAttempt(ctx context.Context, event ValidationAttemptEvent) error

	// LogTokenInvalidation logs token invalidation events.
	LogTokenInvalidation(ctx context.Context, event TokenInvalidationEvent) error
}

// TokenGenerationEvent represents a token generation audit event.
type TokenGenerationEvent struct {
	UserID       string
	TokenID      googleUuid.UUID // Token identifier (NOT the token value)
	TokenType    string          // "sms_otp", "email_otp", "magic_link"
	Provider     string          // "sms", "email"
	ExpiresAt    time.Time
	PhoneNumber  string // For SMS (last 4 digits only in logs)
	EmailAddress string // For email (domain only in logs)
}

// ValidationAttemptEvent represents a token validation audit event.
type ValidationAttemptEvent struct {
	UserID            string
	TokenID           googleUuid.UUID
	TokenType         string
	Success           bool
	FailureReason     string // "expired", "invalid", "rate_limited", "not_found"
	RemainingAttempts int
	IPAddress         string
}

// TokenInvalidationEvent represents a token invalidation audit event.
type TokenInvalidationEvent struct {
	UserID    string
	TokenID   googleUuid.UUID
	TokenType string
	Reason    string // "used", "expired", "manual_revoke", "security_incident"
}

// TelemetryAuditLogger implements AuditLogger with OpenTelemetry metrics.
type TelemetryAuditLogger struct {
	meterProvider       metric.MeterProvider
	generationCounter   metric.Int64Counter
	validationCounter   metric.Int64Counter
	invalidationCounter metric.Int64Counter
}

// NewTelemetryAuditLogger creates a new telemetry-based audit logger.
func NewTelemetryAuditLogger(meterProvider metric.MeterProvider) (*TelemetryAuditLogger, error) {
	meter := meterProvider.Meter("identity.audit")

	generationCounter, err := meter.Int64Counter(
		"identity.audit.token.generated",
		metric.WithDescription("Total OTP tokens generated"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create generation counter: %w", err)
	}

	validationCounter, err := meter.Int64Counter(
		"identity.audit.token.validated",
		metric.WithDescription("Total token validation attempts"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation counter: %w", err)
	}

	invalidationCounter, err := meter.Int64Counter(
		"identity.audit.token.invalidated",
		metric.WithDescription("Total tokens invalidated"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create invalidation counter: %w", err)
	}

	return &TelemetryAuditLogger{
		meterProvider:       meterProvider,
		generationCounter:   generationCounter,
		validationCounter:   validationCounter,
		invalidationCounter: invalidationCounter,
	}, nil
}

// LogTokenGeneration logs a token generation event.
func (l *TelemetryAuditLogger) LogTokenGeneration(ctx context.Context, event TokenGenerationEvent) error {
	// CRITICAL: Log token ID only, NEVER the token value.
	attrs := []attribute.KeyValue{
		attribute.String("user_id", event.UserID),
		attribute.String("token_id", event.TokenID.String()),
		attribute.String(cryptoutilSharedMagic.ParamTokenType, event.TokenType),
		attribute.String("provider", event.Provider),
		attribute.Int64("expires_at", event.ExpiresAt.Unix()),
	}

	// For SMS: log last 4 digits of phone only (PII protection).
	if event.PhoneNumber != "" && len(event.PhoneNumber) >= 4 {
		attrs = append(attrs, attribute.String("phone_last4", event.PhoneNumber[len(event.PhoneNumber)-4:]))
	}

	// For email: log domain only (PII protection).
	if event.EmailAddress != "" {
		domain := extractDomain(event.EmailAddress)
		attrs = append(attrs, attribute.String("email_domain", domain))
	}

	l.generationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	return nil
}

// LogValidationAttempt logs a token validation attempt event.
func (l *TelemetryAuditLogger) LogValidationAttempt(ctx context.Context, event ValidationAttemptEvent) error {
	// CRITICAL: Log token ID only, NEVER the token value.
	attrs := []attribute.KeyValue{
		attribute.String("user_id", event.UserID),
		attribute.String("token_id", event.TokenID.String()),
		attribute.String(cryptoutilSharedMagic.ParamTokenType, event.TokenType),
		attribute.Bool("success", event.Success),
		attribute.Int("remaining_attempts", event.RemainingAttempts),
	}

	if !event.Success && event.FailureReason != "" {
		attrs = append(attrs, attribute.String("failure_reason", event.FailureReason))
	}

	// Log IP address for security monitoring (but mask last octet for privacy).
	if event.IPAddress != "" {
		maskedIP := maskIPAddress(event.IPAddress)
		attrs = append(attrs, attribute.String("ip_address", maskedIP))
	}

	l.validationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	return nil
}

// LogTokenInvalidation logs a token invalidation event.
func (l *TelemetryAuditLogger) LogTokenInvalidation(ctx context.Context, event TokenInvalidationEvent) error {
	// CRITICAL: Log token ID only, NEVER the token value.
	attrs := []attribute.KeyValue{
		attribute.String("user_id", event.UserID),
		attribute.String("token_id", event.TokenID.String()),
		attribute.String(cryptoutilSharedMagic.ParamTokenType, event.TokenType),
		attribute.String("reason", event.Reason),
	}

	l.invalidationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	return nil
}

// extractDomain extracts the domain from an email address.
func extractDomain(email string) string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return email[i+1:]
		}
	}

	return "unknown"
}

// maskIPAddress masks the last octet of an IPv4 address for privacy.
// Example: 192.168.1.100 → 192.168.1.xxx.
func maskIPAddress(ip string) string {
	lastDot := -1

	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == '.' {
			lastDot = i

			break
		}
	}

	if lastDot > 0 {
		return ip[:lastDot+1] + "xxx"
	}

	return "xxx.xxx.xxx.xxx" // Fallback if no dots found
}
