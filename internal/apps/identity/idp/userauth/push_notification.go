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

// PushNotificationAuthenticator authenticates users via mobile push notifications.
type PushNotificationAuthenticator struct {
	generator      OTPGenerator
	pushService    PushNotificationService
	challengeStore ChallengeStore
	rateLimiter    RateLimiter
	userRepo       cryptoutilIdentityRepository.UserRepository
	pushTimeout    time.Duration
}

// PushNotificationService defines interface for push notification delivery.
type PushNotificationService interface {
	SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]any) error
}

// NewPushNotificationAuthenticator creates a new push notification authenticator.
func NewPushNotificationAuthenticator(
	generator OTPGenerator,
	pushService PushNotificationService,
	challengeStore ChallengeStore,
	rateLimiter RateLimiter,
	userRepo cryptoutilIdentityRepository.UserRepository,
) *PushNotificationAuthenticator {
	return &PushNotificationAuthenticator{
		generator:      generator,
		pushService:    pushService,
		challengeStore: challengeStore,
		rateLimiter:    rateLimiter,
		userRepo:       userRepo,
		pushTimeout:    cryptoutilSharedMagic.DefaultPushNotificationTimeout,
	}
}

// Method returns the authentication method name.
func (a *PushNotificationAuthenticator) Method() string {
	return "push_notification"
}

// InitiateAuth initiates push notification authentication for a user.
func (a *PushNotificationAuthenticator) InitiateAuth(ctx context.Context, userID string) (*AuthChallenge, error) {
	// Check rate limit.
	if err := a.rateLimiter.CheckLimit(ctx, userID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Fetch user to get device token.
	user, err := a.userRepo.GetBySub(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.PushDeviceToken == "" {
		return nil, fmt.Errorf("user has no push device token configured")
	}

	// Create challenge.
	challenge := &AuthChallenge{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		UserID:    userID,
		Method:    a.Method(),
		ExpiresAt: time.Now().UTC().Add(a.pushTimeout),
		Metadata:  map[string]any{"device_token": user.PushDeviceToken},
	}

	// Generate approval token.
	approvalToken, err := a.generator.GenerateSecureToken(cryptoutilSharedMagic.DefaultPushNotificationTokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate approval token: %w", err)
	}

	// Hash approval token before storage.
	hashedToken, err := HashToken(approvalToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash approval token: %w", err)
	}

	// Store challenge with hashed approval token.
	if err := a.challengeStore.Store(ctx, challenge, hashedToken); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	// Send push notification with approval data.
	title := "Authentication Request"
	body := "Tap to approve login request"
	data := map[string]any{
		"challenge_id":   challenge.ID.String(),
		"approval_token": approvalToken,
		"expires_at":     challenge.ExpiresAt.Unix(),
	}

	if err := a.pushService.SendPushNotification(ctx, user.PushDeviceToken, title, body, data); err != nil {
		return nil, fmt.Errorf("failed to send push notification: %w", err)
	}

	// Record successful attempt.
	if err := a.rateLimiter.RecordAttempt(ctx, userID, true); err != nil {
		fmt.Printf("warning: failed to record rate limit attempt: %v\n", err)
	}

	return challenge, nil
}

// VerifyAuth verifies the push notification approval and returns the authenticated user.
func (a *PushNotificationAuthenticator) VerifyAuth(ctx context.Context, challengeID, response string) (*cryptoutilIdentityDomain.User, error) {
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
		if err := a.challengeStore.Delete(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete expired challenge: %v\n", err)
		}

		return nil, fmt.Errorf("push notification expired")
	}

	// Verify approval token against stored hash.
	if err := VerifyToken(response, storedHashedToken); err != nil {
		if err := a.rateLimiter.RecordAttempt(ctx, challenge.UserID, false); err != nil {
			fmt.Printf("warning: failed to record failed attempt: %v\n", err)
		}

		return nil, fmt.Errorf("invalid approval token")
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
