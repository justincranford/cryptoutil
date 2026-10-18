// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"encoding/base64"
	json "encoding/json"
	"testing"
	"time"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWTService_CreateJWTFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 2)
	require.NoError(t, err)

	tests := []struct {
		name          string
		getError      bool
		wrongTenant   bool
		wrongUse      bool
		materialError bool
		privateJWKJWE string
		algorithm     string
		claims        *JWTClaims
		wantError     string
	}{
		{name: testErrorCaseElasticLookup, getError: true, wantError: testErrorFailedGetElastic},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: testErrorElasticJWKNotFound},
		{name: testErrorCaseWrongKeyUse, wrongUse: true, wantError: testErrorNotConfiguredSign},
		{name: testErrorCaseActiveMaterial, materialError: true, wantError: testErrorFailedGetActive},
		{name: "private JWK base64 error", privateJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePrivate},
		{name: "private JWK decrypt error", privateJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad")), wantError: testErrorFailedDecryptPrivate},
		{name: "private JWK parse error", privateJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePrivate},
		{name: "signer error", algorithm: cryptoutilSharedMagic.JoseAlgHS256, wantError: testErrorFailedCreateSigner},
		{name: "claims serialization error", claims: &JWTClaims{Custom: map[string]any{cryptoutilSharedMagic.DevSetupTestInvalidName: make(chan int)}}, wantError: "failed to create JWT"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			elastic := *elasticJWK
			testMaterial := *material

			if tc.wrongTenant {
				elastic.TenantID = newTestUUID(t)
			}

			if tc.wrongUse {
				elastic.Use = cryptoutilAppsSmKmsModel.KeyUseEnc
			}

			if tc.privateJWKJWE != "" {
				testMaterial.PrivateJWKJWE = tc.privateJWKJWE
			}

			if tc.algorithm != "" {
				elastic.Algorithm = tc.algorithm
			}

			elasticRepo := &elasticJWKRepositoryStub{
				getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
					if tc.getError {
						return nil, errJWKServiceInjected
					}

					return &elastic, nil
				},
			}
			materialRepo := &materialJWKRepositoryStub{
				getActiveFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					if tc.materialError {
						return nil, errJWKServiceInjected
					}

					return &testMaterial, nil
				},
			}

			claims := tc.claims
			if claims == nil {
				claims = &JWTClaims{Subject: testJWTSubject}
			}

			_, err := NewJWTService(elasticRepo, materialRepo, testBarrierService).CreateJWT(testContext, tenantID, elastic.ID, claims)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestJWTService_ValidateJWTFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 2)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewJWTService(elasticRepo, materialRepo, testBarrierService)
	token, err := service.CreateJWT(testContext, tenantID, elasticJWK.ID, &JWTClaims{Subject: testJWTSubject})
	require.NoError(t, err)

	_, err = service.ValidateJWT(testContext, tenantID, newTestUUID(t), token)
	require.Error(t, err)
	_, err = service.ValidateJWT(testContext, tenantID, elasticJWK.ID, "malformed")
	require.ErrorContains(t, err, "failed to parse JWT")

	foreignTenantJWK := *elasticJWK
	foreignTenantJWK.TenantID = newTestUUID(t)
	elasticRepoStub := &elasticJWKRepositoryStub{
		getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
			return &foreignTenantJWK, nil
		},
	}
	_, err = NewJWTService(elasticRepoStub, materialRepo, testBarrierService).ValidateJWT(testContext, tenantID, elasticJWK.ID, token)
	require.ErrorContains(t, err, testErrorElasticJWKNotFound)

	elasticRepoStub.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return elasticJWK, nil
	}
	materialRepoStub := &materialJWKRepositoryStub{
		getByKIDFn: func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
			return nil, errJWKServiceInjected
		},
	}
	_, err = NewJWTService(elasticRepoStub, materialRepoStub, testBarrierService).ValidateJWT(testContext, tenantID, elasticJWK.ID, token)
	require.ErrorContains(t, err, testErrorFailedGetMaterialKID)

	materialRepoStub.getByKIDFn = func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
		foreignMaterial := *material
		foreignMaterial.ElasticJWKID = newTestUUID(t)

		return &foreignMaterial, nil
	}
	_, err = NewJWTService(elasticRepoStub, materialRepoStub, testBarrierService).ValidateJWT(testContext, tenantID, elasticJWK.ID, token)
	require.ErrorContains(t, err, testErrorMaterialNotBelong)

	_, mismatchedMaterial, err := createTestElasticKey(
		t,
		tenantID,
		cryptoutilSharedMagic.JoseAlgES256,
		cryptoutilAppsSmKmsModel.KeyUseSig,
		1,
	)
	require.NoError(t, err)

	for _, tc := range []struct {
		name         string
		publicJWKJWE string
		wantError    string
	}{
		{name: testErrorCasePublicJWKBase64, publicJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePublic},
		{name: testErrorCasePublicJWKDecrypt, publicJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad")), wantError: testErrorFailedDecryptPublic},
		{name: testErrorCasePublicJWKParse, publicJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePublic},
		{name: "signature validation error", publicJWKJWE: mismatchedMaterial.PublicJWKJWE, wantError: "JWT validation failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testMaterial := *material
			testMaterial.PublicJWKJWE = tc.publicJWKJWE
			testMaterialRepo := &materialJWKRepositoryStub{
				getByKIDFn: func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					return &testMaterial, nil
				},
			}
			_, validateErr := NewJWTService(elasticRepoStub, testMaterialRepo, testBarrierService).ValidateJWT(
				testContext,
				tenantID,
				elasticJWK.ID,
				token,
			)
			require.ErrorContains(t, validateErr, tc.wantError)
		})
	}

	expiredToken, err := service.CreateJWT(testContext, tenantID, elasticJWK.ID, &JWTClaims{
		ExpiresAt: timePointer(time.Now().UTC().Add(-time.Minute)),
	})
	require.NoError(t, err)
	_, err = service.ValidateJWT(testContext, tenantID, elasticJWK.ID, expiredToken)
	require.ErrorContains(t, err, "JWT has expired")

	futureToken, err := service.CreateJWT(testContext, tenantID, elasticJWK.ID, &JWTClaims{
		NotBefore: timePointer(time.Now().UTC().Add(time.Hour)),
	})
	require.NoError(t, err)
	_, err = service.ValidateJWT(testContext, tenantID, elasticJWK.ID, futureToken)
	require.ErrorContains(t, err, "JWT is not yet valid")
}

