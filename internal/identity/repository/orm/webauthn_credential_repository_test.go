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
)

// Validates requirements:
// - R07-01: Repository tests run against SQLite
// - R07-03: Repository tests validate concurrent operations.
func TestNewWebAuthnCredentialRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		db                any
		wantError         bool
		wantErrorContains string
	}{
		{
			name:      "valid database connection creates repository successfully",
			db:        setupTestDB(t),
			wantError: false,
		},
		{
			name:              "nil database connection returns error",
			db:                nil,
			wantError:         true,
			wantErrorContains: "database connection cannot be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				repo *WebAuthnCredentialRepository
				err  error
			)

			if tc.db != nil {
				repo, err = NewWebAuthnCredentialRepository(tc.db.(*testDB).db) //nolint:errcheck // Error checked conditionally below based on test case expectations
			} else {
				repo, err = NewWebAuthnCredentialRepository(nil) //nolint:errcheck // Error checked conditionally below based on test case expectations
			}

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrorContains)
				require.Nil(t, repo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, repo)
			}
		})
	}
}

// Validates requirements:
// - R07-04: Repository tests validate GORM transaction patterns.
func TestWebAuthnCredentialRepository_StoreCredential(t *testing.T) {
	t.Parallel()

	invalidUserID := googleUuid.Must(googleUuid.NewV7()).String() + "-invalid"
	tests := []struct {
		name              string
		credential        *Credential
		wantError         bool
		wantErrorContains string
	}{
		{
			name: "store new credential succeeds",
			credential: &Credential{
				ID:              "new-cred-id-webauthn-store-new-1",
				UserID:          googleUuid.Must(googleUuid.NewV7()).String(),
				Type:            CredentialTypePasskey,
				PublicKey:       []byte("public-key-data"),
				AttestationType: "none",
				AAGUID:          []byte{1, 2, 3, 4},
				SignCount:       0,
				CreatedAt:       time.Now().UTC(),
				LastUsedAt:      time.Now().UTC(),
				Metadata: map[string]any{
					"device_name": "Test Device",
				},
			},
			wantError: false,
		},
		{
			name:              "nil credential returns error",
			credential:        nil,
			wantError:         true,
			wantErrorContains: "credential cannot be nil",
		},
		{
			name: "invalid user ID returns error",
			credential: &Credential{
				ID:              "invalid-user-id-cred",
				UserID:          invalidUserID,
				Type:            CredentialTypePasskey,
				PublicKey:       []byte("public-key-data"),
				AttestationType: "none",
				SignCount:       0,
				CreatedAt:       time.Now().UTC(),
				LastUsedAt:      time.Now().UTC(),
			},
			wantError:         true,
			wantErrorContains: "invalid user ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo, err := NewWebAuthnCredentialRepository(db.db)
			require.NoError(t, err)

			// Seed test user if credential has valid user ID.
			if tc.credential != nil && tc.credential.UserID != invalidUserID {
				seedTestUser(ctx, t, db.db, tc.credential.UserID)
			}

			err = repo.StoreCredential(ctx, tc.credential)

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrorContains)
			} else {
				require.NoError(t, err)

				// Verify credential was stored.
				retrieved, err := repo.GetCredential(ctx, tc.credential.ID)
				require.NoError(t, err)
				require.Equal(t, tc.credential.ID, retrieved.ID)
				require.Equal(t, tc.credential.UserID, retrieved.UserID)
				require.Equal(t, tc.credential.Type, retrieved.Type)
				require.Equal(t, tc.credential.PublicKey, retrieved.PublicKey)
				require.Equal(t, tc.credential.SignCount, retrieved.SignCount)
			}
		})
	}
}

