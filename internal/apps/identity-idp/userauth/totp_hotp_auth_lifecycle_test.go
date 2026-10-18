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

const testOTPSecret = "JBSWY3DPEHPK3PXP"

var (
	errOTPChallengeStore = errors.New("injected challenge store failure")
	errOTPRetrieve       = errors.New("injected challenge retrieval failure")
	errOTPDelete         = errors.New("injected challenge deletion failure")
	errOTPCounterGet     = errors.New("injected counter get failure")
	errOTPCounterSet     = errors.New("injected counter set failure")
)

type scriptedOTPChallengeStore struct {
	challenge   *cryptoutilIdentityIdpUserauth.AuthChallenge
	secret      string
	storeErr    error
	retrieveErr error
	deleteErr   error
	deleteCalls int
}

func (s *scriptedOTPChallengeStore) Store(
	_ context.Context,
	challenge *cryptoutilIdentityIdpUserauth.AuthChallenge,
	secret string,
) error {
	if s.storeErr != nil {
		return s.storeErr
	}

	s.challenge = challenge
	s.secret = secret

	return nil
}

func (s *scriptedOTPChallengeStore) Retrieve(
	_ context.Context,
	_ googleUuid.UUID,
) (*cryptoutilIdentityIdpUserauth.AuthChallenge, string, error) {
	if s.retrieveErr != nil {
		return nil, "", s.retrieveErr
	}

	return s.challenge, s.secret, nil
}

func (s *scriptedOTPChallengeStore) Update(
	_ context.Context,
	_ *cryptoutilIdentityIdpUserauth.AuthChallenge,
) error {
	return nil
}

func (s *scriptedOTPChallengeStore) Delete(_ context.Context, _ googleUuid.UUID) error {
	s.deleteCalls++

	return s.deleteErr
}

type scriptedOTPCounterStore struct {
	counter uint64
	getErr  error
	setErr  error
}

func (s *scriptedOTPCounterStore) GetCounter(context.Context, string) (uint64, error) {
	return s.counter, s.getErr
}

func (s *scriptedOTPCounterStore) IncrementCounter(context.Context, string) (uint64, error) {
	s.counter++

	return s.counter, nil
}

func (s *scriptedOTPCounterStore) SetCounter(_ context.Context, _ string, counter uint64) error {
	if s.setErr != nil {
		return s.setErr
	}

	s.counter = counter

	return nil
}

func TestTOTPAuthenticator_InitiateAuthStoreFailure(t *testing.T) {
	t.Parallel()

	store := &scriptedOTPChallengeStore{storeErr: errOTPChallengeStore}
	auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, nil)

	challenge, err := auth.InitiateAuth(context.Background(), googleUuid.Must(googleUuid.NewV7()).String())
	require.Nil(t, challenge)
	require.ErrorIs(t, err, errOTPChallengeStore)
}

func TestHOTPAuthenticator_InitiateAuthStoreFailure(t *testing.T) {
	t.Parallel()

	store := &scriptedOTPChallengeStore{storeErr: errOTPChallengeStore}
	auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, nil, &scriptedOTPCounterStore{})

	challenge, err := auth.InitiateAuth(context.Background(), googleUuid.Must(googleUuid.NewV7()).String())
	require.Nil(t, challenge)
	require.ErrorIs(t, err, errOTPChallengeStore)
}

