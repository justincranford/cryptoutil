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
	magicLinkErrorPathToken = "magic-link-token"
	magicLinkErrorPathUser  = "magic-link-test-user"
)

func TestMagicLinkAuthenticator_InitiateAuthErrorPaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected failure")
	tests := []struct {
		name         string
		email        string
		token        string
		tokenErr     error
		storeErr     error
		checkErr     error
		recordErr    error
		deliveryFail bool
		userLookup   bool
		wantErr      string
	}{
		{name: testUserauthRateLimitedCase, checkErr: injectedErr, wantErr: testUserauthRateLimitExceededError},
		{name: testUserauthUserLookupFailureCase, email: testTestExampleCom, userLookup: true, wantErr: testUserauthUserNotFoundError},
		{name: "email missing", wantErr: "no email"},
		{name: testUserauthTokenGenerationFailureCase, email: testTestExampleCom, tokenErr: injectedErr, wantErr: "failed to generate token"},
		{name: testUserauthEmptyGeneratedTokenCase, email: testTestExampleCom, wantErr: "failed to hash token"},
		{name: testUserauthChallengeStorageFailureCase, email: testTestExampleCom, token: magicLinkErrorPathToken, storeErr: injectedErr, wantErr: testChallengeStoreFailedError},
		{name: "email delivery failure", email: testTestExampleCom, token: magicLinkErrorPathToken, deliveryFail: true, wantErr: "failed to send email"},
		{name: testUserauthRateRecordingFailureBestEffort, email: testTestExampleCom, token: magicLinkErrorPathToken, recordErr: injectedErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7())
			userRepo := newCoverageUserRepo(&cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), Email: tc.email})
			userRepo.failGet = tc.userLookup
			delivery := cryptoutilIdentityIdpUserauth.NewMockDeliveryService()
			delivery.SetShouldFail(tc.deliveryFail)
			challengeStore := &userauthTestChallengeStoreStub{storeErr: tc.storeErr}
			rateLimiter := &userauthTestRateLimiterStub{checkErr: tc.checkErr, recordErr: tc.recordErr}
			generator := &userauthTestOTPGeneratorStub{token: tc.token, tokenErr: tc.tokenErr}
			auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo, "https://example.test")

			challenge, err := auth.InitiateAuth(ctx, userID.String())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, challenge)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, challenge)
			require.NotEmpty(t, challengeStore.secret)
			require.NotEqual(t, tc.token, challengeStore.secret)
			require.Len(t, delivery.GetSentEmails(), 1)
		})
	}
}

func TestMagicLinkAuthenticator_VerifyAuthErrorPaths(t *testing.T) {
	t.Parallel()

	validTokenHash, err := cryptoutilIdentityIdpUserauth.HashToken(magicLinkErrorPathToken)
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
		response         string
		userLookup       bool
		wantErr          string
	}{
		{name: testInvalidChallengeIDError, challengeID: cryptoutilSharedMagic.DevSetupTestInvalidName, wantErr: testInvalidChallengeIDError},
		{
			name:        testUserauthChallengeRetrievalFailureCase,
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			retrieveErr: errors.New("missing"),
			wantErr:     testUserauthChallengeNotFoundError,
		},
		{
			name:             testUserauthExpiredChallengeCase,
			challengeID:      googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:        &cryptoutilIdentityIdpUserauth.AuthChallenge{},
			challengeExpired: true,
			deleteErr:        errors.New("delete failed"),
			wantErr:          "magic link expired",
		},
		{
			name:        "invalid token",
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:   &cryptoutilIdentityIdpUserauth.AuthChallenge{UserID: magicLinkErrorPathUser},
			secret:      validTokenHash,
			recordErr:   errors.New("record failed"),
			response:    "wrong-token",
			wantErr:     "invalid magic link token",
		},
		{
			name:        "challenge deletion failure",
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:   &cryptoutilIdentityIdpUserauth.AuthChallenge{UserID: magicLinkErrorPathUser},
			secret:      validTokenHash,
			deleteErr:   errors.New("delete failed"),
			response:    magicLinkErrorPathToken,
			wantErr:     "failed to delete challenge",
		},
		{
			name:        "user lookup failure after verification",
			challengeID: googleUuid.Must(googleUuid.NewV7()).String(),
			challenge:   &cryptoutilIdentityIdpUserauth.AuthChallenge{UserID: magicLinkErrorPathUser},
			secret:      validTokenHash,
			response:    magicLinkErrorPathToken,
			userLookup:  true,
			wantErr:     testUserauthUserNotFoundError,
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
			userRepo := newCoverageUserRepo(&cryptoutilIdentityDomain.User{})
			userRepo.failGet = tc.userLookup
			rateLimiter := &userauthTestRateLimiterStub{recordErr: tc.recordErr}
			auth := cryptoutilIdentityIdpUserauth.NewMagicLinkAuthenticator(
				&userauthTestOTPGeneratorStub{},
				cryptoutilIdentityIdpUserauth.NewMockDeliveryService(),
				challengeStore,
				rateLimiter,
				userRepo,
				"https://example.test",
			)

			user, verifyErr := auth.VerifyAuth(context.Background(), tc.challengeID, tc.response)
			require.ErrorContains(t, verifyErr, tc.wantErr)
			require.Nil(t, user)
		})
	}
}
