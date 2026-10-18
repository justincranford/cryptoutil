// Copyright (c) 2025-2026 Justin Cranford.
package userauth_test

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity-idp/userauth"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const contactOTPTestValue = "123456"

func TestSMSOTPAuthenticator_InitiateAuthErrorPaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected failure")
	tests := []struct {
		name         string
		otp          string
		otpErr       error
		storeErr     error
		checkErr     error
		recordErr    error
		deliveryFail bool
		userLookup   bool
		wantErr      string
	}{
		{name: testUserauthRateLimitedCase, checkErr: injectedErr, wantErr: testUserauthRateLimitExceededError},
		{name: testUserauthUserLookupFailureCase, userLookup: true, wantErr: testUserauthUserNotFoundError},
		{name: testUserauthOTPGenerationFailureCase, otpErr: injectedErr, wantErr: testOTPGenerationFailedError},
		{name: testUserauthEmptyGeneratedOTPCase, wantErr: testOTPHashFailedError},
		{name: testUserauthChallengeStorageFailureCase, otp: contactOTPTestValue, storeErr: injectedErr, wantErr: testChallengeStoreFailedError},
		{name: "SMS delivery failure", otp: contactOTPTestValue, deliveryFail: true, wantErr: "failed to send SMS"},
		{name: testUserauthRateRecordingFailureBestEffort, otp: contactOTPTestValue, recordErr: injectedErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7())
			userRepo := newCoverageUserRepo(&cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+15550000000"})
			userRepo.failGet = tc.userLookup
			delivery := newMockSMSDeliveryService()
			delivery.shouldFail = tc.deliveryFail
			challengeStore := &userauthTestChallengeStoreStub{storeErr: tc.storeErr}
			rateLimiter := &userauthTestRateLimiterStub{checkErr: tc.checkErr, recordErr: tc.recordErr}
			generator := &userauthTestOTPGeneratorStub{otp: tc.otp, otpErr: tc.otpErr}
			auth := cryptoutilIdentityIdpUserauth.NewSMSOTPAuthenticator(generator, delivery, challengeStore, rateLimiter, userRepo)

			challenge, err := auth.InitiateAuth(ctx, userID.String())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, challenge)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, challenge)
			require.Len(t, delivery.sentMessages, 1)
			require.NotEmpty(t, challengeStore.secret)
			require.NotEqual(t, tc.otp, challengeStore.secret)
		})
	}
}

func TestPhoneCallOTPAuthenticator_InitiateAuthErrorPaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected failure")
	tests := []struct {
		name         string
		otp          string
		otpErr       error
		storeErr     error
		checkErr     error
		recordErr    error
		deliveryFail bool
		userLookup   bool
		wantErr      string
	}{
		{name: testUserauthRateLimitedCase, checkErr: injectedErr, wantErr: testUserauthRateLimitExceededError},
		{name: "user lookup failure", userLookup: true, wantErr: testUserauthUserNotFoundError},
		{name: "OTP generation failure", otpErr: injectedErr, wantErr: testOTPGenerationFailedError},
		{name: "empty generated OTP", wantErr: testOTPHashFailedError},
		{name: "challenge storage failure", otp: contactOTPTestValue, storeErr: injectedErr, wantErr: testChallengeStoreFailedError},
		{name: "voice call failure", otp: contactOTPTestValue, deliveryFail: true, wantErr: "failed to make voice call"},
		{name: "rate recording failure is best effort", otp: contactOTPTestValue, recordErr: injectedErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7())
			userRepo := newCoverageUserRepo(&cryptoutilIdentityDomain.User{ID: userID, Sub: userID.String(), PhoneNumber: "+15550000000"})
			userRepo.failGet = tc.userLookup
			phoneService := newMockPhoneCallService()
			phoneService.shouldFail = tc.deliveryFail
			challengeStore := &userauthTestChallengeStoreStub{storeErr: tc.storeErr}
			rateLimiter := &userauthTestRateLimiterStub{checkErr: tc.checkErr, recordErr: tc.recordErr}
			generator := &userauthTestOTPGeneratorStub{otp: tc.otp, otpErr: tc.otpErr}
			auth := cryptoutilIdentityIdpUserauth.NewPhoneCallOTPAuthenticator(generator, phoneService, challengeStore, rateLimiter, userRepo)

			challenge, err := auth.InitiateAuth(ctx, userID.String())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, challenge)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, challenge)
			require.Len(t, phoneService.madeCalls, 1)
			require.NotEmpty(t, challengeStore.secret)
			require.NotEqual(t, tc.otp, challengeStore.secret)
		})
	}
}
