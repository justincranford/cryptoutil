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

func TestJWSService_SignDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.Sign(ctx, googleUuid.New(), googleUuid.New(), []byte("test payload"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWSService_SignWithKIDDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.SignWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test payload"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWSService_SignWithKID_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to sign using elasticJWK1 but with material2's KID.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "material key does not belong to elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

func TestJWSService_SignWithKID_WrongKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to sign with non-existent KID.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK.ID, "nonexistent-kid", []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

func TestJWSService_Sign_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	// Create EC key (ES256) then tamper algorithm to RSA (RS256).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm).Error
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create signer")
}

func TestJWSService_Sign_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE with invalid base64 directly in DB.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to sign - should fail on base64 decode.
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

func TestJWSService_Sign_CorruptedJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE with valid base64 but invalid JWE.
	// Use base64 encoding of "not a valid JWE" string.
	invalidJWE := "bm90IGEgdmFsaWQgSldF" // base64("not a valid JWE")
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", invalidJWE).Error
	require.NoError(t, err)

	// Try to sign - should fail on barrier decrypt.
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt private JWK")
}

func TestJWSService_Sign_CorruptedPrivateJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse private JWK")
}

func TestJWSService_Sign_CorruptedPrivateKeyInDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "invalid-base64!!!").Error
	require.NoError(t, err)

	// Try to sign - should fail due to corrupted stored key.
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

func TestJWSService_Sign_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

func TestJWSService_Sign_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWS")
}

func TestJWSService_VerifyDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.Verify(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.dGVzdA.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or verify.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWSService_Verify_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	signed, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, signed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWSService_Verify_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// First sign something valid.
	jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)
	require.NotEmpty(t, jwsCompact)

	// Now corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to verify - Verify catches decode errors and tries next material.
	// Since we only have one material and it fails, returns "no matching key found".
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, jwsCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWSService_Verify_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	signed, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, signed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}
