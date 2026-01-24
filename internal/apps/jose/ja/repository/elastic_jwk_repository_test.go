// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

func TestElasticJWKRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		jwk     func() *cryptoutilAppsJoseJaDomain.ElasticJWK
		wantErr bool
	}{
		{
			name: "valid elastic JWK creation",
			jwk: func() *cryptoutilAppsJoseJaDomain.ElasticJWK {
				id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
				tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

				return &cryptoutilAppsJoseJaDomain.ElasticJWK{
					ID:           *id,
					TenantID:     *tenantID,
					KID:          "test-kid-" + id.String()[:8],
					KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
					Algorithm:    "RS256",
					Use:          "sig",
					MaxMaterials: 10,
					CreatedAt:    time.Now(),
				}
			},
			wantErr: false,
		},
		{
			name: "elastic JWK with EC key type",
			jwk: func() *cryptoutilAppsJoseJaDomain.ElasticJWK {
				id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
				tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

				return &cryptoutilAppsJoseJaDomain.ElasticJWK{
					ID:           *id,
					TenantID:     *tenantID,
					KID:          "test-ec-" + id.String()[:8],
					KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeEC,
					Algorithm:    "ES256",
					Use:          "sig",
					MaxMaterials: 5,
					CreatedAt:    time.Now(),
				}
			},
			wantErr: false,
		},
		{
			name: "elastic JWK with OKP key type",
			jwk: func() *cryptoutilAppsJoseJaDomain.ElasticJWK {
				id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
				tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

				return &cryptoutilAppsJoseJaDomain.ElasticJWK{
					ID:           *id,
					TenantID:     *tenantID,
					KID:          "test-okp-" + id.String()[:8],
					KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeOKP,
					Algorithm:    "EdDSA",
					Use:          "sig",
					MaxMaterials: 20,
					CreatedAt:    time.Now(),
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testJWK := tt.jwk()
			repo := NewElasticJWKRepository(testDB)
			err := repo.Create(ctx, testJWK)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify retrieval works.
			retrieved, err := repo.Get(ctx, testJWK.TenantID, testJWK.KID)
			require.NoError(t, err)
			require.Equal(t, testJWK.ID, retrieved.ID)
			require.Equal(t, testJWK.TenantID, retrieved.TenantID)
			require.Equal(t, testJWK.KID, retrieved.KID)
			require.Equal(t, testJWK.KeyType, retrieved.KeyType)
			require.Equal(t, testJWK.MaxMaterials, retrieved.MaxMaterials)

			// Cleanup.
			require.NoError(t, repo.Delete(ctx, testJWK.ID))
		})
	}
}

func TestElasticJWKRepository_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create a test JWK first - use unique IDs for this test function.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-get-" + id.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, testJWK))

	t.Cleanup(func() {
		_ = repo.Delete(ctx, testJWK.ID)
	})

	// Run subtests sequentially to avoid race conditions with shared test data.
	t.Run("existing JWK", func(t *testing.T) {
		retrieved, err := repo.Get(ctx, testJWK.TenantID, testJWK.KID)
		require.NoError(t, err)
		require.Equal(t, testJWK.ID, retrieved.ID)
		require.Equal(t, testJWK.KID, retrieved.KID)
	})

	t.Run("non-existent KID", func(t *testing.T) {
		_, err := repo.Get(ctx, testJWK.TenantID, "non-existent-kid")
		require.Error(t, err)
	})

	t.Run("wrong tenant ID", func(t *testing.T) {
		_, err := repo.Get(ctx, googleUuid.New(), testJWK.KID)
		require.Error(t, err)
	})
}

func TestElasticJWKRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create multiple test JWKs - use unique tenant for this test to avoid conflicts.
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	var createdJWKs []*cryptoutilAppsJoseJaDomain.ElasticJWK

	for i := 0; i < 5; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
			ID:           *id,
			TenantID:     *tenantID,
			KID:          "test-list-" + id.String(), // Use full UUID to avoid collisions
			KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
			Algorithm:    "RS256",
			Use:          "sig",
			MaxMaterials: 10,
			CreatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, jwk))
		createdJWKs = append(createdJWKs, jwk)
	}

	// CRITICAL: Use t.Cleanup instead of defer to ensure cleanup happens AFTER parallel subtests complete.
	t.Cleanup(func() {
		for _, jwk := range createdJWKs {
			_ = repo.Delete(ctx, jwk.ID)
		}
	})

	tests := []struct {
		name      string
		tenantID  googleUuid.UUID
		offset    int
		limit     int
		wantCount int
		wantTotal int64
	}{
		{
			name:      "list all JWKs",
			tenantID:  *tenantID,
			offset:    0,
			limit:     100,
			wantCount: 5,
			wantTotal: 5,
		},
		{
			name:      "list with pagination - first page",
			tenantID:  *tenantID,
			offset:    0,
			limit:     2,
			wantCount: 2,
			wantTotal: 5,
		},
		{
			name:      "list with pagination - second page",
			tenantID:  *tenantID,
			offset:    2,
			limit:     2,
			wantCount: 2,
			wantTotal: 5,
		},
		{
			name:      "list with wrong tenant",
			tenantID:  googleUuid.New(),
			offset:    0,
			limit:     100,
			wantCount: 0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jwks, total, err := repo.List(ctx, tt.tenantID, tt.offset, tt.limit)
			require.NoError(t, err)
			require.Len(t, jwks, tt.wantCount)
			require.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestElasticJWKRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create a test JWK first.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-update-" + id.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, testJWK))

	defer func() {
		_ = repo.Delete(ctx, testJWK.ID)
	}()

	// Update the JWK.
	testJWK.MaxMaterials = 20

	err := repo.Update(ctx, testJWK)
	require.NoError(t, err)

	// Verify the update.
	retrieved, err := repo.Get(ctx, testJWK.TenantID, testJWK.KID)
	require.NoError(t, err)
	require.Equal(t, 20, retrieved.MaxMaterials)
}

func TestElasticJWKRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create a test JWK first.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-delete-" + id.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, testJWK))

	// Delete the JWK.
	err := repo.Delete(ctx, testJWK.ID)
	require.NoError(t, err)

	// Verify it's deleted.
	_, err = repo.Get(ctx, testJWK.TenantID, testJWK.KID)
	require.Error(t, err)
}

