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

// --- fixedOTPGenerator returns deterministic OTP for testing ---

type fixedOTPGenerator struct {
	otp   string
	token string
}

func newFixedOTPGenerator(otp, token string) *fixedOTPGenerator {
	return &fixedOTPGenerator{otp: otp, token: token}
}

func (g *fixedOTPGenerator) GenerateOTP(_ int) (string, error)         { return g.otp, nil }
func (g *fixedOTPGenerator) GenerateSecureToken(_ int) (string, error) { return g.token, nil }


// --- coverageUserRepo for shared use ---

type coverageUserRepo struct {
	user    *cryptoutilIdentityDomain.User
	failGet bool
}

func newCoverageUserRepo(user *cryptoutilIdentityDomain.User) *coverageUserRepo {
	return &coverageUserRepo{user: user}
}

func (r *coverageUserRepo) GetBySub(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	if r.failGet {
		return nil, fmt.Errorf("user lookup failed")
	}

	return r.user, nil
}

func (r *coverageUserRepo) Create(_ context.Context, u *cryptoutilIdentityDomain.User) error {
	r.user = u

	return nil
}

func (r *coverageUserRepo) GetByID(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}
func (r *coverageUserRepo) GetByUsername(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}
func (r *coverageUserRepo) GetByEmail(_ context.Context, _ string) (*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}
func (r *coverageUserRepo) Update(_ context.Context, _ *cryptoutilIdentityDomain.User) error {
	return nil
}
func (r *coverageUserRepo) Delete(_ context.Context, _ googleUuid.UUID) error { return nil }
func (r *coverageUserRepo) List(_ context.Context, _, _ int) ([]*cryptoutilIdentityDomain.User, error) {
	return nil, nil
}
func (r *coverageUserRepo) Count(_ context.Context) (int64, error) { return 0, nil }

// --- SMS OTP coverage tests ---

func TestSMSOTPAuthenticator_VerifyAuth_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "999888"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000000"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, cryptoutilIdentityIdpUserauth.NewMockDeliveryService(), challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)
	verifiedUser, err := auth.VerifyAuth(ctx, challenge.ID.String(), knownOTP)
	require.NoError(t, err)
	require.NotNil(t, verifiedUser)
	require.Equal(t, userID, verifiedUser.ID)
}

func TestSMSOTPAuthenticator_VerifyAuth_ExpiredOTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "111111"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000001"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, cryptoutilIdentityIdpUserauth.NewMockDeliveryService(), challengeStore, rateLimiter, userRepo)
	// Manually store expired challenge
	hashedOTP, err := cryptoutilIdentityIdpUserauth.HashToken(knownOTP)
	require.NoError(t, err)

	expiredChallenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.New(),
		UserID:    userID.String(),
		Method:    "sms_otp",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, challengeStore.Store(ctx, expiredChallenge, hashedOTP))
	_, err = auth.VerifyAuth(ctx, expiredChallenge.ID.String(), knownOTP)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestSMSOTPAuthenticator_VerifyAuth_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "222222"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000002"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, cryptoutilIdentityIdpUserauth.NewMockDeliveryService(), challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)
	// Make user not found after challenge stored
	userRepo.failGet = true
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), knownOTP)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user not found")
}

// --- Phone Call OTP coverage tests ---

func TestPhoneCallOTPAuthenticator_VerifyAuth_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "777666"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000003"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	phoneService := newMockPhoneCallService()
	auth := cryptoutilIdentityIdpUserauth.NewPhoneCallOTPAuthenticator(generator, phoneService, challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)
	verifiedUser, err := auth.VerifyAuth(ctx, challenge.ID.String(), knownOTP)
	require.NoError(t, err)
	require.NotNil(t, verifiedUser)
	require.Equal(t, userID, verifiedUser.ID)
}

