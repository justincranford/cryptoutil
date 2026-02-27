// Copyright (c) 2025 Justin Cranford

package unit

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
	cryptoutilIdentityIdpUserauthMocks "cryptoutil/internal/apps/identity/idp/userauth/mocks"
)

func TestMagicLinkInvalidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockEmail := cryptoutilIdentityIdpUserauthMocks.NewEmailProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(cryptoutilSharedMagic.JoseJAMaxMaterials, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: "test@example.com",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(
		generator, mockEmail, challengeStore, rateLimiter, userRepo,
		"https://example.com",
	)

	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err)

	// Attempt verification with WRONG token.
	wrongToken := "invalid-token-12345"
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), wrongToken)

	require.Error(t, err, "VerifyAuth should fail with wrong token")
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid magic link token", "Error should indicate invalid token")
}

// TestMagicLinkRateLimitEnforcement tests per-user rate limiting for magic links.
func TestMagicLinkRateLimitEnforcement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockEmail := cryptoutilIdentityIdpUserauthMocks.NewEmailProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, 1*time.Minute) // Only 5 attempts per minute
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: "test@example.com",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(
		generator, mockEmail, challengeStore, rateLimiter, userRepo,
		"https://example.com",
	)

	// First 5 attempts should succeed.
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		_, err := authenticator.InitiateAuth(ctx, testUser.Sub)
		require.NoError(t, err, "Attempt %d should succeed", i+1)
	}

	// 6th attempt should fail due to rate limiting.
	_, err = authenticator.InitiateAuth(ctx, testUser.Sub)
	require.Error(t, err, "6th attempt should fail due to rate limit")
	require.Contains(t, err.Error(), "rate limit", "Error should indicate rate limiting")
}

// Helper functions and mock implementations.

type mockChallengeStore struct {
	mu         sync.Mutex
	challenges map[googleUuid.UUID]challengeEntry
}

type challengeEntry struct {
	challenge *cryptoutilIdentityIdpUserauth.AuthChallenge
	token     string // Hashed token
}

func newMockChallengeStore() *mockChallengeStore {
	return &mockChallengeStore{
		challenges: make(map[googleUuid.UUID]challengeEntry),
	}
}

func (s *mockChallengeStore) Store(_ context.Context, challenge *cryptoutilIdentityIdpUserauth.AuthChallenge, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.challenges[challenge.ID] = challengeEntry{
		challenge: challenge,
		token:     token,
	}

	return nil
}

func (s *mockChallengeStore) Retrieve(_ context.Context, id googleUuid.UUID) (*cryptoutilIdentityIdpUserauth.AuthChallenge, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.challenges[id]
	if !exists {
		return nil, "", fmt.Errorf("challenge not found")
	}

	return entry.challenge, entry.token, nil
}

func (s *mockChallengeStore) Update(_ context.Context, challenge *cryptoutilIdentityIdpUserauth.AuthChallenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.challenges[challenge.ID]
	if !exists {
		return fmt.Errorf("challenge not found")
	}

	entry.challenge = challenge
	s.challenges[challenge.ID] = entry

	return nil
}

func (s *mockChallengeStore) Delete(_ context.Context, id googleUuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.challenges, id)

	return nil
}

type mockRateLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

func newMockRateLimiter(maxAttempts int, window time.Duration) *mockRateLimiter {
	return &mockRateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

func (r *mockRateLimiter) CheckLimit(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	attempts := r.attempts[userID]

	// Remove expired attempts.
	validAttempts := []time.Time{}

	for _, t := range attempts {
		if now.Sub(t) < r.window {
			validAttempts = append(validAttempts, t)
		}
	}

	if len(validAttempts) >= r.maxAttempts {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

func (r *mockRateLimiter) RecordAttempt(_ context.Context, userID string, _ bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.attempts[userID] = append(r.attempts[userID], time.Now().UTC())

	return nil
}

type mockUserRepository struct {
	mu    sync.Mutex
	users map[string]*cryptoutilIdentityDomain.User // key: sub
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*cryptoutilIdentityDomain.User),
	}
}

func (r *mockUserRepository) Create(_ context.Context, user *cryptoutilIdentityDomain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.Sub] = user

	return nil
}

func (r *mockUserRepository) GetBySub(_ context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[sub]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r *mockUserRepository) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, fmt.Errorf("GetByID not implemented in mock")
}

func (r *mockUserRepository) GetByUsername(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, fmt.Errorf("GetByUsername not implemented in mock")
}

func (r *mockUserRepository) GetByEmail(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, fmt.Errorf("GetByEmail not implemented in mock")
}

func (r *mockUserRepository) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return fmt.Errorf("Update not implemented in mock")
}

func (r *mockUserRepository) Delete(_ context.Context, _ googleUuid.UUID) error {
	return fmt.Errorf("Delete not implemented in mock")
}

func (r *mockUserRepository) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, fmt.Errorf("List not implemented in mock")
}

func (r *mockUserRepository) Count(_ context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return int64(len(r.users)), nil
}

func splitLines(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
}
