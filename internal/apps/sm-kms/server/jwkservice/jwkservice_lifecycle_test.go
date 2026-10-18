// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"testing"
	"time"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestElasticJWKService_Lifecycle(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	service := newTestElasticJWKService()

	elasticJWK, material, err := service.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseAlgES256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		0,
	)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, elasticJWK.MaxMaterials)

	got, err := service.GetElasticJWK(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, elasticJWK.ID, got.ID)

	keys, total, err := service.ListElasticJWKs(testContext, tenantID, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, keys, 1)

	require.NoError(t, service.DeleteElasticJWK(testContext, tenantID, elasticJWK.ID))
	_, err = service.GetElasticJWK(testContext, tenantID, elasticJWK.ID)
	require.ErrorContains(t, err, "not found")
}

func TestElasticJWKService_ValidationAndOwnership(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	service := newTestElasticJWKService()

	tests := []struct {
		name      string
		algorithm string
		use       string
	}{
		{name: testErrorCaseUnsupportedAlg, algorithm: testErrorUnsupported, use: cryptoutilAppsSmKmsModel.KeyUseSig},
		{name: "unsupported use", algorithm: cryptoutilSharedMagic.JoseAlgES256, use: testErrorUnsupported},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := service.CreateElasticJWK(testContext, tenantID, tc.algorithm, tc.use, 1)
			require.Error(t, err)
		})
	}

	elasticJWK, _, err := service.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseAlgES256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		1,
	)
	require.NoError(t, err)

	_, err = service.GetElasticJWK(testContext, newTestUUID(t), elasticJWK.ID)
	require.ErrorContains(t, err, "not found")

	otherTenant := newTestUUID(t)
	_, _, err = service.ListElasticJWKs(testContext, otherTenant, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
	require.NoError(t, err)

	err = service.DeleteElasticJWK(testContext, otherTenant, elasticJWK.ID)
	require.ErrorContains(t, err, "not found")
}

func TestMaterialRotationService_Lifecycle(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticService := newTestElasticJWKService()
	elasticJWK, originalMaterial, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseKeyTypeECP256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		2,
	)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
	rotated, err := service.RotateMaterial(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotEqual(t, originalMaterial.MaterialKID, rotated.MaterialKID)

	active, err := service.GetActiveMaterial(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, rotated.ID, active.ID)

	materials, err := service.ListMaterials(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, materials, 2)

	byKID, err := service.GetMaterialByKID(testContext, tenantID, elasticJWK.ID, originalMaterial.MaterialKID)
	require.NoError(t, err)
	require.Equal(t, originalMaterial.ID, byKID.ID)

	require.NoError(t, service.RetireMaterial(testContext, tenantID, elasticJWK.ID, originalMaterial.ID))
	retired, err := materialRepo.GetByID(testContext, originalMaterial.ID)
	require.NoError(t, err)
	require.NotNil(t, retired.RetiredAt)

	_, err = service.RotateMaterial(testContext, tenantID, elasticJWK.ID)
	require.ErrorContains(t, err, testErrorCaseMaxMaterials)
}

func TestMaterialRotationService_OwnershipAndLookupErrors(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticService := newTestElasticJWKService()
	elasticJWK, material, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseKeyTypeECP256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		2,
	)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
	otherTenant := newTestUUID(t)

	_, err = service.RotateMaterial(testContext, otherTenant, elasticJWK.ID)
	require.ErrorContains(t, err, "not found")
	err = service.RetireMaterial(testContext, otherTenant, elasticJWK.ID, material.ID)
	require.ErrorContains(t, err, "not found")
	_, err = service.ListMaterials(testContext, otherTenant, elasticJWK.ID)
	require.ErrorContains(t, err, "not found")
	_, err = service.GetActiveMaterial(testContext, otherTenant, elasticJWK.ID)
	require.ErrorContains(t, err, "not found")
	_, err = service.GetMaterialByKID(testContext, otherTenant, elasticJWK.ID, material.MaterialKID)
	require.ErrorContains(t, err, "not found")

	anotherKeyID := newTestUUID(t)
	_, err = service.GetActiveMaterial(testContext, tenantID, anotherKeyID)
	require.Error(t, err)
	_, err = service.GetMaterialByKID(testContext, tenantID, elasticJWK.ID, "absent-kid")
	require.Error(t, err)

	foreignMaterialID := newTestUUID(t)
	foreignMaterial := &cryptoutilAppsSmKmsModel.MaterialJWK{
		ID:           foreignMaterialID,
		ElasticJWKID: newTestUUID(t),
		MaterialKID:  "foreign-kid-" + foreignMaterialID.String(),
	}
	require.NoError(t, testDB.Create(foreignMaterial).Error)
	err = service.RetireMaterial(testContext, tenantID, elasticJWK.ID, foreignMaterial.ID)
	require.ErrorContains(t, err, "not found")
	_, err = service.GetMaterialByKID(testContext, tenantID, elasticJWK.ID, foreignMaterial.MaterialKID)
	require.ErrorContains(t, err, "not found")
}