func TestPhoneCallOTPAuthenticator_VerifyAuth_ExpiredOTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "333333"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000004"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	phoneService := newMockPhoneCallService()
	auth := cryptoutilIdentityIdpUserauth.NewPhoneCallOTPAuthenticator(generator, phoneService, challengeStore, rateLimiter, userRepo)
	hashedOTP, err := cryptoutilIdentityIdpUserauth.HashToken(knownOTP)
	require.NoError(t, err)

	expiredChallenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.New(),
		UserID:    userID.String(),
		Method:    "phone_call_otp",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, challengeStore.Store(ctx, expiredChallenge, hashedOTP))
	_, err = auth.VerifyAuth(ctx, expiredChallenge.ID.String(), knownOTP)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestPhoneCallOTPAuthenticator_VerifyAuth_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "444444"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000005"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	phoneService := newMockPhoneCallService()
	auth := cryptoutilIdentityIdpUserauth.NewPhoneCallOTPAuthenticator(generator, phoneService, challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)

	userRepo.failGet = true
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), knownOTP)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user not found")
}

func TestPhoneCallOTPAuthenticator_VerifyAuth_MaxRetries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownOTP = "555555"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+1555000006"}
	generator := newFixedOTPGenerator(knownOTP, "")
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	phoneService := newMockPhoneCallService()
	// Default maxRetries is 3; submit wrong OTP 3 times to trigger max retries exceeded
	auth := cryptoutilIdentityIdpUserauth.NewPhoneCallOTPAuthenticator(generator, phoneService, challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)

	for range 3 {
		_, err = auth.VerifyAuth(ctx, challenge.ID.String(), "000000")
		require.Error(t, err)
	}
}

// --- Push Notification coverage tests ---

func TestPushNotificationAuthenticator_VerifyAuth_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownToken = "push-approval-token-xyz"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PushDeviceToken: "device-token-abc"}
	generator := newFixedOTPGenerator("", knownToken)
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	pushService := newMockPushNotificationService()
	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, pushService, challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)
	verifiedUser, err := auth.VerifyAuth(ctx, challenge.ID.String(), knownToken)
	require.NoError(t, err)
	require.NotNil(t, verifiedUser)
	require.Equal(t, userID, verifiedUser.ID)
}

func TestPushNotificationAuthenticator_VerifyAuth_ExpiredToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownToken = "push-token-expired"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PushDeviceToken: "device-token-exp"}
	generator := newFixedOTPGenerator("", knownToken)
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	pushService := newMockPushNotificationService()
	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, pushService, challengeStore, rateLimiter, userRepo)
	hashedToken, err := cryptoutilIdentityIdpUserauth.HashToken(knownToken)
	require.NoError(t, err)

	expiredChallenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.New(),
		UserID:    userID.String(),
		Method:    "push_notification",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	require.NoError(t, challengeStore.Store(ctx, expiredChallenge, hashedToken))
	_, err = auth.VerifyAuth(ctx, expiredChallenge.ID.String(), knownToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestPushNotificationAuthenticator_VerifyAuth_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const knownToken = "push-token-userfail"

	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PushDeviceToken: "device-token-uf"}
	generator := newFixedOTPGenerator("", knownToken)
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	pushService := newMockPushNotificationService()
	auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, pushService, challengeStore, rateLimiter, userRepo)
	challenge, err := auth.InitiateAuth(ctx, userID.String())
	require.NoError(t, err)

	userRepo.failGet = true
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), knownToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user not found")
}

// --- Additional magic link coverage ---

func TestMagicLinkAuthenticator_VerifyAuth_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := googleUuid.New()
	user := &cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), Email: "cov@example.com"}
	generator := &cryptoutilIdentityIdpUserauth.DefaultOTPGenerator{}
	userRepo := newCoverageUserRepo(user)
	challengeStore := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	rateLimiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, cryptoutilIdentityIdpUserauth.NewMockDeliveryService(), challengeStore, rateLimiter, userRepo, "https://example.com")
	token, err := generator.GenerateSecureToken(cryptoutilIdentityMagic.DefaultMagicLinkLength)
	require.NoError(t, err)
	hashedToken, err := cryptoutilIdentityIdpUserauth.HashToken(token)
	require.NoError(t, err)

	challenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        googleUuid.New(),
		UserID:    userID.String(),
		Method:    "magic_link",
		ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
		Metadata:  map[string]any{"email": user.Email},
	}
	require.NoError(t, challengeStore.Store(ctx, challenge, hashedToken))

	userRepo.failGet = true
	_, err = auth.VerifyAuth(ctx, challenge.ID.String(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user not found")
}
