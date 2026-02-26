// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestJWKSService_GetJWKSForElasticKey_CorruptedPublicJWK tests JWKS with corrupted public JWK JSON.
func TestJWKSService_GetJWKSForElasticKey_CorruptedPublicJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestJWKSService_GetPublicJWK_CorruptedPublicJWKParse tests GetPublicJWK with corrupted public JWK.
func TestJWKSService_GetPublicJWK_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

// TestJWKSService_GetJWKS_CorruptedPublicJWKParse tests GetJWKS skip on corrupted public JWK.
func TestJWKSService_GetJWKS_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestJWKSService_GetJWKSForElasticKey_InactiveMaterials tests JWKS returns empty when materials retired.
func TestJWKSService_GetJWKSForElasticKey_InactiveMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestJWKSService_GetPublicJWK_CorruptedBarrierDecrypt tests GetPublicJWK when barrier decrypt fails.
func TestJWKSService_GetPublicJWK_CorruptedBarrierDecrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

// TestGetJWKS_NoActiveMaterial tests GetJWKS with no active material (skip path).
func TestGetJWKS_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestGetJWKS_BarrierDecryptFailure tests GetJWKS skip when barrier decrypt fails.
func TestGetJWKS_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestGetJWKS_Base64DecodeFailure tests GetJWKS skip when base64 decode fails.
func TestGetJWKS_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestGetJWKSForElasticKey_Base64DecodeSkip tests GetJWKSForElasticKey skip on base64 error.
func TestGetJWKSForElasticKey_Base64DecodeSkip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestGetJWKSForElasticKey_BarrierDecryptSkip tests GetJWKSForElasticKey skip on barrier error.
func TestGetJWKSForElasticKey_BarrierDecryptSkip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestMaterialRotationService_RotateMaterial_UnsupportedAlgorithm tests rotation with unsupported algorithm.
func TestMaterialRotationService_RotateMaterial_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", "UNSUPPORTED-ALG").Error
	require.NoError(t, err)
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm")
}

// TestGetActiveMaterial_NoActiveMaterial tests GetActiveMaterial after retirement.
func TestGetActiveMaterial_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	_, err = rotationSvc.GetActiveMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

// TestJWKSService_GetPublicJWK_Base64DecodeFailure tests GetPublicJWK when base64 decode fails.
func TestJWKSService_GetPublicJWK_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWKSService_GetPublicJWK_ElasticJWKDeleted tests GetPublicJWK when elastic JWK is deleted.
func TestJWKSService_GetPublicJWK_ElasticJWKDeleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Delete elastic JWK record directly (bypass service which also deletes materials).
	err = testDB.Delete(&cryptoutilAppsJoseJaDomain.ElasticJWK{}, "id = ?", elasticJWK.ID).Error
	require.NoError(t, err)

	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get elastic JWK")
}

// TestAuditLogsByElasticJWK_TenantMismatch tests ListAuditLogsByElasticJWK with wrong tenant.
func TestAuditLogsByElasticJWK_TenantMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	auditSvc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	differentTenantID := googleUuid.New()

	_, _, err = auditSvc.ListAuditLogsByElasticJWK(ctx, differentTenantID, elasticJWK.ID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.Contains(t, err.Error(), "elastic JWK not found")
}

// TestDeleteElasticJWK_NotFound tests DeleteElasticJWK with non-existent ID.
func TestDeleteElasticJWK_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	err := elasticSvc.DeleteElasticJWK(ctx, tenantID, googleUuid.New())
	require.Error(t, err)
}

// closedDBMaterialRepo creates a material repository backed by a closed database.
// This triggers DB errors when the service tries to query materials, while the elastic
// repository (using the shared test DB) still works normally.
func closedDBMaterialRepo(t *testing.T) cryptoutilAppsJoseJaRepository.MaterialJWKRepository {
	t.Helper()

	tmpSQLDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	tmpGormDB, err := gorm.Open(sqlite.Dialector{Conn: tmpSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	require.NoError(t, tmpSQLDB.Close())

	return cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(tmpGormDB)
}

// TestRotateMaterial_CountMaterialsDBError covers material_rotation_service.go line 78 (CountMaterials DB error).
func TestRotateMaterial_CountMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create rotation service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	rotationSvc := NewMaterialRotationService(testElasticRepo, brokenMaterialRepo, testJWKGenService, testBarrierService)

	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to count materials")
}

// TestListMaterials_DBError covers material_rotation_service.go line 150 (ListByElasticJWK DB error).
func TestListMaterials_DBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create rotation service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	rotationSvc := NewMaterialRotationService(testElasticRepo, brokenMaterialRepo, testJWKGenService, testBarrierService)

	_, err = rotationSvc.ListMaterials(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list materials")
}

// TestGetJWKSForElasticKey_ListMaterialsDBError covers jwks_service.go line 119 (ListByElasticJWK DB error).
func TestGetJWKSForElasticKey_ListMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create JWKS service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	jwksSvc := NewJWKSService(testElasticRepo, brokenMaterialRepo, testBarrierService)

	_, err = jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list materials")
}

// TestVerify_ListMaterialsDBError covers jws_service.go line 109 (ListByElasticJWK DB error in Verify).
func TestVerify_ListMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB and sign something.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Sign something to get a valid JWS token.
	jws, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)

	// Create JWS service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	brokenJWSSvc := NewJWSService(testElasticRepo, brokenMaterialRepo, testBarrierService)

	_, err = brokenJWSSvc.Verify(ctx, tenantID, elasticJWK.ID, jws)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list materials")
}
