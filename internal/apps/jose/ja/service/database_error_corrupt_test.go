// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestJWSService_Sign_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE with invalid base64 directly in DB.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to sign - should fail on base64 decode.
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// TestJWSService_Sign_CorruptedJWE tests that Sign returns error when
// material's PrivateJWKJWE contains valid base64 but invalid JWE content.
func TestJWSService_Sign_CorruptedJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE with valid base64 but invalid JWE.
	// Use base64 encoding of "not a valid JWE" string.
	invalidJWE := "bm90IGEgdmFsaWQgSldF" // base64("not a valid JWE")
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", invalidJWE).Error
	require.NoError(t, err)

	// Try to sign - should fail on barrier decrypt.
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt private JWK")
}

// TestJWEService_Encrypt_CorruptedBase64 tests that Encrypt returns error when
// material's PublicJWKJWE contains invalid base64.
func TestJWEService_Encrypt_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	// Use RSA/2048 key type which maps to RSA-OAEP-256 algorithm for encryption.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE with invalid base64.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to encrypt - should fail on base64 decode.
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWEService_Decrypt_CorruptedBase64 tests that Decrypt returns error when
// material's PrivateJWKJWE contains invalid base64.
// NOTE: Decrypt loops through all materials and catches errors silently,
// so when decoding fails it returns "no matching key found" not the decode error.
func TestJWEService_Decrypt_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	// Use RSA/2048 key type which maps to RSA-OAEP-256 algorithm for encryption.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// First encrypt something valid.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Now corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to decrypt - Decrypt catches decode errors and tries next material.
	// Since we only have one material and it fails, returns "no matching key found".
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWSService_Verify_CorruptedBase64 tests that Verify returns error when
// material's PublicJWKJWE contains invalid base64.
// NOTE: Verify loops through all materials and catches errors silently,
// so when decoding fails it returns "no matching key found" not the decode error.
func TestJWSService_Verify_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// First sign something valid.
	jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)
	require.NotEmpty(t, jwsCompact)

	// Now corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to verify - Verify catches decode errors and tries next material.
	// Since we only have one material and it fails, returns "no matching key found".
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, jwsCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWKSService_GetPublicJWK_CorruptedBase64 tests that GetPublicJWK returns
// error when material's PublicJWKJWE contains invalid base64.
func TestJWKSService_GetPublicJWK_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
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

// TestJWKSService_GetJWKS_CorruptedBase64 tests that GetJWKS skips materials
// with corrupted PublicJWKJWE (graceful degradation).
// NOTE: GetJWKS uses continue pattern - it skips corrupted materials rather than failing.
func TestJWKSService_GetJWKS_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
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

// ====================
// JWT Service Corrupted Key Tests
// ====================

// TestJWTService_ValidateJWT_CorruptedPublicKeyDB tests that ValidateJWT returns error
// when material's PublicJWKJWE is corrupted in the database.
func TestJWTService_ValidateJWT_CorruptedPublicKeyDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Create a valid JWT first.
	claims := &JWTClaims{
		Subject:  "test-subject",
		Issuer:   "test-issuer",
		Audience: []string{"test-audience"},
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to validate - should fail on base64 decode.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWTService_CreateJWT_CorruptedPrivateKeyDB tests that CreateJWT returns error
// when material's PrivateJWKJWE is corrupted in the database.
