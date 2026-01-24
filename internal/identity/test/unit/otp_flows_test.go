// Copyright (c) 2025 Justin Cranford

package unit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/identity/idp/userauth"
	cryptoutilIdentityIdpUserauthMocks "cryptoutil/internal/identity/idp/userauth/mocks"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// TestSMSOTPCompleteFlow tests the complete SMS OTP flow: generate → send → validate.
// NOTE: This test uses local mocks only (no HTTP servers required).
func TestSMSOTPCompleteFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup mock delivery service.
	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()

	// Setup mock challenge store.
	challengeStore := newMockChallengeStore()

	// Setup mock rate limiter (permissive for E2E test).
	rateLimiter := newMockRateLimiter(100, 5*time.Minute) // 100 attempts per 5 minutes

	// Setup mock user repository with test user.
	userRepo := newMockUserRepository()
	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		Email:       "test@example.com",
		PhoneNumber: "+15551234567",
		Name:        "Test User",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	// Create SMS OTP authenticator.
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator,
		mockSMS,
		challengeStore,
		rateLimiter,
		userRepo,
	)

	// Step 1: Initiate authentication (generate OTP, send SMS).
	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge)
	require.Equal(t, "sms_otp", challenge.Method)
	require.Equal(t, testUser.Sub, challenge.UserID)
	require.True(t, time.Now().Before(challenge.ExpiresAt), "Challenge should not be expired")

	// Verify SMS sent.
	require.Equal(t, 1, mockSMS.GetCallCount(), "SMS should be sent")
	messages := mockSMS.GetSentMessages()
	require.Len(t, messages, 1)
	require.Contains(t, messages[0].Message, "Your verification code is:", "SMS should contain OTP")

	// Extract OTP from mock SMS (format: "Your verification code is: 123456 ...").
	var otp string

	_, err = fmt.Sscanf(messages[0].Message, "Your verification code is: %s", &otp)
	require.NoError(t, err, "OTP should be extractable from SMS")
	require.Len(t, otp, cryptoutilIdentityMagic.DefaultOTPLength, "OTP should be 6 digits")

	// Step 2: Verify authentication with correct OTP.
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), otp)
	require.NoError(t, err, "VerifyAuth should succeed with correct OTP")
	require.NotNil(t, user)
	require.Equal(t, testUser.Sub, user.Sub)
	require.Equal(t, testUser.PhoneNumber, user.PhoneNumber)

	// Step 3: Verify challenge deleted (single-use).
	_, _, err = challengeStore.Retrieve(ctx, challenge.ID)
	require.Error(t, err, "Challenge should be deleted after successful verification")
}

// TestSMSOTPInvalidToken tests OTP validation with incorrect token.
func TestSMSOTPInvalidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(100, 5*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		PhoneNumber: "+15551234567",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator, mockSMS, challengeStore, rateLimiter, userRepo,
	)

	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err)

	// Attempt verification with WRONG OTP.
	wrongOTP := "000000"
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), wrongOTP)

	require.Error(t, err, "VerifyAuth should fail with wrong OTP")
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid OTP", "Error should indicate invalid OTP")
}

// TestSMSOTPExpiredChallenge tests OTP validation after expiration.
func TestSMSOTPExpiredChallenge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(100, 5*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		PhoneNumber: "+15551234567",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator, mockSMS, challengeStore, rateLimiter, userRepo,
	)

	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err)

	// Manually expire challenge.
	challenge.ExpiresAt = time.Now().Add(-1 * time.Minute)
	challengeStore.challenges[challenge.ID] = challengeEntry{
		challenge: challenge,
		token:     challengeStore.challenges[challenge.ID].token,
	}

	// Attempt verification with expired challenge.
	otp := "123456" // Doesn't matter, challenge expired
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), otp)

	require.Error(t, err, "VerifyAuth should fail with expired challenge")
	require.Nil(t, user)
	require.Contains(t, err.Error(), "expired", "Error should indicate expiration")
}

