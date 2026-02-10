// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestClientRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-create",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	// Verify client was created.
	retrieved, err := repo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, client.ClientID, retrieved.ClientID)

	// Verify ClientSecretVersion was created (version 1, active).
	var secretVersions []cryptoutilIdentityDomain.ClientSecretVersion

	err = testDB.db.Where("client_id = ?", client.ID).Find(&secretVersions).Error
	require.NoError(t, err)
	require.Len(t, secretVersions, 1, "Expected exactly 1 initial secret version")
	require.Equal(t, 1, secretVersions[0].Version, "Expected version 1 for initial secret")
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, secretVersions[0].Status, "Expected active status")
	require.Nil(t, secretVersions[0].ExpiresAt, "Expected no expiration for initial secret")
	require.NotEmpty(t, secretVersions[0].SecretHash, "Expected non-empty secret hash")

	// Verify KeyRotationEvent was created.
	var events []cryptoutilIdentityDomain.KeyRotationEvent

	err = testDB.db.Where("key_id = ?", client.ID.String()).Find(&events).Error
	require.NoError(t, err)
	require.Len(t, events, 1, "Expected exactly 1 audit event for client creation")
	require.Equal(t, "secret_created", events[0].EventType, "Expected secret_created event type")
	require.Equal(t, "client_secret", events[0].KeyType, "Expected client_secret key type")
	require.Equal(t, client.ID.String(), events[0].KeyID, "Expected client ID in event")
	require.Equal(t, "system", events[0].Initiator, "Expected system initiator")
	require.NotNil(t, events[0].OldKeyVersion, "Expected OldKeyVersion to be set")
	require.Equal(t, 0, *events[0].OldKeyVersion, "Expected OldKeyVersion = 0")
	require.NotNil(t, events[0].NewKeyVersion, "Expected NewKeyVersion to be set")
	require.Equal(t, 1, *events[0].NewKeyVersion, "Expected NewKeyVersion = 1")
	require.NotNil(t, events[0].Success, "Expected Success to be set")
	require.True(t, *events[0].Success, "Expected successful event")
}

func TestClientRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr error
	}{
		{
			name:    "client not found",
			id:      googleUuid.Must(googleUuid.NewV7()),
			wantErr: cryptoutilIdentityAppErr.ErrClientNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, err := repo.GetByID(ctx, tc.id)
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, client)
		})
	}
}

func TestClientRepository_GetByClientID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-get",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, testClient)
	require.NoError(t, err)

	tests := []struct {
		name     string
		clientID string
		wantErr  error
	}{
		{
			name:     "client found",
			clientID: "test-client-get",
			wantErr:  nil,
		},
		{
			name:     "client not found",
			clientID: "nonexistent",
			wantErr:  cryptoutilIdentityAppErr.ErrClientNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, err := repo.GetByClientID(ctx, tc.clientID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				require.Equal(t, tc.clientID, client.ClientID)
			}
		})
	}
}

func TestClientRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-update",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	client.Name = "Updated Client"
	err = repo.Update(ctx, client)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Client", retrieved.Name)
}

func TestClientRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-delete",
		ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	err = repo.Delete(ctx, client.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, client.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrClientNotFound)
	require.Nil(t, retrieved)
}

func TestClientRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	for i := range 5 {
		client := &cryptoutilIdentityDomain.Client{
			ID:                      googleUuid.Must(googleUuid.NewV7()),
			ClientID:                "test-client-" + string(rune('0'+i)),
			ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:                    "Test Client",
			TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			AllowedGrantTypes:       []string{"authorization_code"},
			AllowedResponseTypes:    []string{"code"},
			AllowedScopes:           []string{"openid"},
			RedirectURIs:            []string{"https://example.com/callback"},
			RequirePKCE:             boolPtr(true),
			AccessTokenLifetime:     3600,
			RefreshTokenLifetime:    86400,
			IDTokenLifetime:         3600,
			Enabled:                 boolPtr(true),
		}
		err := repo.Create(ctx, client)
		require.NoError(t, err)
	}

	clients, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	require.Len(t, clients, 3)

	clients, err = repo.List(ctx, 3, 3)
	require.NoError(t, err)
	require.Len(t, clients, 2)
}

func TestClientRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := range 5 {
		client := &cryptoutilIdentityDomain.Client{
			ID:                      googleUuid.Must(googleUuid.NewV7()),
			ClientID:                "test-client-count-" + string(rune('0'+i)),
			ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:                    "Test Client",
			TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			AllowedGrantTypes:       []string{"authorization_code"},
			AllowedResponseTypes:    []string{"code"},
			AllowedScopes:           []string{"openid"},
			RedirectURIs:            []string{"https://example.com/callback"},
			RequirePKCE:             boolPtr(true),
			AccessTokenLifetime:     3600,
			RefreshTokenLifetime:    86400,
			IDTokenLifetime:         3600,
			Enabled:                 boolPtr(true),
		}
		err := repo.Create(ctx, client)
		require.NoError(t, err)
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(5), count)
}

func TestClientRepository_GetAll(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	for i := range 5 {
		client := &cryptoutilIdentityDomain.Client{
			ID:                      googleUuid.Must(googleUuid.NewV7()),
			ClientID:                "test-client-getall-" + string(rune('0'+i)),
			ClientSecret:            googleUuid.Must(googleUuid.NewV7()).String(),
			ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:                    "Test Client",
			TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			AllowedGrantTypes:       []string{"authorization_code"},
			AllowedResponseTypes:    []string{"code"},
			AllowedScopes:           []string{"openid"},
			RedirectURIs:            []string{"https://example.com/callback"},
			RequirePKCE:             boolPtr(true),
			AccessTokenLifetime:     3600,
			RefreshTokenLifetime:    86400,
			IDTokenLifetime:         3600,
			Enabled:                 boolPtr(true),
		}
		err := repo.Create(ctx, client)
		require.NoError(t, err)
	}

	clients, err := repo.GetAll(ctx)
	require.NoError(t, err)
	require.Len(t, clients, 5)
}

func TestClientRepository_RotateSecret(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	// Create a test client.
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-client-rotate-secret",
		ClientSecret:            "initial-secret-hash",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client for Secret Rotation",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}

	err := repo.Create(ctx, client)
	require.NoError(t, err)

	tests := []struct {
		name          string
		clientID      googleUuid.UUID
		newSecretHash string
		rotatedBy     string
		reason        string
		wantErr       error
		setup         func(t *testing.T)
	}{
		{
			name:          "successful_rotation",
			clientID:      client.ID,
			newSecretHash: "new-secret-hash",
			rotatedBy:     "admin@example.com",
			reason:        "Scheduled rotation",
			wantErr:       nil,
			setup:         func(_ *testing.T) {},
		},
		{
			name:          "client_not_found",
			clientID:      googleUuid.Must(googleUuid.NewV7()),
			newSecretHash: "new-secret-hash",
			rotatedBy:     "admin@example.com",
			reason:        "Test rotation",
			wantErr:       cryptoutilIdentityAppErr.ErrClientNotFound,
			setup:         func(_ *testing.T) {},
		},
		{
			name:          "empty_new_secret_hash",
			clientID:      client.ID,
			newSecretHash: "",
			rotatedBy:     "admin@example.com",
			reason:        "Test rotation",
			wantErr:       nil, // Repository allows empty hash (validation should happen at service layer)
			setup:         func(_ *testing.T) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)

			// Capture old secret before rotation.
			oldClient, err := repo.GetByID(ctx, client.ID)

			var oldSecretHash string
			if err == nil {
				oldSecretHash = oldClient.ClientSecret
			}

			err = repo.RotateSecret(ctx, tc.clientID, tc.newSecretHash, tc.rotatedBy, tc.reason)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)

				// Verify client secret was updated.
				updated, err := repo.GetByID(ctx, tc.clientID)
				require.NoError(t, err)
				require.Equal(t, tc.newSecretHash, updated.ClientSecret)

				// Verify old secret was archived to history.
				var history []cryptoutilIdentityDomain.ClientSecretHistory

				err = testDB.db.Where("client_id = ? AND secret_hash = ?", tc.clientID, oldSecretHash).
					Find(&history).Error
				require.NoError(t, err)

				if tc.name == "successful_rotation" {
					require.Len(t, history, 1, "Expected old secret in history")
					require.Equal(t, tc.clientID, history[0].ClientID)
					require.Equal(t, oldSecretHash, history[0].SecretHash)
					require.Equal(t, tc.rotatedBy, history[0].RotatedBy)
					require.Equal(t, tc.reason, history[0].Reason)
				}
			}
		})
	}
}

