// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// MagicLinkAuthenticator authenticates users via email magic links.
type MagicLinkAuthenticator struct {
	generator      OTPGenerator
	delivery       DeliveryService
	challengeStore ChallengeStore
	rateLimiter    RateLimiter
	userRepo       cryptoutilIdentityRepository.UserRepository
	baseURL        string
	tokenLength    int
	linkExpiration time.Duration
}

// NewMagicLinkAuthenticator creates a new magic link authenticator.
func NewMagicLinkAuthenticator(
	generator OTPGenerator,
	delivery DeliveryService,
	challengeStore ChallengeStore,
	rateLimiter RateLimiter,
	userRepo cryptoutilIdentityRepository.UserRepository,
	baseURL string,
) *MagicLinkAuthenticator {
	return &MagicLinkAuthenticator{
		generator:      generator,
		delivery:       delivery,
		challengeStore: challengeStore,
		rateLimiter:    rateLimiter,
		userRepo:       userRepo,
		baseURL:        baseURL,
		tokenLength:    cryptoutilIdentityMagic.DefaultMagicLinkLength,
		linkExpiration: cryptoutilIdentityMagic.DefaultMagicLinkLifetime,
	}
}

// Method returns the authentication method name.
func (a *MagicLinkAuthenticator) Method() string {
	return "magic_link"
}

// InitiateAuth initiates magic link authentication for a user.
func (a *MagicLinkAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Check rate limit.
	if err := a.rateLimiter.CheckLimit(ctx, userID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Fetch user to get email.
	user, err := a.userRepo.GetBySub(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Email == "" {
		return nil, fmt.Errorf("user has no email configured")
	}

	// Generate secure token.
	token, err := a.generator.GenerateSecureToken(a.tokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create challenge.
	challenge := &AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		Method:    a.Method(),
		ExpiresAt: time.Now().UTC().Add(a.linkExpiration),
		Metadata:  map[string]any{"email": user.Email},
	}

	// Hash token before storage (SECURITY: never store plaintext tokens).
	hashedToken, err := HashToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	// Store challenge with hashed token.
	if err := a.challengeStore.Store(ctx, challenge, hashedToken); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	// Create magic link URL.
	magicLink := fmt.Sprintf("%s/auth/magic-link/verify?token=%s&challenge=%s",
		a.baseURL, token, challenge.ID.String())

	// Send email with magic link.
	subject := "Your Magic Link"
	body := fmt.Sprintf(`Click the link below to sign in (expires in %d minutes):

%s

If you didn't request this link, please ignore this email.`,
		int(a.linkExpiration.Minutes()), magicLink)

	if err := a.delivery.SendEmail(ctx, user.Email, subject, body); err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	// Record successful attempt (best-effort, don't fail if rate limiter has issues).
	if err := a.rateLimiter.RecordAttempt(ctx, userID, true); err != nil {
		// Log but don't fail - rate limiting is supplementary.
		fmt.Printf("warning: failed to record rate limit attempt: %v\n", err)
	}

	return challenge, nil
}

// VerifyAuth verifies the magic link token and returns the authenticated user.
func (a *MagicLinkAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
	// Parse challenge ID.
	id, err := googleUuid.Parse(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	// Retrieve challenge.
	challenge, storedHashedToken, err := a.challengeStore.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// Check expiration.
	if time.Now().UTC().After(challenge.ExpiresAt) {
		// Best-effort cleanup of expired challenge.
		if err := a.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("magic link expired")
	}

	// Verify token against stored hash (constant-time comparison).
	if err := VerifyToken(response, storedHashedToken); err != nil {
		// Best-effort rate limit tracking.
		if err := a.rateLimiter.RecordAttempt(ctx, challenge.UserID, false); err != nil {
			fmt.Printf("warning: failed to record failed attempt: %v\n", err)
		}

		return nil, fmt.Errorf("invalid magic link token")
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
