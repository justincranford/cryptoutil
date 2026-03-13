package service

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose/ja/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	googleUuid "github.com/google/uuid"
)

func TestJWTService_CreateEncryptedJWTDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	claims := &JWTClaims{Issuer: "test-issuer"}
	_, err := svc.CreateEncryptedJWT(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New(), claims)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWTService_CreateEncryptedJWT_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

func TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create signing key for tenant1.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create encryption key for tenant2.
	encryptionKey, _, err := elasticSvc.CreateElasticJWK(ctx, otherTenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

func TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create another signing key (not encryption).
	anotherSigningKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

func TestJWTService_CreateEncryptedJWT_UnsupportedEncryptionAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm for JWE")
}

func TestJWTService_CreateJWTDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	claims := &JWTClaims{Issuer: "test-issuer"}
	_, err := svc.CreateJWT(ctx, googleUuid.New(), googleUuid.New(), claims)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWTService_CreateJWT_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	// Create EC key then tamper algorithm to RSA.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", elasticJWK.ID).Update("alg", cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create signer")
}

func TestJWTService_CreateJWT_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt private JWK")
}

func TestJWTService_CreateJWT_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

func TestJWTService_CreateJWT_CorruptedPrivateJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("private_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse private JWK")
}

func TestJWTService_CreateJWT_CorruptedPrivateKeyDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
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

func TestJWTService_CreateJWT_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK.ID, material.ID)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

func TestJWTService_CreateJWT_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create signer")
}

func TestJWTService_CreateJWT_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key (not signing).
	encryptionKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

func TestJWTService_ValidateJWTDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	_, err := svc.ValidateJWT(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0ZXN0In0.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or validate.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWTService_ValidateJWT_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

func TestJWTService_ValidateJWT_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestJWTService_ValidateJWT_CorruptedPublicJWKParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)

	invalidJSON := []byte("not-valid-json")
	encryptedInvalid, err := testBarrierService.EncryptContentWithContext(ctx, invalidJSON)
	require.NoError(t, err)

	corruptedJWE := base64.StdEncoding.EncodeToString(encryptedInvalid)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", material.ID).Update("public_jwk_jwe", corruptedJWE).Error
	require.NoError(t, err)
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse public JWK")
}

func TestJWTService_ValidateJWT_CorruptedPublicKeyDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to validate - should fail on base64 decode.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestJWTService_ValidateJWT_MaterialKIDNotBelongToElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	signingKey1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	signingKey2, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

func TestJWTService_ValidateJWT_NoHeaders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	// Create a malformed token (not a valid JWT structure).
	invalidToken := "invalid.token.structure"

	// Try to validate - should fail parsing.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, signingKey.ID, invalidToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWT")
}
