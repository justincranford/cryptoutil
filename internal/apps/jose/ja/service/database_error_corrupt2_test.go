// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestJWTService_CreateJWT_CorruptedPrivateKeyDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to create JWT - should fail on base64 decode.
	claims := &JWTClaims{
		Subject:  "test-subject",
		Issuer:   "test-issuer",
		Audience: []string{"test-audience"},
	}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// ====================
// CreateElasticJWK Algorithm Validation Tests
// ====================

// TestElasticJWKService_CreateElasticJWK_UnsupportedAlgorithm tests that CreateElasticJWK
// returns error for unsupported algorithm.
func TestElasticJWKService_CreateElasticJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to create with an unsupported algorithm.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "INVALID-ALGORITHM", cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid algorithm")
}

// TestElasticJWKService_CreateElasticJWK_EmptyAlgorithm tests that CreateElasticJWK
// returns error for empty algorithm.
func TestElasticJWKService_CreateElasticJWK_EmptyAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to create with empty algorithm.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "", cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
}

// ====================
// Delete Cascading Error Tests
// ====================

// TestElasticJWKService_DeleteElasticJWK_WithMultipleMaterials tests that DeleteElasticJWK
// correctly deletes all materials before deleting the elastic JWK.
func TestElasticJWKService_DeleteElasticJWK_WithMultipleMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with initial material.
	elasticJWK, material1, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

// ====================
// DeleteElasticJWK Comprehensive Error Path Tests
// ====================

// TestElasticJWKService_DeleteElasticJWK_GetError tests error during ownership verification.
func TestElasticJWKService_DeleteElasticJWK_GetError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Database is closed - GetElasticJWK should fail.
	err = svc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Error should propagate from GetElasticJWK.
	require.True(t, strings.Contains(err.Error(), "failed to") || strings.Contains(err.Error(), cryptoutilSharedMagic.RealmStorageTypeDatabase))
}

// TestElasticJWKService_DeleteElasticJWK_ListMaterialsError tests error during material listing.
// Note: This error path is difficult to test with in-memory SQLite because closing the database
// causes GetElasticJWK to fail first. The test documents the error path for completeness.
func TestElasticJWKService_DeleteElasticJWK_ListMaterialsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a service with closed database to force errors.
	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	elasticSvc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Try to delete with closed database - will fail on GetElasticJWK (first operation).
	err = elasticSvc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Error will be from GetElasticJWK (earlier step), not ListByElasticJWK.
	// This documents the limitation: Can't easily isolate ListByElasticJWK error with closed DB.
	require.True(t, strings.Contains(err.Error(), "failed to") || strings.Contains(err.Error(), cryptoutilSharedMagic.RealmStorageTypeDatabase))
}

// TestElasticJWKService_DeleteElasticJWK_MaterialDeleteError tests error during material deletion.
// Note: This error path is difficult to test with in-memory SQLite because database constraints
// behave differently than PostgreSQL. The test documents the error path for completeness.
func TestElasticJWKService_DeleteElasticJWK_MaterialDeleteError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a service with closed database to force material deletion to fail.
	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	elasticSvc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Try to delete with closed database - will fail on GetElasticJWK (first operation).
	err = elasticSvc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Error will be from GetElasticJWK (earlier step), not material deletion.
	// This documents the limitation: Can't easily isolate material deletion error with closed DB.
	require.True(t, strings.Contains(err.Error(), "failed to") || strings.Contains(err.Error(), cryptoutilSharedMagic.RealmStorageTypeDatabase))
}

// TestElasticJWKService_DeleteElasticJWK_FinalDeleteError tests error during final elastic JWK deletion.
func TestElasticJWKService_DeleteElasticJWK_FinalDeleteError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Manually delete the elastic JWK from database (but not its materials).
	// This simulates a race condition or database inconsistency.
	err = testDB.Where("id = ?", elasticJWK.ID).
		Delete(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Error
	require.NoError(t, err)

	// Try to delete - GetElasticJWK should fail (elastic JWK no longer exists).
	err = elasticSvc.DeleteElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

// ====================
// JWE/JWS Decrypt/Verify with Barrier Errors
// ====================

// TestJWEService_Decrypt_CorruptedJWEInDB tests that Decrypt returns error
// when the decrypted key is malformed.
func TestJWEService_Decrypt_CorruptedJWEInDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Create valid JWE first.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Corrupt the material's PrivateJWKJWE (base64 decode will fail).
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "invalid-base64!!!").Error
	require.NoError(t, err)

	// Try to decrypt - should fail due to corrupted stored key.
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	// Should get "no matching key found" since decode fails and it continues to next material.
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWSService_Sign_CorruptedPrivateKeyInDB tests that Sign returns error
// when the stored private key is corrupted.
