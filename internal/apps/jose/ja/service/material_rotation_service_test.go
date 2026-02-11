// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestMaterialRotationService_RotateMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 5 materials.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Rotate material.
	newMaterial, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, newMaterial)
	require.NotEqual(t, initialMaterial.MaterialKID, newMaterial.MaterialKID)
	require.True(t, newMaterial.Active)
}

func TestMaterialRotationService_RotateMaterial_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to rotate with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = rotationSvc.RotateMaterial(ctx, wrongTenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMaterialRotationService_RotateMaterial_MaxReached(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 2 materials.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 2)
	require.NoError(t, err)

	// Rotate once - should succeed (now 2 materials).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Rotate again - should fail (max 2 reached).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max materials reached")
}

func TestMaterialRotationService_RetireMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 3 materials.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 3)
	require.NoError(t, err)

	// Rotate to create second material.
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Retire the initial material.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, initialMaterial.ID)
	require.NoError(t, err)
}

func TestMaterialRotationService_RetireMaterial_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 3)
	require.NoError(t, err)

	// Try to retire with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	err = rotationSvc.RetireMaterial(ctx, wrongTenantID, elasticJWK.ID, initialMaterial.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMaterialRotationService_ListMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 5 materials.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Rotate twice.
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// List materials - should have 3.
	materials, err := rotationSvc.ListMaterials(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, materials, 3)
}

func TestMaterialRotationService_ListMaterials_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to list with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = rotationSvc.ListMaterials(ctx, wrongTenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMaterialRotationService_GetActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Get active material - should return initial.
	activeMaterial, err := rotationSvc.GetActiveMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, activeMaterial)
	require.True(t, activeMaterial.Active)

	// Rotate.
	newMaterial, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Get active material - should return new.
	activeMaterial, err = rotationSvc.GetActiveMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, newMaterial.MaterialKID, activeMaterial.MaterialKID)
}

func TestMaterialRotationService_GetActiveMaterial_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to get with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = rotationSvc.GetActiveMaterial(ctx, wrongTenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMaterialRotationService_GetMaterialByKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Get material by KID.
	material, err := rotationSvc.GetMaterialByKID(ctx, tenantID, elasticJWK.ID, initialMaterial.MaterialKID)
	require.NoError(t, err)
	require.Equal(t, initialMaterial.ID, material.ID)
}

func TestMaterialRotationService_GetMaterialByKID_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to get with wrong tenant - should fail.
	wrongTenantID := googleUuid.New()
	_, err = rotationSvc.GetMaterialByKID(ctx, wrongTenantID, elasticJWK.ID, initialMaterial.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMaterialRotationService_GetMaterialByKID_InvalidKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to get with invalid KID - should fail.
	_, err = rotationSvc.GetMaterialByKID(ctx, tenantID, elasticJWK.ID, "invalid-kid")
	require.Error(t, err)
}

func TestMaterialRotationService_RotateMaterial_NonExistentKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to rotate non-existent elastic JWK - should fail.
	_, err := rotationSvc.RotateMaterial(ctx, tenantID, googleUuid.New())
	require.Error(t, err)
}

func TestMaterialRotationService_RetireMaterial_WrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 3)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 3)
	require.NoError(t, err)

	// Try to retire material2 using elasticJWK1 - should fail.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK1.ID, material2.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "material not found for this elastic JWK")

	// Verify the correct elastic JWK works.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK2.ID, material2.ID)
	require.NoError(t, err)
}

func TestMaterialRotationService_GetMaterialByKID_WrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	_, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to get material2 using elasticJWK1 - should fail.
	_, err = rotationSvc.GetMaterialByKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "material not found for this elastic JWK")
}

func TestMaterialRotationService_RotateMaterial_AllAlgorithms(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		algorithm string
		keyUse    string
	}{
		{"RS384", cryptoutilSharedMagic.JoseAlgRS384, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"RS512", cryptoutilSharedMagic.JoseAlgRS512, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"PS256", cryptoutilSharedMagic.JoseAlgPS256, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"PS384", cryptoutilSharedMagic.JoseAlgPS384, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"PS512", cryptoutilSharedMagic.JoseAlgPS512, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"ES384", cryptoutilSharedMagic.JoseAlgES384, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"ES512", cryptoutilSharedMagic.JoseAlgES512, cryptoutilAppsJoseJaDomain.KeyUseSig},
		{"RSA3072Enc", cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilAppsJoseJaDomain.KeyUseEnc},
		{"RSA4096Enc", cryptoutilSharedMagic.JoseKeyTypeRSA4096, cryptoutilAppsJoseJaDomain.KeyUseEnc},
		{"ECP384Enc", cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilAppsJoseJaDomain.KeyUseEnc},
		{"ECP521Enc", cryptoutilSharedMagic.JoseKeyTypeECP521, cryptoutilAppsJoseJaDomain.KeyUseEnc},
		{"Oct192Enc", cryptoutilSharedMagic.JoseKeyTypeOct192, cryptoutilAppsJoseJaDomain.KeyUseEnc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
			tenantID := googleUuid.New()

			elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, tt.algorithm, tt.keyUse, 5)
			require.NoError(t, err)

			newMaterial, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
			require.NoError(t, err)
			require.NotNil(t, newMaterial)
			require.True(t, newMaterial.Active)
		})
	}
}

func TestMaterialRotationService_RetireMaterial_NonExistentMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 5)
	require.NoError(t, err)

	// Try to retire non-existent material - should fail.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, googleUuid.New())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

func TestMaterialRotationService_RetireMaterial_NonExistentElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to retire material on non-existent elastic JWK - should fail.
	err := rotationSvc.RetireMaterial(ctx, tenantID, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
}

func TestMaterialRotationService_ListMaterials_NonExistentElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to list materials on non-existent elastic JWK - should fail.
	_, err := rotationSvc.ListMaterials(ctx, tenantID, googleUuid.New())
	require.Error(t, err)
}

func TestMaterialRotationService_GetActiveMaterial_NonExistentElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to get active material on non-existent elastic JWK - should fail.
	_, err := rotationSvc.GetActiveMaterial(ctx, tenantID, googleUuid.New())
	require.Error(t, err)
}

func TestMaterialRotationService_GetMaterialByKID_NonExistentElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to get material by KID on non-existent elastic JWK - should fail.
	_, err := rotationSvc.GetMaterialByKID(ctx, tenantID, googleUuid.New(), "some-kid")
	require.Error(t, err)
}