// TestElasticJWKRepository_GetByID tests retrieving an Elastic JWK by its UUID.
func TestElasticJWKRepository_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create a test JWK first.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-getbyid-" + id.String()[:8],
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
		CreatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, testJWK))

	defer func() {
		_ = repo.Delete(ctx, testJWK.ID)
	}()

	// Test GetByID with existing ID.
	retrieved, err := repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, testJWK.ID, retrieved.ID)
	require.Equal(t, testJWK.KID, retrieved.KID)
	require.Equal(t, testJWK.TenantID, retrieved.TenantID)

	// Test GetByID with non-existent ID.
	nonExistentID := googleUuid.New()
	_, err = repo.GetByID(ctx, nonExistentID)
	require.Error(t, err)
}

// TestElasticJWKRepository_IncrementMaterialCount tests atomic material count increment.
func TestElasticJWKRepository_IncrementMaterialCount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create a test JWK with initial count of 0.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   *id,
		TenantID:             *tenantID,
		KID:                  "test-increment-" + id.String()[:8],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            "RS256",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 0,
		CreatedAt:            time.Now(),
	}
	require.NoError(t, repo.Create(ctx, testJWK))

	defer func() {
		_ = repo.Delete(ctx, testJWK.ID)
	}()

	// Increment count.
	err := repo.IncrementMaterialCount(ctx, testJWK.ID)
	require.NoError(t, err)

	// Verify increment.
	retrieved, err := repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 1, retrieved.CurrentMaterialCount)

	// Increment again.
	err = repo.IncrementMaterialCount(ctx, testJWK.ID)
	require.NoError(t, err)

	retrieved, err = repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 2, retrieved.CurrentMaterialCount)

	// Increment multiple times to verify atomicity.
	for i := 0; i < 3; i++ {
		err = repo.IncrementMaterialCount(ctx, testJWK.ID)
		require.NoError(t, err)
	}

	retrieved, err = repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 5, retrieved.CurrentMaterialCount)
}

// TestElasticJWKRepository_DecrementMaterialCount tests atomic material count decrement.
func TestElasticJWKRepository_DecrementMaterialCount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewElasticJWKRepository(testDB)

	// Create a test JWK with initial count of 5.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   *id,
		TenantID:             *tenantID,
		KID:                  "test-decrement-" + id.String()[:8],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            "RS256",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 5,
		CreatedAt:            time.Now(),
	}
	require.NoError(t, repo.Create(ctx, testJWK))

	defer func() {
		_ = repo.Delete(ctx, testJWK.ID)
	}()

	// Decrement count.
	err := repo.DecrementMaterialCount(ctx, testJWK.ID)
	require.NoError(t, err)

	// Verify decrement.
	retrieved, err := repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 4, retrieved.CurrentMaterialCount)

	// Decrement multiple times.
	for i := 0; i < 3; i++ {
		err = repo.DecrementMaterialCount(ctx, testJWK.ID)
		require.NoError(t, err)
	}

	retrieved, err = repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 1, retrieved.CurrentMaterialCount)

	// Decrement to 0.
	err = repo.DecrementMaterialCount(ctx, testJWK.ID)
	require.NoError(t, err)

	retrieved, err = repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 0, retrieved.CurrentMaterialCount)

	// Attempt to decrement when already at 0 (should not go negative).
	err = repo.DecrementMaterialCount(ctx, testJWK.ID)
	require.NoError(t, err)

	retrieved, err = repo.GetByID(ctx, testJWK.ID)
	require.NoError(t, err)
	require.Equal(t, 0, retrieved.CurrentMaterialCount)
}
