package service

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	crand "crypto/rand"
	rsa "crypto/rsa"
	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func TestCreateEncryptedJWT_AlgorithmKeyMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	// Create RSA enc key, then tamper algorithm to EC.
	encJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.ElasticJWK{}).Where("id = ?", encJWK.ID).Update("alg", cryptoutilSharedMagic.JoseKeyTypeECP256).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create encrypter")
}

func TestCreateEncryptedJWT_BarrierDecryptFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	corruptedBase64 := base64.StdEncoding.EncodeToString([]byte("not-barrier"))
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", corruptedBase64).Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt public JWK")
}

func TestCreateEncryptedJWT_Base64DecodeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", encMaterial.ID).Update("public_jwk_jwe", "===invalid===").Error
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

func TestCreateEncryptedJWT_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, encMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, encJWK.ID, encMaterial.ID)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encryption material")
}

func TestCreateEncryptedJWT_SigningNoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	sigJWK, sigMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	encJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyUseEnc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)
	err = rotationSvc.RetireMaterial(ctx, tenantID, sigJWK.ID, sigMaterial.ID)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, sigJWK.ID, encJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get active material")
}

func TestValidateJWT_FallbackNoActiveMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

func TestValidateJWT_JWTValidationFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs with different keys.
	elasticJWK_A, materialA, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	_, materialB, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	expiry := time.Now().UTC().Add(time.Hour)
	claims := &JWTClaims{Subject: "test", ExpiresAt: &expiry}

	// Create JWT signed with key A.
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK_A.ID, claims)
	require.NoError(t, err)

	// Swap material A's public key with material B's public key.
	err = testDB.Model(&cryptoutilAppsJoseJaModel.MaterialJWK{}).Where("id = ?", materialA.ID).Update("public_jwk_jwe", materialB.PublicJWKJWE).Error
	require.NoError(t, err)

	// Validate: public key B won't verify JWT signed with A's private key.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK_A.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWT validation failed")
}

func TestValidateJWT_MaterialKIDNotFound(t *testing.T) {
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

	// Delete material from DB so KID lookup fails.
	err = testDB.Delete(&cryptoutilAppsJoseJaModel.MaterialJWK{}, "id = ?", material.ID).Error
	require.NoError(t, err)

	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material by KID")
}