func TestWebAuthnCredentialRepository_UpdateCredential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	repo, err := NewWebAuthnCredentialRepository(db.db)
	require.NoError(t, err)

	// Create initial credential.
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	cred := &Credential{
		ID:              "update-cred-id-webauthn-update-1",
		UserID:          userID,
		Type:            CredentialTypePasskey,
		PublicKey:       []byte("public-key-data"),
		AttestationType: "none",
		AAGUID:          []byte{1, 2, 3, 4},
		SignCount:       5,
		CreatedAt:       time.Now().UTC(),
		LastUsedAt:      time.Now().UTC(),
		Metadata: map[string]any{
			"device_name": "Test Device",
		},
	}

	err = repo.StoreCredential(ctx, cred)
	require.NoError(t, err)

	// Update sign counter (replay prevention).
	cred.SignCount = 10
	cred.LastUsedAt = time.Now().UTC().Add(1 * time.Hour)

	err = repo.StoreCredential(ctx, cred)
	require.NoError(t, err)

	// Verify updated credential.
	retrieved, err := repo.GetCredential(ctx, cred.ID)
	require.NoError(t, err)
	require.Equal(t, uint32(10), retrieved.SignCount)
}

func TestWebAuthnCredentialRepository_GetCredential(t *testing.T) {
	t.Parallel()

	userID := googleUuid.Must(googleUuid.NewV7()).String()

	tests := []struct {
		name         string
		credentialID string
		setupCred    *Credential
		wantError    bool
	}{
		{
			name:         "get existing credential succeeds",
			credentialID: "existing-cred-id-webauthn-get-1",
			setupCred: &Credential{
				ID:              "existing-cred-id-webauthn-get-1",
				UserID:          userID,
				Type:            CredentialTypePasskey,
				PublicKey:       []byte("public-key-data"),
				AttestationType: "none",
				SignCount:       0,
				CreatedAt:       time.Now().UTC(),
				LastUsedAt:      time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name:         "get non-existent credential returns not found error",
			credentialID: "non-existent-cred-id",
			setupCred:    nil,
			wantError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo, err := NewWebAuthnCredentialRepository(db.db)
			require.NoError(t, err)

			// Setup credential if provided (seed user first).
			if tc.setupCred != nil {
				seedTestUser(ctx, t, db.db, tc.setupCred.UserID)
				err = repo.StoreCredential(ctx, tc.setupCred)
				require.NoError(t, err)
			}

			// Get credential.
			retrieved, err := repo.GetCredential(ctx, tc.credentialID)

			if tc.wantError {
				require.Error(t, err)
				require.Nil(t, retrieved)
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrCredentialNotFound)
			} else {
				require.NoError(t, err)
				require.NotNil(t, retrieved)
				require.Equal(t, tc.credentialID, retrieved.ID)
			}
		})
	}
}

func TestWebAuthnCredentialRepository_GetUserCredentials(t *testing.T) {
	t.Parallel()

	userID1 := googleUuid.Must(googleUuid.NewV7()).String()
	userID2 := googleUuid.Must(googleUuid.NewV7()).String()
	invalidUserID := googleUuid.Must(googleUuid.NewV7()).String() + "-invalid"

	tests := []struct {
		name       string
		userID     string
		setupCreds []*Credential
		wantCount  int
		wantError  bool
	}{
		{
			name:   "get user with multiple credentials succeeds",
			userID: userID1,
			setupCreds: []*Credential{
				{
					ID:              "user-cred-1-webauthn-list-1",
					UserID:          userID1,
					Type:            CredentialTypePasskey,
					PublicKey:       []byte("public-key-1"),
					AttestationType: "none",
					SignCount:       0,
					CreatedAt:       time.Now().UTC(),
					LastUsedAt:      time.Now().UTC(),
				},
				{
					ID:              "user-cred-2-webauthn-list-2",
					UserID:          userID1,
					Type:            CredentialTypePasskey,
					PublicKey:       []byte("public-key-2"),
					AttestationType: "none",
					SignCount:       0,
					CreatedAt:       time.Now().UTC().Add(1 * time.Minute),
					LastUsedAt:      time.Now().UTC().Add(1 * time.Minute),
				},
			},
			wantCount: 2,
			wantError: false,
		},
		{
			name:       "get user with no credentials returns empty list",
			userID:     userID2,
			setupCreds: nil,
			wantCount:  0,
			wantError:  false,
		},
		{
			name:       "invalid user ID returns error",
			userID:     invalidUserID,
			setupCreds: nil,
			wantCount:  0,
			wantError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo, err := NewWebAuthnCredentialRepository(db.db)
			require.NoError(t, err)

			// Seed user for credential tests.
			if tc.userID != invalidUserID {
				seedTestUser(ctx, t, db.db, tc.userID)
			}

			// Setup credentials if provided.
			if tc.setupCreds != nil {
				for _, cred := range tc.setupCreds {
					err = repo.StoreCredential(ctx, cred)
					require.NoError(t, err)
				}
			}

			// Get user credentials.
			creds, err := repo.GetUserCredentials(ctx, tc.userID)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, creds, tc.wantCount)

				// Verify ordering (most recent first).
				if len(creds) > 1 {
					require.True(t, creds[0].CreatedAt.After(creds[1].CreatedAt) || creds[0].CreatedAt.Equal(creds[1].CreatedAt))
				}
			}
		})
	}
}

