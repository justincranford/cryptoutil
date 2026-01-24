// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"testing"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Tests in this file use the shared testElasticJWKSvc from elastic_jwk_service_test.go TestMain.

func TestRotateMaterial_Success(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.Equal(t, 1, createResp.ElasticJWK.CurrentMaterialCount)

	// Rotate material.
	rotationResp, err := testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, rotationResp)
	require.NotNil(t, rotationResp.MaterialJWK)
	require.Equal(t, createResp.ElasticJWK.ID, rotationResp.MaterialJWK.ElasticJWKID)

	// Verify material count updated.
	updatedElastic, err := testElasticRepo.Get(testCtx, tenantID, realmID, createResp.ElasticJWK.KID)
	require.NoError(t, err)
	require.Equal(t, 2, updatedElastic.CurrentMaterialCount)
}

func TestRotateMaterial_TenantMismatch(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create an elastic JWK for tenantID.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Try to rotate with different tenant.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, otherTenantID, realmID, createResp.ElasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tenant/realm")
}

func TestRotateMaterial_RealmMismatch(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	otherRealmID := googleUuid.New()

	// Create an elastic JWK for realmID.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Try to rotate with different realm.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, otherRealmID, createResp.ElasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tenant/realm")
}

func TestRotateMaterial_NotFound(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	nonExistentID := googleUuid.New()

	// Try to rotate non-existent elastic JWK.
	_, err := testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, nonExistentID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestRotateMaterial_AtLimit(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Manually set material count to limit via direct query.
	err = testDB.Exec("UPDATE elastic_jwks SET current_material_count = ? WHERE id = ?",
		MaxMaterialsPerElasticJWK, createResp.ElasticJWK.ID).Error
	require.NoError(t, err)

	// Also need to add fake material rows for count check.
	for i := 1; i < MaxMaterialsPerElasticJWK; i++ {
		material := &cryptoutilJoseDomain.MaterialJWK{
			ID:             googleUuid.New(),
			ElasticJWKID:   createResp.ElasticJWK.ID,
			MaterialKID:    googleUuid.New().String(),
			PrivateJWKJWE:  "fake-encrypted-private",
			PublicJWKJWE:   "fake-encrypted-public",
			Active:         false,
			BarrierVersion: 1,
		}
		err = testMaterialRepo.Create(testCtx, material)
		require.NoError(t, err)
	}

	// Try to rotate - should fail.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max")
}

func TestCanRotate(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Should be able to rotate.
	canRotate, count, err := testElasticJWKSvc.CanRotate(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.True(t, canRotate)
	require.Equal(t, int64(1), count)
}

func TestGetMaterialCount(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Check material count.
	count, err := testElasticJWKSvc.GetMaterialCount(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Rotate once.
	_, err = testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	// Check material count again.
	count, err = testElasticJWKSvc.GetMaterialCount(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestRotateMaterial_MultipleRotations(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "EC",
		ALG:      "ES256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)
	require.Equal(t, 1, createResp.ElasticJWK.CurrentMaterialCount)

	// Rotate 5 times.
	materialKIDs := []string{createResp.MaterialJWK.MaterialKID} // Initial material.

	for i := 0; i < 5; i++ {
		rotationResp, err := testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
		require.NoError(t, err)

		materialKIDs = append(materialKIDs, rotationResp.MaterialJWK.MaterialKID)
	}

	// Verify all material KIDs are unique.
	require.Len(t, materialKIDs, 6)

	kidSet := make(map[string]bool)
	for _, kid := range materialKIDs {
		require.False(t, kidSet[kid], "duplicate material KID: %s", kid)
		kidSet[kid] = true
	}

	// Verify final material count.
	count, err := testElasticJWKSvc.GetMaterialCount(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(6), count)

	// List all materials.
	materials, err := testMaterialRepo.ListByElasticJWK(testCtx, createResp.ElasticJWK.ID, 0, 100)
	require.NoError(t, err)
	require.Len(t, materials, 6)
}

func TestRotateMaterial_SymmetricKey(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an elastic JWK with HMAC algorithm.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "oct",
		ALG:      "HS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)
	require.Equal(t, "oct", createResp.ElasticJWK.KTY)

	// Rotate material.
	rotationResp, err := testElasticJWKSvc.RotateMaterial(testCtx, tenantID, realmID, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, rotationResp)

	// Verify material count.
	count, err := testElasticJWKSvc.GetMaterialCount(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}
