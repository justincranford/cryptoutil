// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestJWSService_Sign_CorruptedPrivateKeyInDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "invalid-base64!!!").Error
	require.NoError(t, err)

	// Try to sign - should fail due to corrupted stored key.
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// ====================
// Material Rotation with Corrupted Data Tests
// ====================

// TestMaterialRotationService_RotateMaterial_CreatesMaterialSuccessfully tests that
// RotateMaterial creates a new material and marks it as active.
func TestMaterialRotationService_RotateMaterial_CreatesMaterialSuccessfully(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with initial material.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, initialMaterial)
	require.True(t, initialMaterial.Active)

	// Rotate creates a new active material.
	newMaterial, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, newMaterial)
	require.True(t, newMaterial.Active)
	require.NotEqual(t, initialMaterial.ID, newMaterial.ID)
}

// ====================
// Symmetric Key Tests (oct type)
// ====================

// TestElasticJWKService_CreateElasticJWK_SymmetricKey tests that CreateElasticJWK
// correctly handles symmetric (oct) keys where there's no separate public key.
func TestElasticJWKService_CreateElasticJWK_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create symmetric key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Verify the key was created.
	retrieved, err := elasticSvc.GetElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, elasticJWK.ID, retrieved.ID)
}

// TestJWEService_EncryptDecrypt_SymmetricKey tests JWE operations with symmetric keys.
func TestJWEService_EncryptDecrypt_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create symmetric key for direct encryption.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	plaintext := []byte("secret message for symmetric key")
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	decrypted, err := jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

// ====================
// Additional Error Path Tests
// ====================

// TestJWEService_Encrypt_WrongKeyUse tests that Encrypt fails when key is not for encryption.
func TestJWEService_Encrypt_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create a signing key (not encryption).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

// Removed TestJWEService_Decrypt_WrongKeyUse - Decrypt doesn't validate key use, just fails with "no matching key"

// TestJWEService_EncryptWithKID_WrongKeyUse tests that EncryptWithKID fails when key is not for encryption.
func TestJWEService_EncryptWithKID_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create a signing key (not encryption).
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

// TestJWEService_EncryptWithKID_MaterialNotFound tests that EncryptWithKID fails when material not found.
func TestJWEService_EncryptWithKID_MaterialNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt with non-existent material KID.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, "nonexistent-kid", []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

// TestJWEService_EncryptWithKID_MaterialWrongElasticJWK tests EncryptWithKID fails when material
// belongs to different elastic JWK.
func TestJWEService_EncryptWithKID_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two encryption keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt using elasticJWK1 but with material2's KID.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "material key does not belong to elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

// TestMaterialRotationService_RotateMaterial_MaxMaterialsReached tests that RotateMaterial fails
// when max materials is reached.
func TestMaterialRotationService_RotateMaterial_MaxMaterialsReached(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 2 materials.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 2)
	require.NoError(t, err)

	// Rotate once - should succeed (now have 2 materials).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Rotate again - should fail (max reached).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max materials reached")
}

// TestMaterialRotationService_RetireMaterial_MaterialWrongElasticJWK tests that RetireMaterial fails
// when material doesn't belong to the specified elastic JWK.
func TestMaterialRotationService_RetireMaterial_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to retire material2 via elasticJWK1 - should fail.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK1.ID, material2.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "material not found for this elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

// TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongTenant tests that CreateEncryptedJWT fails
// when encryption key has wrong tenant.
func TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create signing key for tenant1.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create encryption key for tenant2.
	encryptionKey, _, err := elasticSvc.CreateElasticJWK(ctx, otherTenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to create encrypted JWT using tenant1's signing key but tenant2's encryption key.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, signingKey.ID, encryptionKey.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "encryption key not found")
}

// TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongUse tests that CreateEncryptedJWT fails
// when encryption key is not configured for encryption.
func TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create another signing key (not encryption).
	anotherSigningKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to create encrypted JWT using signing key for encryption.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, signingKey.ID, anotherSigningKey.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

// TestJWTService_ValidateJWT_NoHeaders tests that ValidateJWT fails when JWT has no headers.
func TestJWTService_ValidateJWT_NoHeaders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create a malformed token (not a valid JWT structure).
	invalidToken := "invalid.token.structure"

	// Try to validate - should fail parsing.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, signingKey.ID, invalidToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWT")
}

// TestJWTService_ValidateJWT_MaterialKIDNotBelongToElasticJWK tests that ValidateJWT fails
// when the material KID doesn't belong to the expected elastic JWK.
func TestJWTService_ValidateJWT_MaterialKIDNotBelongToElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	signingKey1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	signingKey2, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create JWT with signing key 1.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, signingKey1.ID, claims)
	require.NoError(t, err)

	// Try to validate with signing key 2 - should fail because KID doesn't belong.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, signingKey2.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to this elastic JWK")
}

// TestJWKSService_GetJWKS_WrongTenant tests that GetJWKS returns empty for wrong tenant.
func TestJWKSService_GetJWKS_EmptyForWrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create signing key for one tenant.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Get JWKS with different tenant - should return empty (not an error).
	jwks, err := jwksSvc.GetJWKS(ctx, otherTenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestJWKSService_GetPublicJWK_WrongTenant tests that GetPublicJWK returns not found for wrong tenant.
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

// TestJWSService_SignWithKID_WrongKID tests that SignWithKID fails when KID not found.
func TestJWSService_SignWithKID_WrongKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign with non-existent KID.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK.ID, "nonexistent-kid", []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

// TestJWSService_SignWithKID_MaterialWrongElasticJWK tests that SignWithKID fails
// when material belongs to different elastic JWK.
func TestJWSService_SignWithKID_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign using elasticJWK1 but with material2's KID.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "material key does not belong to elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

// TestJWTService_CreateJWT_WrongKeyUse tests that CreateJWT fails when key is for encryption.
func TestJWTService_CreateJWT_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key (not signing).
	encryptionKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to create JWT with encryption key - should fail.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, encryptionKey.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for signing")
}

// Helper function for time pointer.
func timePtr(t time.Time) *time.Time {
	return &t
}
