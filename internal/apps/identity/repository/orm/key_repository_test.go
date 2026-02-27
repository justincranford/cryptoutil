// Copyright (c) 2025 Justin Cranford

package orm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestKeyRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewKeyRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name    string
		key     *cryptoutilIdentityDomain.Key
		wantErr bool
	}{
		{
			name: "create_signing_key",
			key: &cryptoutilIdentityDomain.Key{
				ID:         googleUuid.Must(googleUuid.NewV7()),
				Usage:      cryptoutilSharedMagic.KeyUsageSigning,
				Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
				PrivateKey: "test-private-key-data",
				Active:     true,
			},
			wantErr: false,
		},
		{
			name: "create_encryption_key",
			key: &cryptoutilIdentityDomain.Key{
				ID:         googleUuid.Must(googleUuid.NewV7()),
				Usage:      cryptoutilSharedMagic.KeyUsageEncryption,
				Algorithm:  cryptoutilSharedMagic.JoseAlgRSAOAEP256,
				PrivateKey: "test-encryption-key",
				Active:     true,
			},
			wantErr: false,
		},
		{
			name:    "nil_key",
			key:     nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(ctx, tc.key)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, tc.key.ID)
			}
		})
	}
}

func TestKeyRepository_FindByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewKeyRepository(testDB.db)
	ctx := context.Background()

	// Create a test key.
	key := &cryptoutilIdentityDomain.Key{
		ID:         googleUuid.Must(googleUuid.NewV7()),
		Usage:      cryptoutilSharedMagic.KeyUsageSigning,
		Algorithm:  cryptoutilSharedMagic.JoseAlgES256,
		PrivateKey: "test-ec-key",
		Active:     true,
	}
	err := repo.Create(ctx, key)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr bool
	}{
		{
			name:    "found",
			id:      key.ID,
			wantErr: false,
		},
		{
			name:    "not_found",
			id:      googleUuid.Must(googleUuid.NewV7()),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.FindByID(ctx, tc.id)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, key.ID, result.ID)
				require.Equal(t, key.Usage, result.Usage)
				require.Equal(t, key.Algorithm, result.Algorithm)
				require.True(t, result.Active)
			}
		})
	}
}

// TODO: Add TestKeyRepository_FindByUsage tests - requires fixing active field query logic.

func TestKeyRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewKeyRepository(testDB.db)
	ctx := context.Background()

	// Create a test key.
	key := &cryptoutilIdentityDomain.Key{
		ID:         googleUuid.Must(googleUuid.NewV7()),
		Usage:      cryptoutilSharedMagic.KeyUsageSigning,
		Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		PrivateKey: "original-key-data",
		Active:     true,
	}
	err := repo.Create(ctx, key)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     *cryptoutilIdentityDomain.Key
		wantErr bool
	}{
		{
			name: "update_active_status",
			key: &cryptoutilIdentityDomain.Key{
				ID:         key.ID,
				Usage:      key.Usage,
				Algorithm:  key.Algorithm,
				PrivateKey: key.PrivateKey,
				Active:     false, // Deactivate key
			},
			wantErr: false,
		},
		{
			name:    "nil_key",
			key:     nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(ctx, tc.key)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify update persisted.
				updated, err := repo.FindByID(ctx, tc.key.ID)
				require.NoError(t, err)
				require.Equal(t, tc.key.Active, updated.Active)
			}
		})
	}
}

func TestKeyRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewKeyRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() googleUuid.UUID
		wantErr bool
	}{
		{
			name: "delete_existing_key",
			setup: func() googleUuid.UUID {
				key := &cryptoutilIdentityDomain.Key{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageSigning,
					Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
					PrivateKey: "to-be-deleted",
					Active:     true,
				}
				err := repo.Create(ctx, key)
				require.NoError(t, err)

				return key.ID
			},
			wantErr: false,
		},
		{
			name: "delete_non_existent_key",
			setup: func() googleUuid.UUID {
				return googleUuid.Must(googleUuid.NewV7())
			},
			wantErr: false, // GORM doesn't error on delete of non-existent record
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			id := tc.setup()

			err := repo.Delete(ctx, id)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify key was deleted (soft delete means record exists but DeletedAt is set).
				_, err := repo.FindByID(ctx, id)
				require.Error(t, err, "Expected key to be soft deleted")
			}
		})
	}
}

func TestKeyRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewKeyRepository(testDB.db)
	ctx := context.Background()

	// Create 5 test keys.
	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		key := &cryptoutilIdentityDomain.Key{
			ID:         googleUuid.Must(googleUuid.NewV7()),
			Usage:      cryptoutilSharedMagic.KeyUsageSigning,
			Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			PrivateKey: "test-key",
			Active:     true,
		}
		require.NoError(t, repo.Create(ctx, key))
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "all_keys",
			limit:         0,
			offset:        0,
			expectedCount: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
		},
		{
			name:          "limit_2",
			limit:         2,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "offset_2",
			limit:         0,
			offset:        2,
			expectedCount: 3,
		},
		{
			name:          "limit_2_offset_1",
			limit:         2,
			offset:        1,
			expectedCount: 2,
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			keys, err := repo.List(ctx, tc.limit, tc.offset)
			require.NoError(t, err)
			require.NotNil(t, keys)
			require.Len(t, keys, tc.expectedCount)

			// Verify ordering (most recent first).
			if len(keys) > 1 {
				for i := 0; i < len(keys)-1; i++ {
					require.True(t, keys[i].CreatedAt.After(keys[i+1].CreatedAt) ||
						keys[i].CreatedAt.Equal(keys[i+1].CreatedAt),
						"Expected keys ordered by created_at DESC")
				}
			}
		})
	}
}

func TestKeyRepository_FindByUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupKeys     []*cryptoutilIdentityDomain.Key
		usage         string
		active        bool
		expectedCount int
	}{
		{
			name: "find_active_signing_keys_filter",
			setupKeys: []*cryptoutilIdentityDomain.Key{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageSigning,
					Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
					PrivateKey: "signing-key-1",
					Active:     true,
				},
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageEncryption,
					Algorithm:  cryptoutilSharedMagic.JoseAlgRSAOAEP256,
					PrivateKey: "encryption-key-1",
					Active:     true,
				},
			},
			usage:         cryptoutilSharedMagic.KeyUsageSigning,
			active:        true,
			expectedCount: 1,
		},
		{
			name: "find_all_signing_keys_no_active_filter",
			setupKeys: []*cryptoutilIdentityDomain.Key{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageSigning,
					Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
					PrivateKey: "signing-key-3",
					Active:     true,
				},
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageSigning,
					Algorithm:  cryptoutilSharedMagic.JoseAlgES256,
					PrivateKey: "signing-key-4",
					Active:     true,
				},
			},
			usage:         cryptoutilSharedMagic.KeyUsageSigning,
			active:        false,
			expectedCount: 2,
		},
		{
			name: "find_all_encryption_keys",
			setupKeys: []*cryptoutilIdentityDomain.Key{
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageEncryption,
					Algorithm:  cryptoutilSharedMagic.JoseAlgRSAOAEP256,
					PrivateKey: "encryption-key-1",
					Active:     true,
				},
				{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageEncryption,
					Algorithm:  cryptoutilSharedMagic.JoseAlgRSAOAEP256,
					PrivateKey: "encryption-key-2",
					Active:     false,
				},
			},
			usage:         cryptoutilSharedMagic.KeyUsageEncryption,
			active:        false,
			expectedCount: 2,
		},
		{
			name:          "no_matching_keys",
			setupKeys:     []*cryptoutilIdentityDomain.Key{},
			usage:         "nonexistent",
			active:        true,
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testDB := setupTestDB(t)
			repo := NewKeyRepository(testDB.db)
			ctx := context.Background()

			for _, key := range tc.setupKeys {
				require.NoError(t, repo.Create(ctx, key))

				// Verify key was created with correct Active value.
				created, err := repo.FindByID(ctx, key.ID)
				require.NoError(t, err)
				require.Equal(t, key.Active, created.Active, "Active field mismatch after Create")
			}

			keys, err := repo.FindByUsage(ctx, tc.usage, tc.active)
			require.NoError(t, err)

			if len(keys) != tc.expectedCount {
				t.Logf("Expected %d keys, got %d. Keys returned:", tc.expectedCount, len(keys))

				for i, k := range keys {
					t.Logf("  Key[%d]: usage=%s, active=%v, algorithm=%s", i, k.Usage, k.Active, k.Algorithm)
				}
			}

			require.Len(t, keys, tc.expectedCount)

			// Verify ordering (most recent first).
			if len(keys) > 1 {
				for i := 0; i < len(keys)-1; i++ {
					require.True(t, keys[i].CreatedAt.After(keys[i+1].CreatedAt) ||
						keys[i].CreatedAt.Equal(keys[i+1].CreatedAt),
						"Expected keys ordered by created_at DESC")
				}
			}
		})
	}
}

func TestKeyRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewKeyRepository(testDB.db)
	ctx := context.Background()

	tests := []struct {
		name          string
		setupCount    int
		expectedCount int64
	}{
		{
			name:          "empty_database",
			setupCount:    0,
			expectedCount: 0,
		},
		{
			name:          "three_keys",
			setupCount:    3,
			expectedCount: 3,
		},
	}

	for _, tc := range tests {
		// Capture loop variable for parallel execution.
		t.Run(tc.name, func(t *testing.T) {
			// Create test keys for this subtest.
			for i := 0; i < tc.setupCount; i++ {
				key := &cryptoutilIdentityDomain.Key{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					Usage:      cryptoutilSharedMagic.KeyUsageSigning,
					Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
					PrivateKey: "test-key",
					Active:     true,
				}
				require.NoError(t, repo.Create(ctx, key))
			}

			count, err := repo.Count(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.expectedCount, count)
		})
	}
}
