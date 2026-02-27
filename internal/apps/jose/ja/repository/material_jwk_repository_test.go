// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
		KID:                  "test-elastic-" + elasticJWKID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:                  cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials:         cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
		MaterialKID:    "material-" + id.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
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
		KID:                  "test-elastic-" + elasticJWKID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:                  cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials:         cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
	materialKID := "material-kid-" + id.String()[:cryptoutilSharedMagic.IMMinPasswordLength]
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
		KID:                  "test-elastic-" + elasticJWKID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:                  cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials:         cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
		MaterialKID:    "material-" + id.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
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
		KID:                  "test-elastic-" + elasticJWKID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:                  cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials:         cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
		MaterialKID:    "inactive-" + inactiveID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
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
		MaterialKID:    "active-" + activeID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
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
		KID:                  "test-elastic-" + elasticJWKID.String()[:cryptoutilSharedMagic.IMMinPasswordLength],
		KeyType:              cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:            cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:                  cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials:         cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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
	materials, total, err := repo.ListByElasticJWK(ctx, *elasticJWKID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
	materials, total, err = repo.ListByElasticJWK(ctx, nonExistentID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, materials)
}