func TestTOTPAuthenticator_VerifyAuthLifecycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expired       bool
		retrieveErr   error
		deleteErr     error
		userExists    bool
		validResponse bool
		wantErr       string
	}{
		{name: testOTPValidResponseCase, userExists: true, validResponse: true},
		{name: testCaseChallengeRetrievalFails, retrieveErr: errOTPRetrieve, wantErr: testChallengeNotFound},
		{name: testCaseExpiredChallenge, expired: true, wantErr: testChallengeExpired},
		{name: testOTPExpiredChallengeCleanupFailsCase, expired: true, deleteErr: errOTPDelete, wantErr: testChallengeExpired},
		{name: testOTPInvalidResponseCase, wantErr: "invalid TOTP code"},
		{name: testOTPSingleUseCleanupFailsCase, deleteErr: errOTPDelete, userExists: true, validResponse: true},
		{name: testCaseUserLookupFails, validResponse: true, wantErr: testOTPFailedToGetUserError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			challengeID := googleUuid.Must(googleUuid.NewV7())
			userID := googleUuid.Must(googleUuid.NewV7()).String()

			expiresAt := time.Now().UTC().Add(time.Hour)
			if tc.expired {
				expiresAt = time.Now().UTC().Add(-time.Hour)
			}

			store := &scriptedOTPChallengeStore{
				challenge: &cryptoutilIdentityIdpUserauth.AuthChallenge{
					ID:        challengeID,
					UserID:    userID,
					ExpiresAt: expiresAt,
				},
				secret:      testOTPSecret,
				retrieveErr: tc.retrieveErr,
				deleteErr:   tc.deleteErr,
			}
			repo := newMockSMSUserRepo()

			user := &cryptoutilIdentityDomain.User{Sub: userID}
			if tc.userExists {
				repo.AddUser(user)
			}

			auth := cryptoutilIdentityIdpUserauth.NewTOTPAuthenticator("test-issuer", store, repo)
			response := cryptoutilSharedMagic.DevSetupTestInvalidName

			if tc.validResponse {
				var err error

				response, err = auth.GenerateTOTP(ctx, store.secret)
				require.NoError(t, err)
			}

			result, err := auth.VerifyAuth(ctx, challengeID.String(), response)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Same(t, user, result)
			}

			if tc.expired || tc.validResponse {
				require.Equal(t, 1, store.deleteCalls)
			} else {
				require.Zero(t, store.deleteCalls)
			}
		})
	}
}

func TestHOTPAuthenticator_VerifyAuthLifecycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expired       bool
		retrieveErr   error
		deleteErr     error
		counterGetErr error
		counterSetErr error
		invalidSecret bool
		userExists    bool
		validResponse bool
		wantErr       string
	}{
		{name: testOTPValidResponseCase, userExists: true, validResponse: true},
		{name: testCaseChallengeRetrievalFails, retrieveErr: errOTPRetrieve, wantErr: testChallengeNotFound},
		{name: testOTPExpiredChallengeCleanupFailsCase, expired: true, deleteErr: errOTPDelete, wantErr: testChallengeExpired},
		{name: "invalid response", wantErr: "invalid HOTP code"},
		{name: "counter lookup fails", validResponse: true, counterGetErr: errOTPCounterGet, wantErr: testHOTPValidationError},
		{name: "invalid secret", invalidSecret: true, wantErr: testHOTPValidationError},
		{name: "counter update fails", validResponse: true, counterSetErr: errOTPCounterSet, wantErr: testHOTPValidationError},
		{name: testOTPSingleUseCleanupFailsCase, deleteErr: errOTPDelete, userExists: true, validResponse: true},
		{name: testCaseUserLookupFails, validResponse: true, wantErr: testOTPFailedToGetUserError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			challengeID := googleUuid.Must(googleUuid.NewV7())
			userID := googleUuid.Must(googleUuid.NewV7()).String()

			expiresAt := time.Now().UTC().Add(time.Hour)
			if tc.expired {
				expiresAt = time.Now().UTC().Add(-time.Hour)
			}

			secret := testOTPSecret
			if tc.invalidSecret {
				secret = "invalid base32!"
			}

			store := &scriptedOTPChallengeStore{
				challenge: &cryptoutilIdentityIdpUserauth.AuthChallenge{
					ID:        challengeID,
					UserID:    userID,
					ExpiresAt: expiresAt,
				},
				secret:      secret,
				retrieveErr: tc.retrieveErr,
				deleteErr:   tc.deleteErr,
			}
			counterStore := &scriptedOTPCounterStore{
				counter: 0,
				getErr:  tc.counterGetErr,
				setErr:  tc.counterSetErr,
			}
			repo := newMockSMSUserRepo()

			user := &cryptoutilIdentityDomain.User{Sub: userID}
			if tc.userExists {
				repo.AddUser(user)
			}

			auth := cryptoutilIdentityIdpUserauth.NewHOTPAuthenticator("test-issuer", store, repo, counterStore)
			response := cryptoutilSharedMagic.DevSetupTestInvalidName

			if tc.validResponse {
				var err error

				response, err = auth.GenerateHOTP(ctx, secret, counterStore.counter)
				if tc.invalidSecret {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}

			result, err := auth.VerifyAuth(ctx, challengeID.String(), response)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Same(t, user, result)
			}

			if tc.expired || (tc.validResponse && tc.counterGetErr == nil && tc.counterSetErr == nil && !tc.invalidSecret) {
				require.Equal(t, 1, store.deleteCalls)
			} else {
				require.Zero(t, store.deleteCalls)
			}
		})
	}
}