func TestWebAuthnCredentialRepository_DeleteCredential(t *testing.T) {
	t.Parallel()

	userID1 := googleUuid.Must(googleUuid.NewV7()).String()
	tests := []struct {
		name         string
		credentialID string
		setupCred    *Credential
		wantError    bool
	}{
		{
			name:         "delete existing credential succeeds",
			credentialID: "delete-cred-id-webauthn-delete-1",
			setupCred: &Credential{
				ID:              "delete-cred-id-webauthn-delete-1",
				UserID:          userID1,
				Type:            CredentialTypePasskey,
				PublicKey:       []byte("public-key-data"),
				AttestationType: "none",
				SignCount:       0,
				CreatedAt:       time.Now().UTC(),
				LastUsedAt:      time.Now().UTC(),
			},
			wantError: false,
		},
		{
			name:         "delete non-existent credential returns not found error",
			credentialID: "non-existent-cred-id",
			setupCred:    nil,
			wantError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo, err := NewWebAuthnCredentialRepository(db.db)
			require.NoError(t, err)

			// Setup credential if provided (seed user first).
			if tc.setupCred != nil {
				seedTestUser(ctx, t, db.db, tc.setupCred.UserID)
				err = repo.StoreCredential(ctx, tc.setupCred)
				require.NoError(t, err)
			}

			// Delete credential.
			err = repo.DeleteCredential(ctx, tc.credentialID)

			if tc.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrCredentialNotFound)
			} else {
				require.NoError(t, err)

				// Verify credential is deleted.
				_, err = repo.GetCredential(ctx, tc.credentialID)
				require.Error(t, err)
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrCredentialNotFound)
			}
		})
	}
}

func TestWebAuthnCredentialRepository_CounterIncrement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	repo, err := NewWebAuthnCredentialRepository(db.db)
	require.NoError(t, err)

	// Seed user first.
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	seedTestUser(ctx, t, db.db, userID)

	// Create credential with initial counter.
	cred := &Credential{
		ID:              "counter-cred-id-webauthn-counter-1",
		UserID:          userID,
		Type:            CredentialTypePasskey,
		PublicKey:       []byte("public-key-data"),
		AttestationType: "none",
		AAGUID:          []byte{1, 2, 3, 4},
		SignCount:       5,
		CreatedAt:       time.Now().UTC(),
		LastUsedAt:      time.Now().UTC(),
	}

	err = repo.StoreCredential(ctx, cred)
	require.NoError(t, err)

	// Simulate authentication (counter increment).
	for i := uint32(6); i <= 10; i++ {
		cred.SignCount = i
		cred.LastUsedAt = time.Now().UTC().Add(time.Duration(i) * time.Second)

		err = repo.StoreCredential(ctx, cred)
		require.NoError(t, err)

		// Verify counter incremented.
		retrieved, err := repo.GetCredential(ctx, cred.ID)
		require.NoError(t, err)
		require.Equal(t, uint32(i), retrieved.SignCount)
	}

	// Final verification.
	final, err := repo.GetCredential(ctx, cred.ID)
	require.NoError(t, err)
	require.Equal(t, uint32(10), final.SignCount)
}
