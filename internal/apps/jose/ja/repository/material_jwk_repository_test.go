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

func TestMaterialJWKRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)

	// Create parent ElasticJWK first.
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

	// Create MaterialJWK.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *id,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "material-" + id.String()[:8],
		PrivateJWKJWE:  "encrypted-private-jwk-content",
		PublicJWKJWE:   "encrypted-public-jwk-content",
		Active:         true,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}

	err := repo.Create(ctx, testMaterial)
	require.NoError(t, err)

	defer func() {
		_ = repo.Delete(ctx, testMaterial.ID)
	}()

	// Verify created.
	retrieved, err := repo.GetByID(ctx, testMaterial.ID)
	require.NoError(t, err)
	require.Equal(t, testMaterial.ID, retrieved.ID)
	require.Equal(t, testMaterial.MaterialKID, retrieved.MaterialKID)
	require.Equal(t, testMaterial.Active, retrieved.Active)
}

func TestMaterialJWKRepository_GetByMaterialKID(t *testing.T) {
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

	// Create MaterialJWK.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	materialKID := "material-kid-" + id.String()[:8]
	testMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *id,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    materialKID,
		PrivateJWKJWE:  "encrypted-private-content",
		PublicJWKJWE:   "encrypted-public-content",
		Active:         true,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, testMaterial))

	defer func() {
		_ = repo.Delete(ctx, testMaterial.ID)
	}()

	// Test successful retrieval by MaterialKID.
	retrieved, err := repo.GetByMaterialKID(ctx, materialKID)
	require.NoError(t, err)
	require.Equal(t, testMaterial.ID, retrieved.ID)
	require.Equal(t, materialKID, retrieved.MaterialKID)

	// Test error on non-existent MaterialKID.
	_, err = repo.GetByMaterialKID(ctx, "non-existent-material-kid")
	require.Error(t, err)
}

func TestMaterialJWKRepository_GetByID(t *testing.T) {
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

	// Create MaterialJWK.
	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	testMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *id,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "material-" + id.String()[:8],
		PrivateJWKJWE:  "encrypted-private-content",
		PublicJWKJWE:   "encrypted-public-content",
		Active:         false,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, testMaterial))

	defer func() {
		_ = repo.Delete(ctx, testMaterial.ID)
	}()

	// Test successful retrieval by ID.
	retrieved, err := repo.GetByID(ctx, *id)
	require.NoError(t, err)
	require.Equal(t, testMaterial.ID, retrieved.ID)
	require.Equal(t, testMaterial.MaterialKID, retrieved.MaterialKID)

	// Test error on non-existent ID.
	nonExistentID := googleUuid.New()
	_, err = repo.GetByID(ctx, nonExistentID)
	require.Error(t, err)
}

func TestMaterialJWKRepository_GetActiveMaterial(t *testing.T) {
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

	// Create inactive material.
	inactiveID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	inactiveMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *inactiveID,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "inactive-" + inactiveID.String()[:8],
		PrivateJWKJWE:  "encrypted-private",
		PublicJWKJWE:   "encrypted-public",
		Active:         false,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, inactiveMaterial))

	defer func() {
		_ = repo.Delete(ctx, inactiveMaterial.ID)
	}()

	// Create active material.
	activeID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	activeMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             *activeID,
		ElasticJWKID:   *elasticJWKID,
		MaterialKID:    "active-" + activeID.String()[:8],
		PrivateJWKJWE:  "encrypted-private-active",
		PublicJWKJWE:   "encrypted-public-active",
		Active:         true,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
	require.NoError(t, repo.Create(ctx, activeMaterial))

	defer func() {
		_ = repo.Delete(ctx, activeMaterial.ID)
	}()

	// Test retrieval of active material.
	retrieved, err := repo.GetActiveMaterial(ctx, *elasticJWKID)
	require.NoError(t, err)
	require.Equal(t, activeMaterial.ID, retrieved.ID)
	require.Equal(t, activeMaterial.MaterialKID, retrieved.MaterialKID)
	require.True(t, retrieved.Active)

	// Test error when no active material exists.
	nonExistentElasticID := googleUuid.New()
	_, err = repo.GetActiveMaterial(ctx, nonExistentElasticID)
	require.Error(t, err)
}

func TestMaterialJWKRepository_ListByElasticJWK(t *testing.T) {
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

	// Create 3 material JWKs.
	materialIDs := make([]googleUuid.UUID, 3)

	for i := 0; i < 3; i++ {
		id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		materialKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
		materialIDs[i] = *id
		material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
			ID:             *id,
			ElasticJWKID:   *elasticJWKID,
			MaterialKID:    "material-" + materialKID.String(),
			PrivateJWKJWE:  "encrypted-private",
			PublicJWKJWE:   "encrypted-public",
			Active:         i == 2, // Only last one active.
			CreatedAt:      time.Now().UTC().Add(time.Duration(i) * time.Second),
			BarrierVersion: 1,
		}
		require.NoError(t, repo.Create(ctx, material))

		defer func(materialID googleUuid.UUID) {
			_ = repo.Delete(ctx, materialID)
		}(*id)
	}

	// Test list all materials.
	materials, total, err := repo.ListByElasticJWK(ctx, *elasticJWKID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, materials, 3)

	// Test pagination - first page.
	materials, total, err = repo.ListByElasticJWK(ctx, *elasticJWKID, 0, 2)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, materials, 2)

	// Test pagination - second page.
	materials, total, err = repo.ListByElasticJWK(ctx, *elasticJWKID, 2, 2)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, materials, 1)

	// Test empty result for non-existent ElasticJWK.
	nonExistentID := googleUuid.New()
	materials, total, err = repo.ListByElasticJWK(ctx, nonExistentID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, materials)
}

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
