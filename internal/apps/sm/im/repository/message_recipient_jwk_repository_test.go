// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsSmImDomain "cryptoutil/internal/apps/sm/im/domain"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

func TestMessageRecipientJWKRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		jwk     *cryptoutilAppsSmImDomain.MessageRecipientJWK
		wantErr bool
	}{
		{
			name: "valid JWK creation",
			jwk: &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           *testJWKGenService.GenerateUUIDv7(),
				RecipientID:  *testJWKGenService.GenerateUUIDv7(),
				MessageID:    *testJWKGenService.GenerateUUIDv7(),
				EncryptedJWK: generateTestJWK(t),
			},
			wantErr: false,
		},
		{
			name: "empty JWK field",
			jwk: &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           *testJWKGenService.GenerateUUIDv7(),
				RecipientID:  *testJWKGenService.GenerateUUIDv7(),
				MessageID:    *testJWKGenService.GenerateUUIDv7(),
				EncryptedJWK: "",
			},
			wantErr: false, // Repository no longer validates JWK content (validation moved to handler)
		},
		{
			name: "large JWK payload",
			jwk: &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           *testJWKGenService.GenerateUUIDv7(),
				RecipientID:  *testJWKGenService.GenerateUUIDv7(),
				MessageID:    *testJWKGenService.GenerateUUIDv7(),
				EncryptedJWK: `{"kty":"RSA","n":"` + string(make([]byte, cryptoutilSharedMagic.DefaultMetricsBatchSize)) + `","e":"AQAB"}`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create unique copy of JWK for this test to avoid shared mutations
			testJWK := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           tt.jwk.ID,
				RecipientID:  tt.jwk.RecipientID,
				MessageID:    tt.jwk.MessageID,
				EncryptedJWK: tt.jwk.EncryptedJWK,
			}

			repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)
			err := repo.Create(ctx, testJWK)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify retrieval and decryption works
			retrieved, err := repo.FindByRecipientAndMessage(ctx, testJWK.RecipientID, testJWK.MessageID)
			require.NoError(t, err)
			require.Equal(t, testJWK.EncryptedJWK, retrieved.EncryptedJWK, "JWK should decrypt to original value")
			require.Equal(t, testJWK.ID, retrieved.ID)
			require.Equal(t, testJWK.RecipientID, retrieved.RecipientID)
			require.Equal(t, testJWK.MessageID, retrieved.MessageID)

			// Cleanup
			require.NoError(t, repo.Delete(ctx, testJWK.ID))
		})
	}
}

func TestMessageRecipientJWKRepository_FindByRecipientAndMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

	tests := []struct {
		name        string
		setupJWK    *cryptoutilAppsSmImDomain.MessageRecipientJWK // If non-nil, create this before test
		recipientID googleUuid.UUID
		messageID   googleUuid.UUID
		wantErr     bool
	}{
		{
			name: "found existing JWK",
			setupJWK: &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           *testJWKGenService.GenerateUUIDv7(),
				RecipientID:  *testJWKGenService.GenerateUUIDv7(),
				MessageID:    *testJWKGenService.GenerateUUIDv7(),
				EncryptedJWK: generateTestJWK(t),
			},
			wantErr: false,
		},
		{
			name:        "nonexistent recipient",
			recipientID: *testJWKGenService.GenerateUUIDv7(),
			messageID:   *testJWKGenService.GenerateUUIDv7(),
			wantErr:     true,
		},
		{
			name:        "nonexistent message",
			recipientID: *testJWKGenService.GenerateUUIDv7(),
			messageID:   *testJWKGenService.GenerateUUIDv7(),
			wantErr:     true,
		},
		{
			name:        "both nonexistent",
			recipientID: *testJWKGenService.GenerateUUIDv7(),
			messageID:   *testJWKGenService.GenerateUUIDv7(),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup: Create JWK if test requires it
			if tt.setupJWK != nil {
				require.NoError(t, repo.Create(ctx, tt.setupJWK))

				defer func() { _ = repo.Delete(ctx, tt.setupJWK.ID) }()

				// Use IDs from setup JWK
				tt.recipientID = tt.setupJWK.RecipientID
				tt.messageID = tt.setupJWK.MessageID
			}

			retrieved, err := repo.FindByRecipientAndMessage(ctx, tt.recipientID, tt.messageID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, retrieved)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, retrieved)
			require.Equal(t, tt.setupJWK.EncryptedJWK, retrieved.EncryptedJWK)
		})
	}
}

func TestMessageRecipientJWKRepository_FindByMessageID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

	tests := []struct {
		name      string
		messageID googleUuid.UUID
		wantCount int
		wantErr   bool
	}{
		{
			name:      "find all recipients for message",
			messageID: *testJWKGenService.GenerateUUIDv7(),
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "nonexistent message",
			messageID: *testJWKGenService.GenerateUUIDv7(),
			wantCount: 0,
			wantErr:   false, // No error, just empty result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test JWKs for this specific test case
			var createdJWKs []*cryptoutilAppsSmImDomain.MessageRecipientJWK

			if tt.wantCount > 0 {
				for i := 0; i < tt.wantCount; i++ {
					jwk := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
						ID:           *testJWKGenService.GenerateUUIDv7(),
						RecipientID:  *testJWKGenService.GenerateUUIDv7(),
						MessageID:    tt.messageID,
						EncryptedJWK: generateTestJWK(t),
					}
					require.NoError(t, repo.Create(ctx, jwk))
					createdJWKs = append(createdJWKs, jwk)
				}
			}

			// Cleanup after test
			defer func() {
				for _, jwk := range createdJWKs {
					_ = repo.Delete(ctx, jwk.ID)
				}
			}()

			retrieved, err := repo.FindByMessageID(ctx, tt.messageID)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Len(t, retrieved, tt.wantCount)

			// Verify all JWKs decrypted correctly
			if tt.wantCount > 0 {
				expectedJWKs := make(map[string]bool)
				for _, jwk := range createdJWKs {
					expectedJWKs[jwk.EncryptedJWK] = true
				}

				for _, retrieved := range retrieved {
					require.True(t, expectedJWKs[retrieved.EncryptedJWK], "Retrieved JWK should match one of the created JWKs")
				}
			}
		})
	}
}

func TestMessageRecipientJWKRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

	// Create test JWK
	jwk := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
		ID:           *testJWKGenService.GenerateUUIDv7(),
		RecipientID:  *testJWKGenService.GenerateUUIDv7(),
		MessageID:    *testJWKGenService.GenerateUUIDv7(),
		EncryptedJWK: generateTestJWK(t),
	}
	require.NoError(t, repo.Create(ctx, jwk))

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr bool
	}{
		{
			name:    "delete existing JWK",
			id:      jwk.ID,
			wantErr: false,
		},
		{
			name:    "delete nonexistent JWK (idempotent)",
			id:      *testJWKGenService.GenerateUUIDv7(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repo.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify deletion
			if tt.id == jwk.ID {
				_, err := repo.FindByRecipientAndMessage(ctx, jwk.RecipientID, jwk.MessageID)
				require.Error(t, err, "Should not find deleted JWK")
			}
		})
	}
}

func TestMessageRecipientJWKRepository_DeleteByMessageID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

	messageID := *testJWKGenService.GenerateUUIDv7()

	// Create multiple JWKs for same message
	jwks := []*cryptoutilAppsSmImDomain.MessageRecipientJWK{
		{
			ID:           *testJWKGenService.GenerateUUIDv7(),
			RecipientID:  *testJWKGenService.GenerateUUIDv7(),
			MessageID:    messageID,
			EncryptedJWK: generateTestJWK(t),
		},
		{
			ID:           *testJWKGenService.GenerateUUIDv7(),
			RecipientID:  *testJWKGenService.GenerateUUIDv7(),
			MessageID:    messageID,
			EncryptedJWK: generateTestJWK(t),
		},
	}

	for _, jwk := range jwks {
		require.NoError(t, repo.Create(ctx, jwk))
	}

	tests := []struct {
		name      string
		messageID googleUuid.UUID
		wantErr   bool
	}{
		{
			name:      "delete all JWKs for message",
			messageID: messageID,
			wantErr:   false,
		},
		{
			name:      "delete nonexistent message (idempotent)",
			messageID: *testJWKGenService.GenerateUUIDv7(),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repo.DeleteByMessageID(ctx, tt.messageID)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify all JWKs deleted
			if tt.messageID == messageID {
				retrieved, err := repo.FindByMessageID(ctx, tt.messageID)
				require.NoError(t, err)
				require.Empty(t, retrieved, "Should have deleted all JWKs for message")
			}
		})
	}
}

func TestMessageRecipientJWKRepository_BarrierEncryption_RoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

	tests := []struct {
		name    string
		jwkData string
	}{
		{
			name:    "simple symmetric key",
			jwkData: generateTestJWK(t),
		},
		{
			name:    "RSA public key",
			jwkData: `{"kty":"RSA","n":"0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw","e":"AQAB"}`,
		},
		{
			name:    "EC key",
			jwkData: `{"kty":"EC","crv":"P-256","x":"f83OJ3D2xF1Bg8vub9tLe1gHMzV76e8Tus9uPHvRVEU","y":"x_FEzRu9m36HLN_tue659LNpXW6pCyStikYjKIWI5a0"}`,
		},
		{
			name:    "unicode characters in JWK",
			jwkData: `{"kty":"oct","k":"test-key-with-unicode-测试-🔐"}`,
		},
		{
			name:    "empty JWK",
			jwkData: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jwk := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
				ID:           *testJWKGenService.GenerateUUIDv7(),
				RecipientID:  *testJWKGenService.GenerateUUIDv7(),
				MessageID:    *testJWKGenService.GenerateUUIDv7(),
				EncryptedJWK: tt.jwkData,
			}

			// Create (encrypts with barrier)
			require.NoError(t, repo.Create(ctx, jwk))

			defer func() { require.NoError(t, repo.Delete(ctx, jwk.ID)) }()

			// Retrieve (decrypts with barrier)
			retrieved, err := repo.FindByRecipientAndMessage(ctx, jwk.RecipientID, jwk.MessageID)
			require.NoError(t, err)

			// Verify round-trip: original plaintext == decrypted plaintext
			require.Equal(t, tt.jwkData, retrieved.EncryptedJWK, "Barrier encryption/decryption should preserve original JWK data")
		})
	}
}

// generateTestJWK generates a test JWK using the test JWK generation service.
func generateTestJWK(t *testing.T) string {
	t.Helper()

	// Generate a symmetric key JWK for barrier encryption testing
	_, _, _, jwkJSON, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	return string(jwkJSON)
}
