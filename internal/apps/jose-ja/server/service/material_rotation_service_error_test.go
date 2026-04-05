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

func TestGetActiveMaterial_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	_, err = rotationSvc.GetActiveMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

func TestListMaterials_DBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create rotation service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	rotationSvc := NewMaterialRotationService(testElasticRepo, brokenMaterialRepo, testJWKGenService, testBarrierService)

	_, err = rotationSvc.ListMaterials(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list materials")
}

func TestMaterialRotationService_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.GetActiveMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or get active material.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetMaterialByKIDDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.GetMaterialByKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid")
	require.Error(t, err)
	// Could fail on get elastic JWK or get material by KID.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_ListMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.ListMaterials(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or list materials.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err := svc.RetireMaterial(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK, get material by KID, or update.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RetireMaterial_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to retire material2 via elasticJWK1 - should fail.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK1.ID, material2.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "material not found for this elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

func TestMaterialRotationService_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.RotateMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or later operations.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RotateMaterial_CreatesMaterialSuccessfully(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with initial material.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, initialMaterial)
	require.True(t, initialMaterial.Active)

	// Rotate creates a new active material.
	newMaterial, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, newMaterial)
	require.True(t, newMaterial.Active)
	require.NotEqual(t, initialMaterial.ID, newMaterial.ID)
}

func TestMaterialRotationService_RotateMaterial_MaxMaterialsReached(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 2 materials.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, 2)
	require.NoError(t, err)

	// Rotate once - should succeed (now have 2 materials).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Rotate again - should fail (max reached).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max materials reached")
}

func TestMaterialRotationService_RotateMaterial_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", "UNSUPPORTED-ALG").Error
	require.NoError(t, err)
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm")
}

func TestRotateMaterial_CountMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create rotation service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	rotationSvc := NewMaterialRotationService(testElasticRepo, brokenMaterialRepo, testJWKGenService, testBarrierService)

	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to count materials")
}
