// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// mockPushNotificationService implements PushNotificationService for testing.
type mockPushNotificationService struct {
	sentNotifications []mockPushNotification
	shouldFail        bool
}

type mockPushNotification struct {
	DeviceToken string
	Title       string
	Body        string
	Data        map[string]any
}

func newMockPushNotificationService() *mockPushNotificationService {
	return &mockPushNotificationService{
		sentNotifications: []mockPushNotification{},
	}
}

func (m *mockPushNotificationService) SendPushNotification(_ context.Context, deviceToken, title, body string, data map[string]any) error {
	if m.shouldFail {
		return fmt.Errorf("mock push notification service configured to fail")
	}

	m.sentNotifications = append(m.sentNotifications, mockPushNotification{
		DeviceToken: deviceToken,
		Title:       title,
		Body:        body,
		Data:        data,
	})

	return nil
}

// mockPushUserRepo implements UserRepository for push notification testing.
type mockPushUserRepo struct {
	users map[string]*cryptoutilIdentityDomain.User
}

func newMockPushUserRepo() *mockPushUserRepo {
	return &mockPushUserRepo{
		users: make(map[string]*cryptoutilIdentityDomain.User),
	}
}

func (m *mockPushUserRepo) GetBySub(_ context.Context, sub string) (*cryptoutilIdentityDomain.User, error) {
	user, ok := m.users[sub]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", sub)
	}

	return user, nil
}

func (m *mockPushUserRepo) Create(_ context.Context, user *cryptoutilIdentityDomain.User) error {
	m.users[user.Sub] = user

	return nil
}

func (m *mockPushUserRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockPushUserRepo) GetByUsername(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockPushUserRepo) GetByEmail(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockPushUserRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}

func (m *mockPushUserRepo) Delete(_ context.Context, _ googleUuid.UUID) error {
	return nil
}

func (m *mockPushUserRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}

func (m *mockPushUserRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func TestPushNotificationAuthenticator_NewAuthenticator(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(nil, nil, nil, nil, nil)
	require.NotNil(t, auth, "NewPushNotificationAuthenticator should return non-nil authenticator")
}

func TestPushNotificationAuthenticator_Method(t *testing.T) {
	t.Parallel()

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(nil, nil, nil, nil, nil)
	require.Equal(t, "push_notification", auth.Method())
}

func TestPushNotificationAuthenticator_InitiateAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := googleUuid.New().String()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	mockPush := newMockPushNotificationService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	userRepo := newMockPushUserRepo()

	// Create test user with push device token.
	user := &cryptoutilIdentityDomain.User{
		ID:              googleUuid.New(),
		Sub:             userID,
		Email:           "user@example.com",
		PushDeviceToken: "fcm-device-token-123",
		PasswordHash:    "test-hash",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, mockPush, challengeStore, rateLimiter, userRepo)

	// Initiate push notification authentication.
	beforeInitiate := time.Now().UTC()
	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, challenge)
	require.Equal(t, userID, challenge.UserID)
	require.Equal(t, "push_notification", challenge.Method)
	require.WithinDuration(t, beforeInitiate.Add(cryptoutilIdentityMagic.DefaultPushNotificationTimeout), challenge.ExpiresAt, 5*time.Second)

	// Verify push notification was sent.
	require.Len(t, mockPush.sentNotifications, 1)
	notification := mockPush.sentNotifications[0]
	require.Equal(t, "fcm-device-token-123", notification.DeviceToken)
	require.Contains(t, notification.Title, "Authentication Request")
	require.Contains(t, notification.Body, "approve")
	require.Contains(t, notification.Data, "challenge_id")
	require.Contains(t, notification.Data, "approval_token")
	require.Contains(t, notification.Data, "expires_at")
}

func TestPushNotificationAuthenticator_InitiateAuthUserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := googleUuid.New().String()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	mockPush := newMockPushNotificationService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	userRepo := newMockPushUserRepo()

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, mockPush, challengeStore, rateLimiter, userRepo)

	// Initiate with non-existent user.
	_, err := auth.InitiateAuth(ctx, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user not found")
}

func TestPushNotificationAuthenticator_InitiateAuthNoDeviceToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := googleUuid.New().String()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	mockPush := newMockPushNotificationService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	userRepo := newMockPushUserRepo()

	// Create user without push device token.
	user := &cryptoutilIdentityDomain.User{
		ID:              googleUuid.New(),
		Sub:             userID,
		Email:           "user@example.com",
		PushDeviceToken: "", // No device token.
		PasswordHash:    "test-hash",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, mockPush, challengeStore, rateLimiter, userRepo)

	// Initiate with user that has no device token.
	_, err = auth.InitiateAuth(ctx, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no push device token")
}

func TestPushNotificationAuthenticator_VerifyAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := googleUuid.New().String()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	mockPush := newMockPushNotificationService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	userRepo := newMockPushUserRepo()

	// Create test user.
	user := &cryptoutilIdentityDomain.User{
		ID:              googleUuid.New(),
		Sub:             userID,
		Email:           "user@example.com",
		PushDeviceToken: "fcm-device-token-123",
		PasswordHash:    "test-hash",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, mockPush, challengeStore, rateLimiter, userRepo)

	// Initiate authentication.
	challenge, err := auth.InitiateAuth(ctx, userID)
	require.NoError(t, err)

	// Extract approval token from sent notification.
	require.Len(t, mockPush.sentNotifications, 1)
	notification := mockPush.sentNotifications[0]
	approvalToken, ok := notification.Data["approval_token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, approvalToken)

	// Verify authentication with approval token.
	verifiedUser, err := auth.VerifyAuth(ctx, challenge.ID.String(), approvalToken)
	require.NoError(t, err)
	require.NotNil(t, verifiedUser)
	require.Equal(t, userID, verifiedUser.Sub)
}

func TestPushNotificationAuthenticator_VerifyAuthChallengeNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	mockPush := newMockPushNotificationService()
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	userRepo := newMockPushUserRepo()

	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, mockPush, challengeStore, rateLimiter, userRepo)

	// Verify with non-existent challenge.
	challengeID := googleUuid.New().String()
	_, err := auth.VerifyAuth(ctx, challengeID, "some-token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "challenge not found")
}
