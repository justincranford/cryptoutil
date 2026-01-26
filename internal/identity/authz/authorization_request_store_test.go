// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

func TestInMemoryAuthorizationRequestStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name string
		test func(t *testing.T, store cryptoutilIdentityAuthz.AuthorizationRequestStore)
	}{
		{
			name: "store and retrieve by request ID",
			test: func(t *testing.T, store cryptoutilIdentityAuthz.AuthorizationRequestStore) {
				t.Parallel()

				requestID := googleUuid.New()
				request := &cryptoutilIdentityAuthz.AuthorizationRequest{
					RequestID:           requestID,
					ClientID:            "test-client",
					RedirectURI:         "https://client.example.com/callback",
					ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
					Scope:               "openid profile",
					State:               "test-state",
					CodeChallenge:       "test-challenge",
					CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
					CreatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
					ConsentGranted:      false,
				}

				err := store.Store(ctx, request)
				require.NoError(t, err)

				retrieved, err := store.GetByRequestID(ctx, requestID)
				require.NoError(t, err)
				require.NotNil(t, retrieved)
				require.Equal(t, request.RequestID, retrieved.RequestID)
				require.Equal(t, request.ClientID, retrieved.ClientID)
			},
		},
		{
			name: "store and retrieve by code",
			test: func(t *testing.T, store cryptoutilIdentityAuthz.AuthorizationRequestStore) {
				t.Parallel()

				requestID := googleUuid.New()
				authCode := "test-authorization-code"
				request := &cryptoutilIdentityAuthz.AuthorizationRequest{
					RequestID:           requestID,
					ClientID:            "test-client",
					RedirectURI:         "https://client.example.com/callback",
					ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
					Scope:               "openid profile",
					State:               "test-state",
					Code:                authCode,
					CodeChallenge:       "test-challenge",
					CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
					CreatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
					ConsentGranted:      true,
				}

				err := store.Store(ctx, request)
				require.NoError(t, err)

				retrieved, err := store.GetByCode(ctx, authCode)
				require.NoError(t, err)
				require.NotNil(t, retrieved)
				require.Equal(t, request.Code, retrieved.Code)
				require.Equal(t, request.ClientID, retrieved.ClientID)
			},
		},
		{
			name: "update authorization request",
			test: func(t *testing.T, store cryptoutilIdentityAuthz.AuthorizationRequestStore) {
				t.Parallel()

				requestID := googleUuid.New()
				request := &cryptoutilIdentityAuthz.AuthorizationRequest{
					RequestID:           requestID,
					ClientID:            "test-client",
					RedirectURI:         "https://client.example.com/callback",
					ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
					Scope:               "openid profile",
					State:               "test-state",
					CodeChallenge:       "test-challenge",
					CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
					CreatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
					ConsentGranted:      false,
				}

				err := store.Store(ctx, request)
				require.NoError(t, err)

				// Update with consent granted.
				request.ConsentGranted = true
				request.Code = "test-code"

				err = store.Update(ctx, request)
				require.NoError(t, err)

				retrieved, err := store.GetByRequestID(ctx, requestID)
				require.NoError(t, err)
				require.True(t, retrieved.ConsentGranted)
				require.Equal(t, "test-code", retrieved.Code)
			},
		},
		{
			name: "delete authorization request",
			test: func(t *testing.T, store cryptoutilIdentityAuthz.AuthorizationRequestStore) {
				t.Parallel()

				requestID := googleUuid.New()
				request := &cryptoutilIdentityAuthz.AuthorizationRequest{
					RequestID:           requestID,
					ClientID:            "test-client",
					RedirectURI:         "https://client.example.com/callback",
					ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
					Scope:               "openid profile",
					State:               "test-state",
					CodeChallenge:       "test-challenge",
					CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
					CreatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
					ConsentGranted:      false,
				}

				err := store.Store(ctx, request)
				require.NoError(t, err)

				err = store.Delete(ctx, requestID)
				require.NoError(t, err)

				_, err = store.GetByRequestID(ctx, requestID)
				require.Error(t, err)
			},
		},
		{
			name: "expired request returns error",
			test: func(t *testing.T, store cryptoutilIdentityAuthz.AuthorizationRequestStore) {
				t.Parallel()

				requestID := googleUuid.New()
				request := &cryptoutilIdentityAuthz.AuthorizationRequest{
					RequestID:           requestID,
					ClientID:            "test-client",
					RedirectURI:         "https://client.example.com/callback",
					ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
					Scope:               "openid profile",
					State:               "test-state",
					CodeChallenge:       "test-challenge",
					CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
					CreatedAt:           time.Now().UTC().Add(-10 * time.Minute),
					ExpiresAt:           time.Now().UTC().Add(-5 * time.Minute),
					ConsentGranted:      false,
				}

				err := store.Store(ctx, request)
				require.NoError(t, err)

				_, err = store.GetByRequestID(ctx, requestID)
				require.Error(t, err)
				require.Contains(t, err.Error(), "expired")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
			tc.test(t, store)
		})
	}
}
