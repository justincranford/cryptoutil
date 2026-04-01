package service

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	googleUuid "github.com/google/uuid"
)

func TestGetJWKSForElasticKey_BarrierDecryptSkip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestGetJWKSForElasticKey_Base64DecodeSkip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestGetJWKSForElasticKey_ListMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create JWKS service with working elastic repo but closed-DB material repo.
	brokenMaterialRepo := closedDBMaterialRepo(t)
	jwksSvc := NewJWKSService(testElasticRepo, brokenMaterialRepo, testBarrierService)

	_, err = jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list materials")
}

func TestGetJWKS_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestGetJWKS_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestGetJWKS_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetJWKSDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.GetJWKS(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetJWKSForElasticKeyDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.GetJWKSForElasticKey(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetJWKSForElasticKey_CorruptedPublicJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetJWKSForElasticKey_InactiveMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKSForElasticKey(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetJWKS_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// GetJWKS skips corrupted materials and returns empty JWKS (graceful degradation).
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	// The corrupted material is skipped, resulting in empty keys.
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetJWKS_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetJWKS_EmptyForWrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create signing key for one tenant.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Get JWKS with different tenant - should return empty (not an error).
	jwks, err := jwksSvc.GetJWKS(ctx, otherTenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

func TestJWKSService_GetPublicJWKDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.GetPublicJWK(ctx, googleUuid.New(), "test-kid")
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetPublicJWK_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestJWKSService_GetPublicJWK_CorruptedBarrierDecrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

func TestJWKSService_GetPublicJWK_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to get public JWK - should fail on base64 decode.
	// GetPublicJWK signature: (ctx, tenantID, kid).
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestJWKSService_GetPublicJWK_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	_, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

func TestJWKSService_GetPublicJWK_ElasticJWKDeleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Delete elastic JWK record directly (bypass service which also deletes materials).
	err = testDB.Delete(&cryptoutilAppsJoseJaModel.ElasticJWK{}, "id = ?", elasticJWK.ID).Error
	require.NoError(t, err)

	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get elastic JWK")
}

func TestJWKSService_GetPublicJWK_WrongKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to get public JWK with non-existent KID.
	_, err := jwksSvc.GetPublicJWK(ctx, tenantID, "nonexistent-kid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

func TestVerify_ListMaterialsDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK in the working shared DB and sign something.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