// TestSMSOTPRateLimitEnforcement tests per-user rate limiting enforcement.
func TestSMSOTPRateLimitEnforcement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSMS := cryptoutilIdentityIdpUserauthMocks.NewSMSProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(3, 1*time.Minute) // Only 3 attempts per minute
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:         googleUuid.Must(googleUuid.NewV7()).String(),
		PhoneNumber: "+15551234567",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(
		generator, mockSMS, challengeStore, rateLimiter, userRepo,
	)

	// First 3 attempts should succeed.
	for i := 0; i < 3; i++ {
		_, err := authenticator.InitiateAuth(ctx, testUser.Sub)
		require.NoError(t, err, "Attempt %d should succeed", i+1)
	}

	// 4th attempt should fail due to rate limiting.
	_, err = authenticator.InitiateAuth(ctx, testUser.Sub)
	require.Error(t, err, "4th attempt should fail due to rate limit")
	require.Contains(t, err.Error(), "rate limit", "Error should indicate rate limiting")
}

// TestEmailMagicLinkCompleteFlow tests the complete magic link flow: generate → send → validate.
func TestEmailMagicLinkCompleteFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockEmail := cryptoutilIdentityIdpUserauthMocks.NewEmailProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(100, 5*time.Minute)
	userRepo := newMockUserRepository()

	testUser := &cryptoutilIdentityDomain.User{
		Sub:   googleUuid.Must(googleUuid.NewV7()).String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	authenticator := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(
		generator, mockEmail, challengeStore, rateLimiter, userRepo,
		"https://example.com", // base URL
	)

	// Step 1: Initiate authentication (generate token, send email).
	challenge, err := authenticator.InitiateAuth(ctx, testUser.Sub)
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge)
	require.Equal(t, "magic_link", challenge.Method)
	require.Equal(t, testUser.Sub, challenge.UserID)

	// Verify email sent.
	require.Equal(t, 1, mockEmail.GetCallCount(), "Email should be sent")
	emails := mockEmail.GetSentEmails()
	require.Len(t, emails, 1)
	require.Contains(t, emails[0].Body, "https://example.com/auth/magic-link/verify", "Email should contain magic link")

	// Extract token from email body (format: "...?token=abc123&challenge=uuid").
	var token string

	lines := splitLines(emails[0].Body)
	for _, line := range lines {
		if strings.Contains(line, "https://example.com/auth/magic-link/verify?token=") {
			parts := strings.Split(line, "token=")
			if len(parts) >= 2 {
				tokenPart := strings.Split(parts[1], "&")[0]
				token = tokenPart

				break
			}
		}
	}

	require.NotEmpty(t, token, "Token should be extractable from email")

	// Step 2: Verify authentication with correct token.
	user, err := authenticator.VerifyAuth(ctx, challenge.ID.String(), token)
	require.NoError(t, err, "VerifyAuth should succeed with correct token")
	require.NotNil(t, user)
	require.Equal(t, testUser.Sub, user.Sub)
	require.Equal(t, testUser.Email, user.Email)

	// Step 3: Verify challenge deleted (single-use).
	_, _, err = challengeStore.Retrieve(ctx, challenge.ID)
	require.Error(t, err, "Challenge should be deleted after successful verification")
}

// TestMagicLinkInvalidToken tests magic link validation with incorrect token.
func TestMagicLinkInvalidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockEmail := cryptoutilIdentityIdpUserauthMocks.NewEmailProvider()
	challengeStore := newMockChallengeStore()
	rateLimiter := newMockRateLimiter(100, 5*time.Minute)
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
	rateLimiter := newMockRateLimiter(5, 1*time.Minute) // Only 5 attempts per minute
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
	for i := 0; i < 5; i++ {
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

	now := time.Now()
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

	r.attempts[userID] = append(r.attempts[userID], time.Now())

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