func TestClientRepository_GetSecretHistory(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name            string
		setupRotations  int
		expectHistoryOK bool
		wantErr         error
	}{
		{
			name:            "no_rotations",
			setupRotations:  0,
			expectHistoryOK: true,
			wantErr:         nil,
		},
		{
			name:            "single_rotation",
			setupRotations:  1,
			expectHistoryOK: true,
			wantErr:         nil,
		},
		{
			name:            "multiple_rotations",
			setupRotations:  3,
			expectHistoryOK: true,
			wantErr:         nil,
		},
		{
			name:            "client_not_found",
			setupRotations:  0,
			expectHistoryOK: true, // Returns empty list, not error
			wantErr:         nil,
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh client for this test case.
			client := &cryptoutilIdentityDomain.Client{
				ID:                      googleUuid.Must(googleUuid.NewV7()),
				ClientID:                fmt.Sprintf("test-client-history-%s", tc.name),
				ClientSecret:            "initial-secret-hash",
				ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
				Name:                    fmt.Sprintf("Test Client for %s", tc.name),
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				AllowedGrantTypes:       []string{"authorization_code"},
				AllowedResponseTypes:    []string{"code"},
				AllowedScopes:           []string{"openid"},
				RedirectURIs:            []string{"https://example.com/callback"},
				RequirePKCE:             boolPtr(true),
				AccessTokenLifetime:     3600,
				RefreshTokenLifetime:    86400,
				IDTokenLifetime:         3600,
				Enabled:                 boolPtr(true),
			}

			err := repo.Create(ctx, client)
			require.NoError(t, err)

			// Use a different client ID for "client_not_found" test.
			clientID := client.ID
			if tc.name == "client_not_found" {
				clientID = googleUuid.Must(googleUuid.NewV7())
			}

			// Setup: Perform rotations.
			for i := 0; i < tc.setupRotations; i++ {
				err := repo.RotateSecret(
					ctx,
					client.ID,
					fmt.Sprintf("secret-hash-%d", i+1),
					fmt.Sprintf("admin-%d@example.com", i+1),
					fmt.Sprintf("Rotation %d", i+1),
				)
				require.NoError(t, err)
			}

			// Test: Get secret history.
			history, err := repo.GetSecretHistory(ctx, clientID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)

				if tc.expectHistoryOK {
					// Verify history count matches rotations.
					require.Len(t, history, tc.setupRotations, "Expected history count to match rotations")

					// Verify history is ordered by rotated_at DESC.
					if len(history) > 1 {
						for i := 0; i < len(history)-1; i++ {
							require.True(t,
								history[i].RotatedAt.After(history[i+1].RotatedAt) ||
									history[i].RotatedAt.Equal(history[i+1].RotatedAt),
								"Expected history ordered by rotated_at DESC")
						}
					}

					// Verify history entries have expected fields.
					for i, h := range history {
						require.Equal(t, client.ID, h.ClientID)
						require.NotEmpty(t, h.SecretHash)
						require.NotEmpty(t, h.RotatedBy)
						require.NotEmpty(t, h.Reason)
						require.NotZero(t, h.RotatedAt)

						// Most recent rotation should be first.
						expectedRotation := tc.setupRotations - i
						require.Contains(t, h.Reason, fmt.Sprintf("Rotation %d", expectedRotation))
					}
				}
			}
		})
	}
}

func TestGenerateRandomSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		length     int
		expectErr  bool
		minEncoded int
	}{
		{
			name:       "valid_32_bytes",
			length:     32,
			expectErr:  false,
			minEncoded: 40, // Base64 encoding: 32 bytes → ~43 chars
		},
		{
			name:       "valid_16_bytes",
			length:     16,
			expectErr:  false,
			minEncoded: 20, // Base64 encoding: 16 bytes → ~21 chars
		},
		{
			name:       "valid_64_bytes",
			length:     64,
			expectErr:  false,
			minEncoded: 80, // Base64 encoding: 64 bytes → ~85 chars
		},
		{
			name:       "zero_length",
			length:     0,
			expectErr:  false, // crypto/rand.Read allows zero length
			minEncoded: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			secret, err := generateRandomSecret(tc.length)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(secret), tc.minEncoded,
					"Expected base64-encoded secret length >= %d, got %d", tc.minEncoded, len(secret))

				// Verify secret is valid base64.
				_, err := base64.URLEncoding.DecodeString(secret)
				require.NoError(t, err, "Expected valid base64 URL encoding")

				// Verify uniqueness (generate second secret, should differ).
				if tc.length > 0 {
					secret2, err := generateRandomSecret(tc.length)
					require.NoError(t, err)
					require.NotEqual(t, secret, secret2, "Expected unique secrets")
				}
			}
		})
	}
}
