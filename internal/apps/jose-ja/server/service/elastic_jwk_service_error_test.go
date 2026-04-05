package service

import (
	"context"
	"strings"
	"testing"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func TestDeleteElasticJWK_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	err := elasticSvc.DeleteElasticJWK(ctx, tenantID, googleUuid.New())
	require.Error(t, err)
}

func TestElasticJWKService_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Valid parameters, but database is closed.
	_, _, err := svc.CreateElasticJWK(ctx, googleUuid.New(), cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create elastic JWK"))
}

func TestElasticJWKService_CreateElasticJWK_EmptyAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to create with empty algorithm.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "", cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
}

func TestElasticJWKService_CreateElasticJWK_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create symmetric key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Verify the key was created.
	retrieved, err := elasticSvc.GetElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, elasticJWK.ID, retrieved.ID)
}

func TestElasticJWKService_CreateElasticJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to create with an unsupported algorithm.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "INVALID-ALGORITHM", cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid algorithm")
}

func TestElasticJWKService_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err := svc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on any database operation.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestElasticJWKService_DeleteElasticJWK_FinalDeleteError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Manually delete the elastic JWK from database (but not its materials).
	// This simulates a race condition or database inconsistency.
	err = testDB.Where("id = ?", elasticJWK.ID).
		Delete(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Error
	require.NoError(t, err)

	// Try to delete - GetElasticJWK should fail (elastic JWK no longer exists).
	err = elasticSvc.DeleteElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

func TestElasticJWKService_DeleteElasticJWK_GetError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Database is closed - GetElasticJWK should fail.
	err := svc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Error should propagate from GetElasticJWK.
	require.True(t, strings.Contains(err.Error(), "failed to") || strings.Contains(err.Error(), cryptoutilSharedMagic.RealmStorageTypeDatabase))
}

func TestElasticJWKService_DeleteElasticJWK_ListMaterialsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a service with closed database to force errors.
	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	elasticSvc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Try to delete with closed database - will fail on GetElasticJWK (first operation).
	err := elasticSvc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Error will be from GetElasticJWK (earlier step), not ListByElasticJWK.
	// This documents the limitation: Can't easily isolate ListByElasticJWK error with closed DB.
	require.True(t, strings.Contains(err.Error(), "failed to") || strings.Contains(err.Error(), cryptoutilSharedMagic.RealmStorageTypeDatabase))
}

func TestElasticJWKService_DeleteElasticJWK_MaterialDeleteError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a service with closed database to force material deletion to fail.
	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	elasticSvc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Try to delete with closed database - will fail on GetElasticJWK (first operation).
	err := elasticSvc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Error will be from GetElasticJWK (earlier step), not material deletion.
	// This documents the limitation: Can't easily isolate material deletion error with closed DB.
	require.True(t, strings.Contains(err.Error(), "failed to") || strings.Contains(err.Error(), cryptoutilSharedMagic.RealmStorageTypeDatabase))
}

func TestElasticJWKService_DeleteElasticJWK_WithMultipleMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with initial material.
	elasticJWK, material1, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material1)

	// Rotate to create second material.
	material2, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, material2)

	// Rotate to create third material.
	material3, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, material3)

	// Verify we have 3 materials.
	materials, err := rotationSvc.ListMaterials(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, materials, 3)

	// Delete elastic JWK should cascade delete all materials.
	err = elasticSvc.DeleteElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Verify elastic JWK is deleted.
	_, err = elasticSvc.GetElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestElasticJWKService_GetDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.GetElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK"))
}

func TestElasticJWKService_ListDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, _, err := svc.ListElasticJWKs(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list elastic JWKs"))
}