func TestJWTService_CreateEncryptedJWTFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	signingJWK, signingMaterial, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 2)
	require.NoError(t, err)
	encryptionJWK, encryptionMaterial, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilAppsSmKmsModel.KeyUseEnc, 2)
	require.NoError(t, err)

	tests := []struct {
		name          string
		signingError  bool
		getError      bool
		wrongTenant   bool
		wrongUse      bool
		materialError bool
		publicJWKJWE  string
		algorithm     string
		wantError     string
	}{
		{name: "signing error", signingError: true, wantError: "failed to create signed JWT"},
		{name: "encryption lookup error", getError: true, wantError: "failed to get encryption key"},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: "encryption key not found"},
		{name: testErrorCaseWrongKeyUse, wrongUse: true, wantError: testErrorNotConfiguredEncrypt},
		{name: testErrorCaseActiveMaterial, materialError: true, wantError: "failed to get encryption material"},
		{name: testErrorCasePublicJWKBase64, publicJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePublic},
		{name: testErrorCasePublicJWKDecrypt, publicJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad")), wantError: testErrorFailedDecryptPublic},
		{name: testErrorCasePublicJWKParse, publicJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePublic},
		{name: testErrorCaseUnsupportedAlg, algorithm: testErrorUnsupported, wantError: testErrorUnsupportedAlgorithmForJWE},
		{name: "encrypter error", algorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP256, wantError: testErrorFailedCreateEncrypter},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			signingKey := *signingJWK
			encryptionKey := *encryptionJWK
			testSigningMaterial := *signingMaterial
			testEncryptionMaterial := *encryptionMaterial

			if tc.wrongTenant {
				encryptionKey.TenantID = newTestUUID(t)
			}

			if tc.wrongUse {
				encryptionKey.Use = cryptoutilAppsSmKmsModel.KeyUseSig
			}

			if tc.publicJWKJWE != "" {
				testEncryptionMaterial.PublicJWKJWE = tc.publicJWKJWE
			}

			if tc.algorithm != "" {
				encryptionKey.Algorithm = tc.algorithm
			}

			elasticRepo := &elasticJWKRepositoryStub{
				getByIDFn: func(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
					if tc.getError && id == encryptionKey.ID {
						return nil, errJWKServiceInjected
					}

					if id == signingKey.ID && tc.signingError {
						return nil, errJWKServiceInjected
					}

					if id == signingKey.ID {
						return &signingKey, nil
					}

					return &encryptionKey, nil
				},
			}
			materialRepo := &materialJWKRepositoryStub{
				getActiveFn: func(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					if id == signingKey.ID {
						return &testSigningMaterial, nil
					}

					if tc.materialError {
						return nil, errJWKServiceInjected
					}

					return &testEncryptionMaterial, nil
				},
			}
			_, err := NewJWTService(elasticRepo, materialRepo, testBarrierService).CreateEncryptedJWT(
				testContext,
				tenantID,
				signingKey.ID,
				encryptionKey.ID,
				&JWTClaims{Subject: testJWTSubject},
			)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func TestJWTService_ParseClaimsMapValueTypes(t *testing.T) {
	t.Parallel()

	service := &jwtServiceImpl{}
	claims := service.parseClaimsMap(map[string]any{
		cryptoutilSharedMagic.ClaimIss: 42,
		cryptoutilSharedMagic.ClaimSub: testJWTSubject,
		cryptoutilSharedMagic.ClaimAud: []any{testJWTAudience, 42},
		cryptoutilSharedMagic.ClaimExp: json.Number(cryptoutilSharedMagic.DevSetupTestInvalidName),
		cryptoutilSharedMagic.ClaimNbf: float64(1_700_000_000),
		cryptoutilSharedMagic.ClaimIat: json.Number("1"),
		cryptoutilSharedMagic.ClaimJti: "token-id",
		"custom":                       true,
	})

	require.Empty(t, claims.Issuer)
	require.Equal(t, testJWTSubject, claims.Subject)
	require.Equal(t, []string{testJWTAudience}, claims.Audience)
	require.Equal(t, time.Unix(0, 0), *claims.ExpiresAt)
	require.Equal(t, time.Unix(1_700_000_000, 0), *claims.NotBefore)
	require.Equal(t, time.Unix(1, 0), *claims.IssuedAt)
	require.Equal(t, "token-id", claims.JTI)
	require.Equal(t, true, claims.Custom["custom"])

	claims = service.parseClaimsMap(map[string]any{
		cryptoutilSharedMagic.ClaimAud: []string{"unsupported"},
		cryptoutilSharedMagic.ClaimExp: float64(1_700_000_001),
		cryptoutilSharedMagic.ClaimNbf: json.Number(cryptoutilSharedMagic.DevSetupTestInvalidName),
		cryptoutilSharedMagic.ClaimIat: cryptoutilSharedMagic.DevSetupTestInvalidName,
		cryptoutilSharedMagic.ClaimJti: 42,
	})
	require.Empty(t, claims.Audience)
	require.Equal(t, time.Unix(1_700_000_001, 0), *claims.ExpiresAt)
	require.Equal(t, time.Unix(0, 0), *claims.NotBefore)
	require.Nil(t, claims.IssuedAt)
	require.Empty(t, claims.JTI)
}
