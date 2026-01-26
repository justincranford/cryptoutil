// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestAuthorizationRequestRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthorizationRequestRepository(testDB.db)

	request := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		ClientID:            "test-client-id",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		Scope:               "openid profile email",
		State:               "random-state",
		Nonce:               "random-nonce",
		CodeChallenge:       "challenge-hash",
		CodeChallengeMethod: "S256",
		Code:                "auth-code-12345",
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
		ConsentGranted:      false,
		Used:                false,
	}

	err := repo.Create(context.Background(), request)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, request.ID)

	retrieved, err := repo.GetByID(context.Background(), request.ID)
	require.NoError(t, err)
	require.Equal(t, request.ClientID, retrieved.ClientID)
	require.Equal(t, request.Code, retrieved.Code)
	require.Equal(t, request.CodeChallengeMethod, retrieved.CodeChallengeMethod)
}

func TestAuthorizationRequestRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthorizationRequestRepository(testDB.db)

	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	_, err := repo.GetByID(context.Background(), nonExistentID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound)
}

func TestAuthorizationRequestRepository_GetByCode(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthorizationRequestRepository(testDB.db)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "authorization_request_found",
			setup: func() string {
				request := &cryptoutilIdentityDomain.AuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					ClientID:            "client-123",
					RedirectURI:         "https://example.com/callback",
					ResponseType:        "code",
					CodeChallenge:       "challenge",
					CodeChallengeMethod: "S256",
					Code:                "test-code-123",
					CreatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
				}
				err := repo.Create(context.Background(), request)
				require.NoError(t, err)

				return request.Code
			},
			wantErr: nil,
		},
		{
			name: "authorization_request_not_found",
			setup: func() string {
				return "nonexistent-code"
			},
			wantErr: cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code := tc.setup()
			_, err := repo.GetByCode(context.Background(), code)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationRequestRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthorizationRequestRepository(testDB.db)

	request := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		ClientID:            "client-456",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		Code:                "update-code",
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
		ConsentGranted:      false,
	}
	err := repo.Create(context.Background(), request)
	require.NoError(t, err)

	request.ConsentGranted = true
	request.Used = true
	err = repo.Update(context.Background(), request)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), request.ID)
	require.NoError(t, err)
	require.True(t, retrieved.ConsentGranted.Bool())
	require.True(t, retrieved.Used.Bool())
}

func TestAuthorizationRequestRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthorizationRequestRepository(testDB.db)

	request := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		ClientID:            "client-789",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		Code:                "delete-code",
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	err := repo.Create(context.Background(), request)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), request.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), request.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthorizationRequestNotFound)
}

func TestAuthorizationRequestRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthorizationRequestRepository(testDB.db)

	// Create expired requests.
	for i := 0; i < 3; i++ {
		expiredRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
			ID:                  googleUuid.Must(googleUuid.NewV7()),
			ClientID:            "client-expired",
			RedirectURI:         "https://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "challenge",
			CodeChallengeMethod: "S256",
			Code:                "expired-code-" + string(rune('a'+i)),
			CreatedAt:           time.Now().UTC().Add(-20 * time.Minute),
			ExpiresAt:           time.Now().UTC().Add(-10 * time.Minute), // Expired.
		}
		err := repo.Create(context.Background(), expiredRequest)
		require.NoError(t, err)
	}

	// Create valid (not expired) requests.
	for i := 0; i < 2; i++ {
		validRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
			ID:                  googleUuid.Must(googleUuid.NewV7()),
			ClientID:            "client-valid",
			RedirectURI:         "https://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "challenge",
			CodeChallengeMethod: "S256",
			Code:                "valid-code-" + string(rune('a'+i)),
			CreatedAt:           time.Now().UTC(),
			ExpiresAt:           time.Now().UTC().Add(10 * time.Minute), // Not expired.
		}
		err := repo.Create(context.Background(), validRequest)
		require.NoError(t, err)
	}

	deletedCount, err := repo.DeleteExpired(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(3), deletedCount)
}
