// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"encoding/base64"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestJWEService_Encrypt_UnsupportedAlgorithm tests Encrypt with unsupported algorithm.
func TestJWEService_Encrypt_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWE")
}

// TestJWEService_Encrypt_CorruptedPublicJWKParse tests Encrypt with barrier-encrypted invalid JSON.
func TestJWEService_Encrypt_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

// TestJWEService_EncryptWithKID_UnsupportedAlgorithm tests EncryptWithKID with unsupported algorithm.
func TestJWEService_EncryptWithKID_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWE")
}

// TestJWEService_EncryptWithKID_CorruptedBase64 tests EncryptWithKID with invalid base64.
func TestJWEService_EncryptWithKID_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWEService_EncryptWithKID_CorruptedJWE tests EncryptWithKID with valid base64 invalid JWE.
func TestJWEService_EncryptWithKID_CorruptedJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-a-jwe"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

// TestJWEService_EncryptWithKID_CorruptedPublicJWKParse tests EncryptWithKID with corrupted public JWK JSON.
func TestJWEService_EncryptWithKID_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

// TestJWEService_Decrypt_CorruptedPrivateJWKParse tests Decrypt loop skip on corrupted private JWK.
func TestJWEService_Decrypt_CorruptedPrivateJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWEService_Encrypt_NoActiveMaterial tests Encrypt after retiring the only material.
func TestJWEService_Encrypt_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

// TestJWEService_Encrypt_BarrierDecryptFailure tests Encrypt when barrier decrypt fails.
func TestJWEService_Encrypt_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

// TestJWEService_Decrypt_BarrierDecryptFailure tests Decrypt loop skip when barrier decrypt fails.
func TestJWEService_Decrypt_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWEService_Decrypt_Base64DecodeFailure tests Decrypt loop skip when base64 decode fails.
func TestJWEService_Decrypt_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWSService_Sign_UnsupportedAlgorithm tests Sign with unsupported algorithm.
func TestJWSService_Sign_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWS")
}

// TestJWSService_Sign_NoActiveMaterial tests Sign after retiring the only material.
func TestJWSService_Sign_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

// TestJWSService_Sign_CorruptedPrivateJWKParse tests Sign with corrupted private JWK.
func TestJWSService_Sign_CorruptedPrivateJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse private JWK")
}

// TestJWSService_Sign_AlgorithmKeyMismatch tests Sign when algorithm is changed to mismatch key type.
func TestJWSService_Sign_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	// Create EC key (ES256) then tamper algorithm to RSA (RS256).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm).Error
	require.NoError(t, err)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create signer")
}

// TestJWSService_Verify_CorruptedPublicJWKParse tests Verify loop skip with corrupted public JWK.
func TestJWSService_Verify_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	signed, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, signed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWSService_Verify_BarrierDecryptFailure tests Verify loop skip when barrier decrypt fails.
func TestJWSService_Verify_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	signed, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, signed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWEService_Encrypt_AlgorithmKeyMismatch tests Encrypt when algorithm mismatches key type.
func TestJWEService_Encrypt_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	// Create RSA key then tamper algorithm to EC (ECDH-ES).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create encrypter")
}

// TestJWEService_EncryptWithKID_AlgorithmKeyMismatch tests EncryptWithKID with algorithm/key mismatch.
func TestJWEService_EncryptWithKID_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create encrypter")
}
