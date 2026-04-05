package service

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func TestJWEService_DecryptDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.Decrypt(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2R0NNIn0.test.test.test.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or decrypt.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWEService_Decrypt_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWEService_Decrypt_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWEService_Decrypt_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	// Use RSA/2048 key type which maps to RSA-OAEP-256 algorithm for encryption.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// First encrypt something valid.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Now corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to decrypt - Decrypt catches decode errors and tries next material.
	// Since we only have one material and it fails, returns "no matching key found".
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWEService_Decrypt_CorruptedJWEInDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Create valid JWE first.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Corrupt the material's PrivateJWKJWE (base64 decode will fail).
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "invalid-base64!!!").Error
	require.NoError(t, err)

	// Try to decrypt - should fail due to corrupted stored key.
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	// Should get "no matching key found" since decode fails and it continues to next material.
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWEService_Decrypt_CorruptedPrivateJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encrypted, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, encrypted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

func TestJWEService_EncryptDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.Encrypt(ctx, googleUuid.New(), googleUuid.New(), []byte("test plaintext"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWEService_EncryptDecrypt_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create symmetric key for direct encryption.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	plaintext := []byte("secret message for symmetric key")
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	decrypted, err := jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestJWEService_EncryptWithKIDDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.EncryptWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test plaintext"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWEService_EncryptWithKID_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create encrypter")
}

func TestJWEService_EncryptWithKID_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestJWEService_EncryptWithKID_CorruptedJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-a-jwe"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

func TestJWEService_EncryptWithKID_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

func TestJWEService_EncryptWithKID_MaterialNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to encrypt with non-existent material KID.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, "nonexistent-kid", []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

func TestJWEService_EncryptWithKID_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two encryption keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to encrypt using elasticJWK1 but with material2's KID.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "material key does not belong to elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

func TestJWEService_EncryptWithKID_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWE")
}

func TestJWEService_EncryptWithKID_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create a signing key (not encryption).
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

func TestJWEService_Encrypt_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	// Create RSA key then tamper algorithm to EC (ECDH-ES).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create encrypter")
}

func TestJWEService_Encrypt_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

func TestJWEService_Encrypt_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	// Use RSA/2048 key type which maps to RSA-OAEP-256 algorithm for encryption.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE with invalid base64.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to encrypt - should fail on base64 decode.
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestJWEService_Encrypt_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

func TestJWEService_Encrypt_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

func TestJWEService_Encrypt_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWE")
}

func TestJWEService_Encrypt_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create a signing key (not encryption).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}