func TestJWEService_RoundTripAndMaterialSelection(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticService := newTestElasticJWKService()
	elasticJWK, originalMaterial, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseKeyTypeECP256,
		cryptoutilAppsSmKmsModel.KeyUseEnc,
		3,
	)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewJWEService(elasticRepo, materialRepo, testBarrierService)
	plaintext := []byte("jwe lifecycle payload")

	compact, err := service.Encrypt(testContext, tenantID, elasticJWK.ID, plaintext)
	require.NoError(t, err)
	decrypted, err := service.Decrypt(testContext, tenantID, elasticJWK.ID, compact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	compact, err = service.EncryptWithKID(testContext, tenantID, elasticJWK.ID, originalMaterial.MaterialKID, plaintext)
	require.NoError(t, err)
	decrypted, err = service.Decrypt(testContext, tenantID, elasticJWK.ID, compact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	rotationService := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
	_, err = rotationService.RotateMaterial(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	decrypted, err = service.Decrypt(testContext, tenantID, elasticJWK.ID, compact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	_, err = service.Decrypt(testContext, tenantID, elasticJWK.ID, "malformed")
	require.ErrorContains(t, err, "failed to parse JWE")
	_, err = service.Decrypt(testContext, newTestUUID(t), elasticJWK.ID, compact)
	require.ErrorContains(t, err, "not found")
	_, err = service.Encrypt(testContext, newTestUUID(t), elasticJWK.ID, plaintext)
	require.ErrorContains(t, err, "not found")
}

func TestJWSService_RoundTripAndPublicKeys(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticService := newTestElasticJWKService()
	elasticJWK, originalMaterial, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseAlgES256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		3,
	)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	jwsService := NewJWSService(elasticRepo, materialRepo, testBarrierService)
	payload := []byte("jws lifecycle payload")

	compact, err := jwsService.Sign(testContext, tenantID, elasticJWK.ID, payload)
	require.NoError(t, err)
	verified, err := jwsService.Verify(testContext, tenantID, elasticJWK.ID, compact)
	require.NoError(t, err)
	require.Equal(t, payload, verified)

	compact, err = jwsService.SignWithKID(testContext, tenantID, elasticJWK.ID, originalMaterial.MaterialKID, payload)
	require.NoError(t, err)
	verified, err = jwsService.Verify(testContext, tenantID, elasticJWK.ID, compact)
	require.NoError(t, err)
	require.Equal(t, payload, verified)

	rotationService := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
	_, err = rotationService.RotateMaterial(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	verified, err = jwsService.Verify(testContext, tenantID, elasticJWK.ID, compact)
	require.NoError(t, err)
	require.Equal(t, payload, verified)

	_, err = jwsService.Verify(testContext, tenantID, elasticJWK.ID, "malformed")
	require.ErrorContains(t, err, "failed to parse JWS")

	jwksService := NewJWKSService(elasticRepo, materialRepo, testBarrierService)
	jwks, err := jwksService.GetJWKS(testContext, tenantID)
	require.NoError(t, err)
	require.NotEmpty(t, jwks.Keys)

	singleKey, err := jwksService.GetPublicJWK(testContext, tenantID, originalMaterial.MaterialKID)
	require.NoError(t, err)
	require.Equal(t, originalMaterial.MaterialKID, singleKey.KeyID)
	require.Equal(t, cryptoutilAppsSmKmsModel.KeyUseSig, singleKey.Use)

	keySet, err := jwksService.GetJWKSForElasticKey(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, keySet.Keys, 1)
	_, err = jwksService.GetJWKSForElasticKey(testContext, newTestUUID(t), elasticJWK.ID)
	require.ErrorContains(t, err, "not found")
	_, err = jwksService.GetPublicJWK(testContext, newTestUUID(t), originalMaterial.MaterialKID)
	require.ErrorContains(t, err, "not found")
	_, err = jwksService.GetPublicJWK(testContext, tenantID, "absent-kid")
	require.Error(t, err)
}

func TestJWTService_CreateValidateAndEncrypt(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticService := newTestElasticJWKService()
	signingKey, _, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseAlgES256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		3,
	)
	require.NoError(t, err)
	encryptionKey, _, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseKeyTypeECP256,
		cryptoutilAppsSmKmsModel.KeyUseEnc,
		3,
	)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewJWTService(elasticRepo, materialRepo, testBarrierService)
	now := time.Now().UTC()
	expiresAt := now.Add(time.Hour)
	notBefore := now.Add(-time.Minute)
	claims := &JWTClaims{
		Issuer:    "https://issuer.example.test",
		Subject:   "test-subject",
		Audience:  []string{"audience-one", "audience-two"},
		ExpiresAt: &expiresAt,
		NotBefore: &notBefore,
		IssuedAt:  &now,
		JTI:       newTestUUID(t).String(),
		Custom:    map[string]any{"role": "operator"},
	}
	token, err := service.CreateJWT(testContext, tenantID, signingKey.ID, claims)
	require.NoError(t, err)

	validated, err := service.ValidateJWT(testContext, tenantID, signingKey.ID, token)
	require.NoError(t, err)
	require.Equal(t, claims.Issuer, validated.Issuer)
	require.Equal(t, claims.Subject, validated.Subject)
	require.Equal(t, claims.Audience, validated.Audience)
	require.Equal(t, claims.JTI, validated.JTI)
	require.Equal(t, "operator", validated.Custom["role"])

	encrypted, err := service.CreateEncryptedJWT(testContext, tenantID, signingKey.ID, encryptionKey.ID, claims)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	expired := now.Add(-time.Hour)
	expiredClaims := &JWTClaims{ExpiresAt: &expired}
	expiredToken, err := service.CreateJWT(testContext, tenantID, signingKey.ID, expiredClaims)
	require.NoError(t, err)
	_, err = service.ValidateJWT(testContext, tenantID, signingKey.ID, expiredToken)
	require.ErrorContains(t, err, "expired")

	future := now.Add(time.Hour)
	futureClaims := &JWTClaims{NotBefore: &future}
	futureToken, err := service.CreateJWT(testContext, tenantID, signingKey.ID, futureClaims)
	require.NoError(t, err)
	_, err = service.ValidateJWT(testContext, tenantID, signingKey.ID, futureToken)
	require.ErrorContains(t, err, "not yet valid")

	_, err = service.ValidateJWT(testContext, tenantID, signingKey.ID, "malformed")
	require.ErrorContains(t, err, "failed to parse JWT")
	_, err = service.CreateJWT(testContext, newTestUUID(t), signingKey.ID, claims)
	require.ErrorContains(t, err, "not found")
	_, err = service.CreateEncryptedJWT(testContext, tenantID, signingKey.ID, newTestUUID(t), claims)
	require.Error(t, err)
}

func TestAuditLogService_PersistenceAndDefaults(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticService := newTestElasticJWKService()
	elasticJWK, _, err := elasticService.CreateElasticJWK(
		testContext,
		tenantID,
		cryptoutilSharedMagic.JoseAlgES256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		1,
	)
	require.NoError(t, err)

	elasticRepo, _ := testRepositories()

	auditLogRepo := cryptoutilAppsSmKmsRepository.NewAuditLogRepository(testDB)
	auditConfigRepo := cryptoutilAppsSmKmsRepository.NewAuditConfigRepository(testDB)
	service := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	config, err := service.GetAuditConfig(testContext, tenantID)
	require.NoError(t, err)
	require.True(t, config.Enabled)
	require.Equal(t, float64(1), config.SamplingRate)

	config.Operation = cryptoutilAppsSmKmsModel.OperationSign
	config.SamplingRate = 1
	require.NoError(t, service.UpdateAuditConfig(testContext, tenantID, config))
	require.Equal(t, tenantID, config.TenantID)

	errorMessage := "sample operation failed"
	require.NoError(t, service.LogOperation(
		testContext,
		tenantID,
		&elasticJWK.ID,
		cryptoutilAppsSmKmsModel.OperationSign,
		newTestUUID(t).String(),
		false,
		&errorMessage,
	))

	logs, total, err := service.ListAuditLogs(testContext, tenantID, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, logs, 1)
	require.Equal(t, errorMessage, *logs[0].ErrorMessage)

	logs, total, err = service.ListAuditLogsByElasticJWK(testContext, tenantID, elasticJWK.ID, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, logs, 1)
	logs, total, err = service.ListAuditLogsByOperation(testContext, tenantID, cryptoutilAppsSmKmsModel.OperationSign, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, logs, 1)

	_, _, err = service.ListAuditLogsByElasticJWK(testContext, newTestUUID(t), elasticJWK.ID, 0, 10)
	require.ErrorContains(t, err, "not found")
	count, err := service.CleanupOldLogs(testContext, tenantID, 1)
	require.NoError(t, err)
	require.Zero(t, count)
}

func newTestElasticJWKService() ElasticJWKService {
	elasticRepo, materialRepo := testRepositories()

	return NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
}

func newTestUUID(t *testing.T) googleUuid.UUID {
	t.Helper()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	return id
}
