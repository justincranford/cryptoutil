// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"encoding/base64"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestJWTService_CreateJWT_CorruptedPrivateJWKParse tests CreateJWT with corrupted private JWK.
func TestJWTService_CreateJWT_CorruptedPrivateJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse private JWK")
}

// TestJWTService_CreateJWT_UnsupportedAlgorithm tests CreateJWT with unsupported oct-128 algorithm.
func TestJWTService_CreateJWT_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create signer")
}

// TestJWTService_ValidateJWT_CorruptedPublicJWKParse tests ValidateJWT with corrupted public JWK.
func TestJWTService_ValidateJWT_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

// TestJWTService_CreateEncryptedJWT_CorruptedPublicJWKParse tests CreateEncryptedJWT with corrupted JWK.
func TestJWTService_CreateEncryptedJWT_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

// TestJWTService_CreateEncryptedJWT_UnsupportedEncryptionAlgorithm tests CreateEncryptedJWT with EdDSA enc.
func TestJWTService_CreateEncryptedJWT_UnsupportedEncryptionAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWE")
}

// TestJWTService_CreateJWT_NoActiveMaterial tests CreateJWT after retiring material.
func TestJWTService_CreateJWT_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

// TestCreateEncryptedJWT_NoActiveMaterial tests CreateEncryptedJWT after retiring enc material.
func TestCreateEncryptedJWT_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, encJWK.ID, encMaterial.ID)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encryption material")
}

// TestJWTService_ValidateJWT_BarrierDecryptFailure tests ValidateJWT when barrier decrypt fails.
func TestJWTService_ValidateJWT_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

// TestJWTService_CreateJWT_BarrierDecryptFailure tests CreateJWT when barrier decrypt fails.
func TestJWTService_CreateJWT_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt private JWK")
}

// TestCreateEncryptedJWT_BarrierDecryptFailure tests CreateEncryptedJWT when barrier decrypt fails.
func TestCreateEncryptedJWT_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

// TestJWTService_CreateJWT_AlgorithmKeyMismatch tests CreateJWT when algorithm mismatches key type.
func TestJWTService_CreateJWT_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	// Create EC key then tamper algorithm to RSA.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create signer")
}

// TestCreateEncryptedJWT_AlgorithmKeyMismatch tests CreateEncryptedJWT when enc key algorithm mismatches.
func TestCreateEncryptedJWT_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	// Create RSA enc key, then tamper algorithm to EC.
	encJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.ElasticJWK{}).Where("id = ?", encJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create encrypter")
}

// TestJWTService_CreateJWT_Base64DecodeFailure tests CreateJWT when base64 decode of private JWK fails.
func TestJWTService_CreateJWT_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// TestJWTService_ValidateJWT_Base64DecodeFailure tests ValidateJWT when base64 decode fails.
func TestJWTService_ValidateJWT_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestCreateEncryptedJWT_Base64DecodeFailure tests CreateEncryptedJWT when base64 decode fails.
func TestCreateEncryptedJWT_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestCreateEncryptedJWT_SigningNoActiveMaterial tests CreateEncryptedJWT after retiring signing material.
func TestCreateEncryptedJWT_SigningNoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, sigMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, sigJWK.ID, sigMaterial.ID)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

// TestValidateJWT_MaterialKIDNotFound tests ValidateJWT after material is deleted from DB.
func TestValidateJWT_MaterialKIDNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}

	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	// Delete material from DB so KID lookup fails.
	err = testDB.Delete(&cryptoutilAppsJoseJaDomain.MaterialJWK{}, "id = ?", material.ID).Error
	require.NoError(t, err)

	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material by KID")
}

// TestValidateJWT_JWTValidationFailed tests ValidateJWT with wrong public key (signature mismatch).
func TestValidateJWT_JWTValidationFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs with different keys.
	elasticJWK_A, materialA, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	_, materialB, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}

	// Create JWT signed with key A.
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK_A.ID, claims)
	require.NoError(t, err)

	// Swap material A's public key with material B's public key.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).Where("id = ?", materialA.ID).Update("public_jwk_jwe", materialB.PublicJWKJWE).Error
	require.NoError(t, err)

	// Validate: public key B won't verify JWT signed with A's private key.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK_A.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWT validation failed")
}

// TestValidateJWT_FallbackNoActiveMaterial tests ValidateJWT fallback path when JWT has no kid.
func TestValidateJWT_FallbackNoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Generate a standalone RSA key and sign a JWT WITHOUT kid header.
	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	signingKey := jose.SigningKey{Algorithm: jose.RS256, Key: rsaKey}
	signer, err := jose.NewSigner(signingKey, nil) // No kid header.
	require.NoError(t, err)

	claimsMap := map[string]any{
		cryptoutilSharedMagic.ClaimSub: "test",
		cryptoutilSharedMagic.ClaimExp: time.Now().UTC().Add(time.Hour).Unix(),
	}

	builder := jwt.Signed(signer).Claims(claimsMap)
	token, err := builder.Serialize()
	require.NoError(t, err)

	// Retire material so GetActiveMaterial fails.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)

	// Validate: kid is empty → fallback to GetActiveMaterial → no active material.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}
