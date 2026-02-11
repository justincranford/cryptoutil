// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// UserAuthenticator defines the interface for user authentication methods.
type UserAuthenticator interface {
	// Method returns the authentication method name.
	Method() string

	// InitiateAuth initiates authentication for a user and returns a challenge.
	InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error)

	// VerifyAuth verifies the authentication response and returns the authenticated user.
	VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error)
}

// AuthChallenge represents an authentication challenge sent to the user.
type AuthChallenge struct {
	ID        googleUuid.UUID // Unique challenge identifier.
	UserID    string          // User identifier.
	Method    string          // Authentication method.
	ExpiresAt time.Time       // Challenge expiration time.
	Metadata  map[string]any  // Additional challenge metadata.
}

// OTPGenerator defines the interface for generating one-time passwords.
type OTPGenerator interface {
	// GenerateOTP generates a numeric OTP of specified length.
	GenerateOTP(length int) (string, error)

	// GenerateSecureToken generates a secure random token.
	GenerateSecureToken(length int) (string, error)
}

// DeliveryService defines the interface for delivering authentication messages.
type DeliveryService interface {
	// SendSMS sends an SMS message to a phone number.
	SendSMS(ctx context.Context, phoneNumber, message string) error

	// SendEmail sends an email message.
	SendEmail(ctx context.Context, to, subject, body string) error
}

// ChallengeStore defines the interface for storing authentication challenges.
type ChallengeStore interface {
	// Store stores an authentication challenge.
	Store(ctx context.Context, challenge *AuthChallenge, secret string) error

	// Retrieve retrieves an authentication challenge and its secret.
	Retrieve(ctx context.Context, challengeID googleUuid.UUID) (*AuthChallenge, string, error)

	// Update updates an existing authentication challenge (e.g., retry count).
	Update(ctx context.Context, challenge *AuthChallenge) error

	// Delete deletes an authentication challenge.
	Delete(ctx context.Context, challengeID googleUuid.UUID) error
}

// RateLimiter defines the interface for rate limiting authentication attempts.
type RateLimiter interface {
	// CheckLimit checks if a rate limit has been exceeded.
	CheckLimit(ctx context.Context, identifier string) error

	// RecordAttempt records an authentication attempt.
	RecordAttempt(ctx context.Context, identifier string, success bool) error
}
