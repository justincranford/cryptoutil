// Copyright (c) 2025-2026 Justin Cranford.
package userauth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity-idp/userauth"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	pushNotificationTestToken = "push-approval-token"
	pushNotificationTestUser  = "push-test-user"
)

type userauthTestOTPGeneratorStub struct {
	otp      string
	otpErr   error
	token    string
	tokenErr error
}

func (g *userauthTestOTPGeneratorStub) GenerateOTP(_ int) (string, error) {
	return g.otp, g.otpErr
}

func (g *userauthTestOTPGeneratorStub) GenerateSecureToken(_ int) (string, error) {
	return g.token, g.tokenErr
}

type userauthTestChallengeStoreStub struct {
	challenge   *cryptoutilIdentityIdpUserauth.AuthChallenge
	secret      string
	storeErr    error
	retrieveErr error
	deleteErr   error
}

func (s *userauthTestChallengeStoreStub) Store(_ context.Context, challenge *cryptoutilIdentityIdpUserauth.AuthChallenge, secret string) error {
	s.challenge = challenge
	s.secret = secret

	return s.storeErr
}

func (s *userauthTestChallengeStoreStub) Retrieve(_ context.Context, _ googleUuid.UUID) (*cryptoutilIdentityIdpUserauth.AuthChallenge, string, error) {
	return s.challenge, s.secret, s.retrieveErr
}

func (s *userauthTestChallengeStoreStub) Update(_ context.Context, _ *cryptoutilIdentityIdpUserauth.AuthChallenge) error {
	return nil
}

func (s *userauthTestChallengeStoreStub) Delete(_ context.Context, _ googleUuid.UUID) error {
	return s.deleteErr
}

type userauthTestRateLimiterStub struct {
	checkErr  error
	recordErr error
}

func (r *userauthTestRateLimiterStub) CheckLimit(_ context.Context, _ string) error {
	return r.checkErr
}

func (r *userauthTestRateLimiterStub) RecordAttempt(_ context.Context, _ string, _ bool) error {
	return r.recordErr
}

func TestPushNotificationAuthenticator_InitiateAuthErrorPaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected failure")
	tests := []struct {
		name          string
		token         string
		generatorErr  error
		storeErr      error
		rateCheckErr  error
		pushFailure   bool
		rateRecordErr error
		wantErr       string
	}{
		{name: testUserauthRateLimitedCase, rateCheckErr: injectedErr, wantErr: testUserauthRateLimitExceededError},
		{name: testUserauthTokenGenerationFailureCase, generatorErr: injectedErr, token: pushNotificationTestToken, wantErr: "failed to generate approval token"},
		{name: testUserauthEmptyGeneratedTokenCase, wantErr: "failed to hash approval token"},
		{name: testUserauthChallengeStorageFailureCase, token: pushNotificationTestToken, storeErr: injectedErr, wantErr: testChallengeStoreFailedError},
		{name: "push delivery failure", token: pushNotificationTestToken, pushFailure: true, wantErr: "failed to send push notification"},
		{name: "rate attempt recording failure", token: pushNotificationTestToken, rateRecordErr: injectedErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7()).String()
			userRepo := newMockPushUserRepo()
			require.NoError(t, userRepo.Create(ctx, &cryptoutilIdentityDomain.User{
				ID:              googleUuid.Must(googleUuid.NewV7()),
				Sub:             userID,
				PushDeviceToken: "push-device-token",
			}))

			pushService := newMockPushNotificationService()
			pushService.shouldFail = tc.pushFailure
			challengeStore := &userauthTestChallengeStoreStub{storeErr: tc.storeErr}
			rateLimiter := &userauthTestRateLimiterStub{checkErr: tc.rateCheckErr, recordErr: tc.rateRecordErr}
			generator := &userauthTestOTPGeneratorStub{token: tc.token, tokenErr: tc.generatorErr}
			auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(generator, pushService, challengeStore, rateLimiter, userRepo)

			challenge, err := auth.InitiateAuth(ctx, userID)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, challenge)
			require.Len(t, pushService.sentNotifications, 1)
			require.NotEmpty(t, challengeStore.secret)
			require.NotEqual(t, tc.token, challengeStore.secret)
		})
	}
}

