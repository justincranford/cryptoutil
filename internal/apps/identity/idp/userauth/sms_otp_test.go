// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
)

func TestDefaultOTPGenerator_GenerateOTP(t *testing.T) {
	t.Parallel()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	otp, err := generator.GenerateOTP(6)
	require.NoError(t, err, "GenerateOTP should succeed")
	require.Len(t, otp, 6, "OTP should be 6 digits")

	// Verify all characters are digits.
	for _, c := range otp {
		require.True(t, c >= '0' && c <= '9', "OTP should contain only digits: %s", otp)
	}
}

func TestDefaultOTPGenerator_GenerateOTPInvalidLength(t *testing.T) {
	t.Parallel()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	_, err := generator.GenerateOTP(0)
	require.Error(t, err, "GenerateOTP should fail with zero length")

	_, err = generator.GenerateOTP(-1)
	require.Error(t, err, "GenerateOTP should fail with negative length")
}

func TestDefaultOTPGenerator_GenerateOTPUniqueness(t *testing.T) {
	t.Parallel()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	otps := make(map[string]bool)

	// Generate multiple OTPs - with 6 digits, duplicates are statistically very unlikely.
	for range 20 {
		otp, err := generator.GenerateOTP(8) // Use 8 digits to reduce collision chance.
		require.NoError(t, err, "GenerateOTP should succeed")

		otps[otp] = true
	}

	// Expect at least 18 unique values (allowing for some statistical collision).
	require.GreaterOrEqual(t, len(otps), 18, "Most OTPs should be unique")
}

func TestDefaultOTPGenerator_GenerateSecureToken(t *testing.T) {
	t.Parallel()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	token, err := generator.GenerateSecureToken(16)
	require.NoError(t, err, "GenerateSecureToken should succeed")
	require.Len(t, token, 32, "Token should be double the byte length (hex encoded)")

	// Verify all characters are hex.
	for _, c := range token {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
		require.True(t, isHex, "Token should be valid hex: %s", token)
	}
}

func TestDefaultOTPGenerator_GenerateSecureTokenInvalidLength(t *testing.T) {
	t.Parallel()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	_, err := generator.GenerateSecureToken(0)
	require.Error(t, err, "GenerateSecureToken should fail with zero length")

	_, err = generator.GenerateSecureToken(-1)
	require.Error(t, err, "GenerateSecureToken should fail with negative length")
}

func TestDefaultOTPGenerator_GenerateSecureTokenUniqueness(t *testing.T) {
	t.Parallel()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	tokens := make(map[string]bool)

	// Generate multiple tokens and ensure they're unique.
	for range 10 {
		token, err := generator.GenerateSecureToken(16)
		require.NoError(t, err, "GenerateSecureToken should succeed")
		require.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
	}
}

func TestSMSOTPAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(nil, nil, nil, nil, nil)
	require.NotNil(t, auth, "NewSMSOTPAuthenticator should return non-nil authenticator")
}

func TestSMSOTPAuthenticator_Method(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(nil, nil, nil, nil, nil)
	require.Equal(t, "sms_otp", auth.Method(), "Method should return 'sms_otp'")
}

// mockSMSUserRepo implements UserRepository for SMS OTP testing.
type mockSMSUserRepo struct {
	users map[string]*cryptoutilIdentityDomain.User
}

func newMockSMSUserRepo() *mockSMSUserRepo {
	return &mockSMSUserRepo{
		users: make(map[string]*cryptoutilIdentityDomain.User),
	}
}

func (m *mockSMSUserRepo) GetBySub(_ context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.users[sub]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", sub)
	}

	return user, nil
}

func (m *mockSMSUserRepo) Create(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockSMSUserRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockSMSUserRepo) GetByUsername(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockSMSUserRepo) GetByEmail(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockSMSUserRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockSMSUserRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockSMSUserRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockSMSUserRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSMSUserRepo) AddUser(user *cryptoutilIdentityDomain.User) {
	m.users[user.Sub] = user
}

// mockSMSDeliveryService implements DeliveryService for testing.
type mockSMSDeliveryService struct {
	sentMessages []string
	shouldFail   bool
}

