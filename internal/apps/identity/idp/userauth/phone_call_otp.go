// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PhoneCallOTPAuthenticator authenticates users via voice call OTP.
type PhoneCallOTPAuthenticator struct {
	generator      OTPGenerator
	phoneService   PhoneCallService
	challengeStore ChallengeStore
	rateLimiter    RateLimiter
	userRepo       cryptoutilIdentityRepository.UserRepository
	otpLength      int
	otpExpiration  time.Duration
	maxRetries     int
}

// PhoneCallService defines interface for voice call delivery.
type PhoneCallService interface {
	MakeVoiceCall(ctx context.Context, phoneNumber, message string) error
}

// NewPhoneCallOTPAuthenticator creates a new phone call OTP authenticator.
func NewPhoneCallOTPAuthenticator(
	generator OTPGenerator,
	phoneService PhoneCallService,
	challengeStore ChallengeStore,
	rateLimiter RateLimiter,
	userRepo cryptoutilIdentityRepository.UserRepository,
) *PhoneCallOTPAuthenticator {
	return &PhoneCallOTPAuthenticator{
		generator:      generator,
		phoneService:   phoneService,
		challengeStore: challengeStore,
		rateLimiter:    rateLimiter,
		userRepo:       userRepo,
		otpLength:      cryptoutilSharedMagic.DefaultOTPLength,
		otpExpiration:  cryptoutilSharedMagic.DefaultPhoneCallOTPTimeout,
		maxRetries:     cryptoutilSharedMagic.DefaultPhoneCallOTPRetries,
	}
}

// Method returns the authentication method name.
func (a *PhoneCallOTPAuthenticator) Method() string {
	return "phone_call_otp"
}

// InitiateAuth initiates phone call OTP authentication for a user.
func (a *PhoneCallOTPAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
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
		Metadata: map[string]any{
			cryptoutilSharedMagic.ScopePhone:       user.PhoneNumber,
			"retry_count": 0,
		},
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

	// Format OTP as spoken digits with pauses.
	spokenOTP := formatOTPForSpeech(otp)
	message := fmt.Sprintf("Your verification code is: %s. I repeat, your verification code is: %s.", spokenOTP, spokenOTP)

	// Make voice call with OTP.
	if err := a.phoneService.MakeVoiceCall(ctx, user.PhoneNumber, message); err != nil {
		return nil, fmt.Errorf("failed to make voice call: %w", err)
	}

	// Record successful attempt (best-effort).
	if err := a.rateLimiter.RecordAttempt(ctx, userID, true); err != nil {
		fmt.Printf("warning: failed to record rate limit attempt: %v\n", err)
	}

	return challenge, nil
}

// VerifyAuth verifies the phone call OTP and returns the authenticated user.
func (a *PhoneCallOTPAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
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
		if err := a.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("OTP expired")
	}

	// Verify OTP against stored hash (constant-time comparison).
	if err := VerifyToken(response, storedHashedOTP); err != nil {
		// Track failed attempt for retry limit.
		retryCount, _ := challenge.Metadata["retry_count"].(int)
		retryCount++

		if retryCount >= a.maxRetries {
			// Max retries exceeded - delete challenge.
			if err := a.challengeStore.Delete(ctx, id); err != nil {
				fmt.Printf("warning: failed to delete challenge after max retries: %v\n", err)
			}

			return nil, fmt.Errorf("maximum retry attempts exceeded")
		}

		// Update retry count.
		challenge.Metadata["retry_count"] = retryCount
		if err := a.challengeStore.Update(ctx, challenge); err != nil {
			fmt.Printf("warning: failed to update challenge retry count: %v\n", err)
		}

		// Best-effort rate limit tracking.
		if err := a.rateLimiter.RecordAttempt(ctx, challenge.UserID, false); err != nil {
			fmt.Printf("warning: failed to record failed attempt: %v\n", err)
		}

		return nil, fmt.Errorf("invalid OTP")
	}

	// Delete challenge (single-use).
	if err := a.challengeStore.Delete(ctx, id); err != nil {
		fmt.Printf("warning: failed to delete challenge after verification: %v\n", err)
	}

	// Fetch and return user.
	user, err := a.userRepo.GetBySub(ctx, challenge.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found after verification: %w", err)
	}

	return user, nil
}

// formatOTPForSpeech formats OTP digits with pauses for voice clarity.
// Example: "123456" becomes "1... 2... 3... 4... 5... 6".
func formatOTPForSpeech(otp string) string {
	result := ""

	for i, digit := range otp {
		if i > 0 {
			result += "... "
		}

		result += string(digit)
	}

	return result
}