func TestPushNotificationAuthenticator_VerifyAuthErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validTokenHash, err := cryptoutilIdentityIdpUserauth.HashToken(pushNotificationTestToken)
	require.NoError(t, err)

	tests := []struct {
		name             string
		challengeID      string
		challenge        *cryptoutilIdentityIdpUserauth.AuthChallenge
		challengeExpired bool
		secret           string
		retrieveErr      error
		deleteErr        error
		recordErr        error
		userExists       bool
		response         string
		wantErr          string
		wantUser         bool
	}{
		{name: testInvalidChallengeIDError, challengeID: cryptoutilSharedMagic.DevSetupTestInvalidName, wantErr: testInvalidChallengeIDError},
		{name: testUserauthChallengeRetrievalFailureCase, challengeID: googleUuid.Must(googleUuid.NewV7()).String(), retrieveErr: errors.New("missing"), wantErr: testUserauthChallengeNotFoundError},
		{
			name:             testUserauthExpiredChallengeCase,
			challengeID:      googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:        &cryptoutilIdentityIdpUserauth.AuthChallenge{ID: googleUuid.Must(googleUuid.NewV7())},
			challengeExpired: true,
			deleteErr:        errors.New("delete failed"),
			wantErr:          "push notification expired",
		},
		{
			name:        "approval token mismatch",
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:   &cryptoutilIdentityIdpUserauth.AuthChallenge{UserID: pushNotificationTestUser},
			secret:      validTokenHash,
			recordErr:   errors.New("record failed"),
			response:    "incorrect-token",
			wantErr:     "invalid approval token",
		},
		{
			name:        "user missing after token verification",
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:   &cryptoutilIdentityIdpUserauth.AuthChallenge{UserID: pushNotificationTestUser},
			secret:      validTokenHash,
			response:    pushNotificationTestToken,
			wantErr:     "user not found after verification",
		},
		{
			name:        "successful verification ignores challenge delete failure",
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:   &cryptoutilIdentityIdpUserauth.AuthChallenge{UserID: pushNotificationTestUser},
			secret:      validTokenHash,
			deleteErr:   errors.New("delete failed"),
			userExists:  true,
			response:    pushNotificationTestToken,
			wantUser:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			challenge := tc.challenge
			if challenge != nil {
				challengeCopy := *challenge
				if tc.challengeExpired {
					challengeCopy.ExpiresAt = time.Now().UTC().Add(-time.Minute)
				} else {
					challengeCopy.ExpiresAt = time.Now().UTC().Add(time.Hour)
				}

				challenge = &challengeCopy
			}

			challengeStore := &userauthTestChallengeStoreStub{
				challenge:   challenge,
				secret:      tc.secret,
				retrieveErr: tc.retrieveErr,
				deleteErr:   tc.deleteErr,
			}
			rateLimiter := &userauthTestRateLimiterStub{recordErr: tc.recordErr}

			userRepo := newMockPushUserRepo()
			if tc.userExists {
				require.NoError(t, userRepo.Create(ctx, &cryptoutilIdentityDomain.User{
					ID:  googleUuid.Must(googleUuid.NewV7()),
					Sub: pushNotificationTestUser,
				}))
			}

			auth := cryptoutilIdentityIdpUserauth.NewPushNotificationAuthenticator(
				&userauthTestOTPGeneratorStub{},
				newMockPushNotificationService(),
				challengeStore,
				rateLimiter,
				userRepo,
			)

			user, verifyErr := auth.VerifyAuth(ctx, tc.challengeID, tc.response)
			if tc.wantErr != "" {
				require.ErrorContains(t, verifyErr, tc.wantErr)
				require.Nil(t, user)

				return
			}

			require.NoError(t, verifyErr)
			require.NotNil(t, user)
			require.Equal(t, pushNotificationTestUser, user.Sub)
		})
	}
}