func newMockSMSDeliveryService() *mockSMSDeliveryService {
	return &mockSMSDeliveryService{
		sentMessages: make([]string, 0),
	}
}

func (m *mockSMSDeliveryService) SendSMS(_ context.Context, _, message string) error {
	if m.shouldFail {
		return fmt.Errorf("failed to send SMS")
	}

	m.sentMessages = append(m.sentMessages, message)

	return nil
}

func (m *mockSMSDeliveryService) SendEmail(_ context.Context, _, _, _ string) error {
	return nil
}

// TestSMSOTPAuthenticator_InitiateAuth tests InitiateAuth.
func TestSMSOTPAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockSMSUserRepo()
	delivery := newMockSMSDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Add user with phone number.
	user := &cryptoutilIdentityDomain.User{
		ID:          userID,
		Sub:         userID.String(),
		PhoneNumber: "+1234567890",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo)

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")
	require.Equal(t, userID.String(), challenge.UserID, "Challenge UserID should match")
	require.Equal(t, "sms_otp", challenge.Method, "Challenge Method should match")

	// Verify SMS was sent.
	require.Len(t, delivery.sentMessages, 1, "One SMS should have been sent")
}

// TestSMSOTPAuthenticator_InitiateAuthUserNotFound tests InitiateAuth with non-existent user.
func TestSMSOTPAuthenticator_InitiateAuthUserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockSMSUserRepo()
	delivery := newMockSMSDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo)

	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	challenge, err := auth.InitiateAuth(ctx, nonExistentID.String())
	require.Error(t, err, "InitiateAuth should fail for non-existent user")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "user not found", "Error should indicate user not found")
}

// TestSMSOTPAuthenticator_InitiateAuthNoPhoneNumber tests InitiateAuth without phone number.
func TestSMSOTPAuthenticator_InitiateAuthNoPhoneNumber(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockSMSUserRepo()
	delivery := newMockSMSDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Add user without phone number.
	user := &cryptoutilIdentityDomain.User{
		ID:          userID,
		Sub:         userID.String(),
		PhoneNumber: "",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo)

	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.Error(t, err, "InitiateAuth should fail without phone number")
	require.Nil(t, challenge, "Challenge should be nil on error")
	require.Contains(t, err.Error(), "no phone number", "Error should indicate missing phone number")
}

// TestSMSOTPAuthenticator_VerifyAuth tests VerifyAuth.
func TestSMSOTPAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockSMSUserRepo()
	delivery := newMockSMSDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	userID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	// Add user with phone number.
	user := &cryptoutilIdentityDomain.User{
		ID:          userID,
		Sub:         userID.String(),
		PhoneNumber: "+1234567890",
	}
	userRepo.AddUser(user)

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo)

	// Initiate auth first.
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err, "InitiateAuth should succeed")
	require.NotNil(t, challenge, "Challenge should not be nil")

	// VerifyAuth with invalid challenge ID.
	_, err = auth.VerifyAuth(ctx, "invalid-uuid", "123456")
	require.Error(t, err, "VerifyAuth should fail with invalid challenge ID")
	require.Contains(t, err.Error(), "invalid challenge ID", "Error should indicate invalid challenge ID")

	// VerifyAuth with wrong OTP (challenge exists but OTP is wrong).
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), "wrong-otp")
	require.Error(t, err, "VerifyAuth should fail with wrong OTP")
}

// TestSMSOTPAuthenticator_VerifyAuthChallengeNotFound tests VerifyAuth with non-existent challenge.
func TestSMSOTPAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := newMockSMSUserRepo()
	delivery := newMockSMSDeliveryService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}

	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo)

	// Generate a valid UUID that doesn't exist as a challenge.
	nonExistentID, err := googleUuid.NewV7()
	require.NoError(t, err, "NewV7 should succeed")

	_, err = auth.VerifyAuth(ctx, nonExistentID.String(), "123456")
	require.Error(t, err, "VerifyAuth should fail with non-existent challenge")
	require.Contains(t, err.Error(), "challenge not found", "Error should indicate challenge not found")
}
