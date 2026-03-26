// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// SMSOTPAuthenticator authenticates users via SMS one-time passwords.
type SMSOTPAuthenticator struct {
	generator      OTPGenerator
	delivery       DeliveryService
	challengeStore ChallengeStore
	rateLimiter    RateLimiter
	userRepo       cryptoutilIdentityRepository.UserRepository
	otpLength      int
	otpExpiration  time.Duration
}

// NewSMSOTPAuthenticator creates a new SMS OTP authenticator.
func NewSMSOTPAuthenticator(
	generator OTPGenerator,
	delivery DeliveryService,
	challengeStore ChallengeStore,
	rateLimiter RateLimiter,
	userRepo cryptoutilIdentityRepository.UserRepository,
) *SMSOTPAuthenticator {
	return &SMSOTPAuthenticator{
		generator:      generator,
		delivery:       delivery,
		challengeStore: challengeStore,
		rateLimiter:    rateLimiter,
		userRepo:       userRepo,
		otpLength:      cryptoutilSharedMagic.DefaultOTPLength,
		otpExpiration:  cryptoutilSharedMagic.DefaultOTPLifetime,
	}
}

// Method returns the authentication method name.
func (a *SMSOTPAuthenticator) Method() string {
	return cryptoutilSharedMagic.AuthMethodSMSOTP
}

// InitiateAuth initiates SMS OTP authentication for a user.
func (a *SMSOTPAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Check rate limit.
	if err := a.rateLimiter.CheckLimit(ctx, userID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Fetch user to get phone number.
	user, err := a.userRepo.GetBySub(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.PhoneNumber == "" {
		return nil, fmt.Errorf("user has no phone number configured")
	}

	// Generate OTP.
	otp, err := a.generator.GenerateOTP(a.otpLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Create challenge.
	challenge := &AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		Method:    a.Method(),
		ExpiresAt: time.Now().UTC().Add(a.otpExpiration),
		Metadata:  map[string]any{cryptoutilSharedMagic.ScopePhone: user.PhoneNumber},
	}

	// Hash OTP before storage (SECURITY: never store plaintext tokens).
	hashedOTP, err := HashToken(otp)
	if err != nil {
		return nil, fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Store challenge with hashed OTP.
	if err := a.challengeStore.Store(ctx, challenge, hashedOTP); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	// Send SMS with OTP.
	message := fmt.Sprintf("Your verification code is: %s (expires in %d minutes)",
		otp, int(a.otpExpiration.Minutes()))
	if err := a.delivery.SendSMS(ctx, user.PhoneNumber, message); err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	// Record successful attempt (best-effort, don't fail if rate limiter has issues).
	if err := a.rateLimiter.RecordAttempt(ctx, userID, true); err != nil {
		// Log but don't fail - rate limiting is supplementary.
		fmt.Printf("warning: failed to record rate limit attempt: %v\n", err)
	}

	return challenge, nil
}

// VerifyAuth verifies the SMS OTP and returns the authenticated user.
func (a *SMSOTPAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, storedHashedOTP, err := a.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().UTC().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := a.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("oTP expired")
	}

	// Verify OTP against stored hash (constant-time comparison).
	if err := VerifyToken(response, storedHashedOTP); err != nil {
		// Best-effort rate limit tracking.
		if err := a.rateLimiter.RecordAttempt(ctx, challenge.UserID, false); err != nil {
			fmt.Printf("warning: failed to record failed attempt: %v\n", err)
		}

		return nil, fmt.Errorf("invalid OTP")
	}

	// Delete challenge (single-use).
	if err := a.challengeStore.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete challenge: %w", err)
	}

	// Fetch and return user.
	user, err := a.userRepo.GetBySub(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// DefaultOTPGenerator implements OTPGenerator with crypto/rand.
type DefaultOTPGenerator struct{}

// GenerateOTP generates a random numeric OTP of specified length.
func (g *DefaultOTPGenerator) GenerateOTP(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid OTP length: %d", length)
	}

	otp := make([]byte, length)
	maxDigit := big.NewInt(cryptoutilSharedMagic.DecimalRadix)

	for i := 0; i < length; i++ {
		digit, err := crand.Int(crand.Reader, maxDigit)
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}

		otp[i] = byte('0' + digit.Int64())
	}

	return string(otp), nil
}

// GenerateSecureToken generates a secure random token.
func (g *DefaultOTPGenerator) GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid token length: %d", length)
	}

	// Generate random bytes.
	bytes := make([]byte, length)
	if _, err := crand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as hex string.
	const hexChars = "0123456789abcdef"

	token := make([]byte, length*2)
	for i, b := range bytes {
		token[i*2] = hexChars[b>>4]
		token[i*2+1] = hexChars[b&0x0f]
	}

	return string(token), nil
}
