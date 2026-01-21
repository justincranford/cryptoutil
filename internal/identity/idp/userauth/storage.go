// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// InMemoryChallengeStore implements ChallengeStore using in-memory storage.
type InMemoryChallengeStore struct {
	mu         sync.RWMutex
	challenges map[googleUuid.UUID]*storedChallenge
}

type storedChallenge struct {
	challenge *AuthChallenge
	secret    string
	expiresAt time.Time
}

// NewInMemoryChallengeStore creates a new in-memory challenge store.
func NewInMemoryChallengeStore() *InMemoryChallengeStore {
	store := &InMemoryChallengeStore{
		challenges: make(map[googleUuid.UUID]*storedChallenge),
	}

	// Start cleanup goroutine.
	go store.cleanup()

	return store
}

// Store stores an authentication challenge.
func (s *InMemoryChallengeStore) Store(_ context.Context, challenge *AuthChallenge, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.challenges[challenge.ID] = &storedChallenge{
		challenge: challenge,
		secret:    secret,
		expiresAt: challenge.ExpiresAt,
	}

	return nil
}

// Retrieve retrieves an authentication challenge and its secret.
func (s *InMemoryChallengeStore) Retrieve(_ context.Context, challengeID googleUuid.UUID) (*AuthChallenge, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stored, ok := s.challenges[challengeID]
	if !ok {
		return nil, "", fmt.Errorf("challenge not found")
	}

	if time.Now().After(stored.expiresAt) {
		return nil, "", fmt.Errorf("challenge expired")
	}

	return stored.challenge, stored.secret, nil
}

// Update updates an existing authentication challenge (e.g., retry count).
func (s *InMemoryChallengeStore) Update(ctx context.Context, challenge *AuthChallenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.challenges[challenge.ID]
	if !ok {
		return fmt.Errorf("challenge not found")
	}

	stored.challenge = challenge

	return nil
}

// Delete deletes an authentication challenge.
func (s *InMemoryChallengeStore) Delete(ctx context.Context, challengeID googleUuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.challenges, challengeID)

	return nil
}

// cleanup periodically removes expired challenges.
func (s *InMemoryChallengeStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		now := time.Now()
		for id, stored := range s.challenges {
			if now.After(stored.expiresAt) {
				delete(s.challenges, id)
			}
		}

		s.mu.Unlock()
	}
}

// InMemoryRateLimiter implements RateLimiter using in-memory storage.
type InMemoryRateLimiter struct {
	mu          sync.RWMutex
	attempts    map[string]*attemptRecord
	maxAttempts int
	window      time.Duration
	lockoutTime time.Duration
}

type attemptRecord struct {
	count       int
	windowStart time.Time
	lockedUntil time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		attempts:    make(map[string]*attemptRecord),
		maxAttempts: cryptoutilIdentityMagic.MaxOTPAttempts,
		window:      cryptoutilIdentityMagic.DefaultRateLimitWindow,
		lockoutTime: cryptoutilIdentityMagic.DefaultOTPLockout,
	}
}

// CheckLimit checks if a rate limit has been exceeded.
func (r *InMemoryRateLimiter) CheckLimit(ctx context.Context, identifier string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, ok := r.attempts[identifier]
	if !ok {
		return nil
	}

	now := time.Now()

	// Check if locked out.
	if !record.lockedUntil.IsZero() && now.Before(record.lockedUntil) {
		remaining := record.lockedUntil.Sub(now)

		return fmt.Errorf("rate limit exceeded, try again in %v", remaining.Round(time.Second))
	}

	// Check if window expired.
	if now.After(record.windowStart.Add(r.window)) {
		return nil
	}

	// Check attempt count.
	if record.count >= r.maxAttempts {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// RecordAttempt records an authentication attempt.
func (r *InMemoryRateLimiter) RecordAttempt(ctx context.Context, identifier string, success bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	record, ok := r.attempts[identifier]

	if !ok {
		record = &attemptRecord{
			count:       0,
			windowStart: now,
		}
		r.attempts[identifier] = record
	}

	// Reset if window expired.
	if now.After(record.windowStart.Add(r.window)) {
		record.count = 0
		record.windowStart = now
		record.lockedUntil = time.Time{}
	}

	// Increment count on failure.
	if !success {
		record.count++

		// Lock out if max attempts exceeded.
		if record.count >= r.maxAttempts {
			record.lockedUntil = now.Add(r.lockoutTime)
		}
	} else {
		// Reset on success.
		record.count = 0
		record.windowStart = now
		record.lockedUntil = time.Time{}
	}

	return nil
}
