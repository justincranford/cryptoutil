// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/base64"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

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
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
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
				AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
				AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
				AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID},
				RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
				RequirePKCE:             boolPtr(true),
				AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
				RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
				IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
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
			length:     cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
			expectErr:  false,
			minEncoded: 40, // Base64 encoding: 32 bytes → ~43 chars
		},
		{
			name:       "valid_16_bytes",
			length:     cryptoutilSharedMagic.RealmMinTokenLengthBytes,
			expectErr:  false,
			minEncoded: cryptoutilSharedMagic.MaxErrorDisplay, // Base64 encoding: 16 bytes → ~21 chars
		},
		{
			name:       "valid_64_bytes",
			length:     cryptoutilSharedMagic.MinSerialNumberBits,
			expectErr:  false,
			minEncoded: cryptoutilSharedMagic.LineWidth, // Base64 encoding: 64 bytes → ~85 chars
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
