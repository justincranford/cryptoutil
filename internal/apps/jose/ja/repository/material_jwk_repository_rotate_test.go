// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMaterialJWKRepository_RotateMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	// Setup parent ElasticJWK.
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   *elasticJWKID,
		TenantID:             *tenantID,
		KID:                  "test-elastic-" + elasticJWKID.String()[:8],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            "RS256",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 0,
		CreatedAt:            time.Now().UTC(),
	}
	elasticRepo := NewElasticJWKRepository(testDB)
	require.NoError(t, elasticRepo.Create(ctx, elasticJWK))

	defer func() {
		_ = elasticRepo.Delete(ctx, elasticJWK.ID)
	}()

	// Create initial active material.
	oldID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	oldMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *oldID,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "old-material-" + oldID.String()[:8],
		PrivateJWKJWE:  "encrypted-old-private",
		PublicJWKJWE:   "encrypted-old-public",
		Active:         true,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, oldMaterial))

	defer func() {
		_ = repo.Delete(ctx, oldMaterial.ID)
	}()

	// Create new material for rotation.
	newID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	newMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *newID,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "new-material-" + newID.String()[:8],
		PrivateJWKJWE:  "encrypted-new-private",
		PublicJWKJWE:   "encrypted-new-public",
		Active:         false, // Will be set to true by RotateMaterial.
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}

	// Perform rotation (atomic transaction).
	err := repo.RotateMaterial(ctx, *elasticJWKID, newMaterial)
	require.NoError(t, err)

	defer func() {
		_ = repo.Delete(ctx, newMaterial.ID)
	}()

	// Verify old material is retired.
	retrievedOld, err := repo.GetByID(ctx, *oldID)
	require.NoError(t, err)
	require.False(t, retrievedOld.Active)
	require.NotNil(t, retrievedOld.RetiredAt)

	// Verify new material is active.
	retrievedNew, err := repo.GetByID(ctx, *newID)
	require.NoError(t, err)
	require.True(t, retrievedNew.Active)

	// Verify GetActiveMaterial returns new material.
	activeMaterial, err := repo.GetActiveMaterial(ctx, *elasticJWKID)
	require.NoError(t, err)
	require.Equal(t, newMaterial.ID, activeMaterial.ID)
}

func TestMaterialJWKRepository_RetireMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	// Setup parent ElasticJWK.
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   *elasticJWKID,
		TenantID:             *tenantID,
		KID:                  "test-elastic-" + elasticJWKID.String()[:8],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            "RS256",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 0,
		CreatedAt:            time.Now().UTC(),
	}
	elasticRepo := NewElasticJWKRepository(testDB)
	require.NoError(t, elasticRepo.Create(ctx, elasticJWK))

	defer func() {
		_ = elasticRepo.Delete(ctx, elasticJWK.ID)
	}()

	// Create active material.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *id,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "material-" + id.String()[:8],
		PrivateJWKJWE:  "encrypted-private",
		PublicJWKJWE:   "encrypted-public",
		Active:         true,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, material))

	defer func() {
		_ = repo.Delete(ctx, material.ID)
	}()

	// Retire material.
	err := repo.RetireMaterial(ctx, *id)
	require.NoError(t, err)

	// Verify material is retired.
	retrieved, err := repo.GetByID(ctx, *id)
	require.NoError(t, err)
	require.False(t, retrieved.Active)
	require.NotNil(t, retrieved.RetiredAt)
}

func TestMaterialJWKRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	// Setup parent ElasticJWK.
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   *elasticJWKID,
		TenantID:             *tenantID,
		KID:                  "test-elastic-" + elasticJWKID.String()[:8],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            "RS256",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 0,
		CreatedAt:            time.Now().UTC(),
	}
	elasticRepo := NewElasticJWKRepository(testDB)
	require.NoError(t, elasticRepo.Create(ctx, elasticJWK))

	defer func() {
		_ = elasticRepo.Delete(ctx, elasticJWK.ID)
	}()

	// Create material to delete.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *id,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "material-delete-" + id.String()[:8],
		PrivateJWKJWE:  "encrypted-private",
		PublicJWKJWE:   "encrypted-public",
		Active:         false,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, material))

	// Delete material.
	err := repo.Delete(ctx, *id)
	require.NoError(t, err)

	// Verify deletion.
	_, err = repo.GetByID(ctx, *id)
	require.Error(t, err)
}

func TestMaterialJWKRepository_CountMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	// Setup parent ElasticJWK.
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   *elasticJWKID,
		TenantID:             *tenantID,
		KID:                  "test-elastic-" + elasticJWKID.String()[:8],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            "RS256",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 0,
		CreatedAt:            time.Now().UTC(),
	}
	elasticRepo := NewElasticJWKRepository(testDB)
	require.NoError(t, elasticRepo.Create(ctx, elasticJWK))

	defer func() {
		_ = elasticRepo.Delete(ctx, elasticJWK.ID)
	}()

	// Test count with no materials.
	count, err := repo.CountMaterials(ctx, *elasticJWKID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Create 3 materials.
	materialIDs := make([]googleUuid.UUID, 3)

	for i := 0; i < 3; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		materialKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		materialIDs[i] = *id
		material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
			ID:             *id,
			ElasticJWKID:   *elasticJWKID,
			MaterialKID:    "material-count-" + materialKID.String(),
			PrivateJWKJWE:  "encrypted-private",
			PublicJWKJWE:   "encrypted-public",
			Active:         i == 2,
			CreatedAt:      time.Now().UTC().Add(time.Duration(i) * time.Second),
			BarrierVersion: 1,
		}
		require.NoError(t, repo.Create(ctx, material))

		defer func(materialID googleUuid.UUID) {
			_ = repo.Delete(ctx, materialID)
		}(*id)
	}

	// Test count with 3 materials.
	count, err = repo.CountMaterials(ctx, *elasticJWKID)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)

	// Test count for non-existent ElasticJWK.
	nonExistentID := googleUuid.New()
	count, err = repo.CountMaterials(ctx, nonExistentID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}
