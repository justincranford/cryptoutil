// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"encoding/base64"
	"testing"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	jose "github.com/go-jose/go-jose/v4"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWSService_SignFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 2)
	require.NoError(t, err)

	tests := []struct {
		name        string
		getError    bool
		wrongTenant bool
		wrongUse    bool
		activeError bool
		wantError   string
	}{
		{name: testErrorCaseElasticRepo, getError: true, wantError: testErrorFailedGetElastic},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: testErrorNotFound},
		{name: "encryption key use", wrongUse: true, wantError: testErrorNotConfiguredSign},
		{name: testErrorCaseActiveMaterial, activeError: true, wantError: testErrorFailedGetActive},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			elastic := *elasticJWK
			if tc.wrongTenant {
				elastic.TenantID = newTestUUID(t)
			}

			if tc.wrongUse {
				elastic.Use = cryptoutilAppsSmKmsModel.KeyUseEnc
			}

			elasticRepo := &elasticJWKRepositoryStub{}
			if tc.getError {
				elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
					return nil, errJWKServiceInjected
				}
			} else {
				elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
					return &elastic, nil
				}
			}

			materialRepo := &materialJWKRepositoryStub{
				getActiveFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					if tc.activeError {
						return nil, errJWKServiceInjected
					}

					return material, nil
				},
			}

			_, err := NewJWSService(elasticRepo, materialRepo, testBarrierService).Sign(testContext, tenantID, elastic.ID, []byte("payload"))
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestJWSService_SignWithKIDFailurePaths(t *testing.T) {
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
		foreign       bool
		wantError     string
	}{
		{name: testErrorCaseElasticLookup, getError: true, wantError: testErrorFailedGetElastic},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: testErrorNotFound},
		{name: testErrorCaseWrongKeyUse, wrongUse: true, wantError: testErrorNotConfiguredSign},
		{name: testErrorCaseMaterialLookup, materialError: true, wantError: testErrorFailedGetMaterialKID},
		{name: testErrorCaseForeignMaterial, foreign: true, wantError: testErrorMaterialNotBelong},
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

			if tc.foreign {
				testMaterial.ElasticJWKID = newTestUUID(t)
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
				getByKIDFn: func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					if tc.materialError {
						return nil, errJWKServiceInjected
					}

					return &testMaterial, nil
				},
			}

			_, err := NewJWSService(elasticRepo, materialRepo, testBarrierService).SignWithKID(
				testContext,
				tenantID,
				elastic.ID,
				material.MaterialKID,
				[]byte("payload"),
			)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestJWSService_VerifyFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 2)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewJWSService(elasticRepo, materialRepo, testBarrierService)
	compact, err := service.Sign(testContext, tenantID, elasticJWK.ID, []byte("verification failure tests"))
	require.NoError(t, err)

	_, err = service.Verify(testContext, tenantID, newTestUUID(t), compact)
	require.Error(t, err)
	_, err = service.Verify(testContext, tenantID, elasticJWK.ID, "malformed")
	require.ErrorContains(t, err, "failed to parse JWS")

	stubElasticRepo := &elasticJWKRepositoryStub{
		getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
			return elasticJWK, nil
		},
	}
	foreignTenantJWK := *elasticJWK
	foreignTenantJWK.TenantID = newTestUUID(t)
	stubElasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return &foreignTenantJWK, nil
	}
	_, err = NewJWSService(stubElasticRepo, materialRepo, testBarrierService).Verify(testContext, tenantID, elasticJWK.ID, compact)
	require.ErrorContains(t, err, testErrorElasticJWKNotFound)

	stubElasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return elasticJWK, nil
	}
	stubMaterialRepo := &materialJWKRepositoryStub{
		listByElasticJWKFn: func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
			return nil, 0, errJWKServiceInjected
		},
	}
	_, err = NewJWSService(stubElasticRepo, stubMaterialRepo, testBarrierService).Verify(testContext, tenantID, elasticJWK.ID, compact)
	require.ErrorContains(t, err, "failed to list materials")

	stubMaterialRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
		return []*cryptoutilAppsSmKmsModel.MaterialJWK{{PublicJWKJWE: testErrorInvalidJWKEnvelope}}, 1, nil
	}
	_, err = NewJWSService(stubElasticRepo, stubMaterialRepo, testBarrierService).Verify(testContext, tenantID, elasticJWK.ID, compact)
	require.ErrorContains(t, err, "no matching key")

	impl, ok := NewJWSService(elasticRepo, materialRepo, testBarrierService).(*jwsServiceImpl)
	require.True(t, ok)

	parsed, err := jose.ParseSigned(compact, []jose.SignatureAlgorithm{jose.ES256})
	require.NoError(t, err)

	for _, tc := range []struct {
		name         string
		publicJWKJWE string
		wantError    string
	}{
		{name: testErrorCaseDecode, publicJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePublic},
		{name: testErrorCaseDecrypt, publicJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad")), wantError: testErrorFailedDecryptPublic},
		{name: testErrorCaseParse, publicJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePublic},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testMaterial := *material
			testMaterial.PublicJWKJWE = tc.publicJWKJWE
			_, verifyErr := impl.verifyWithMaterial(testContext, parsed, &testMaterial)
			require.ErrorContains(t, verifyErr, tc.wantError)
		})
	}

	testMaterial := *material
	testMaterial.PublicJWKJWE = encryptTestValue(t, []byte(`{"kty":"EC"}`))
	_, err = impl.verifyWithMaterial(testContext, parsed, &testMaterial)
	require.Error(t, err)
}

func TestJWSService_SignMaterialFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 1)
	require.NoError(t, err)

	impl, ok := NewJWSService(&elasticJWKRepositoryStub{}, &materialJWKRepositoryStub{}, testBarrierService).(*jwsServiceImpl)
	require.True(t, ok)

	tests := []struct {
		name          string
		privateJWKJWE string
		algorithm     string
		wantError     string
	}{
		{name: testErrorCaseDecode, privateJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePrivate},
		{name: testErrorCaseDecrypt, privateJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad")), wantError: testErrorFailedDecryptPrivate},
		{name: testErrorCaseParse, privateJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePrivate},
		{name: testErrorCaseUnsupportedAlg, algorithm: testErrorUnsupported, wantError: "unsupported algorithm for JWS"},
		{name: "algorithm key mismatch", algorithm: cryptoutilSharedMagic.JoseAlgHS256, wantError: testErrorFailedCreateSigner},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testElastic := *elasticJWK

			testMaterial := *material
			if tc.privateJWKJWE != "" {
				testMaterial.PrivateJWKJWE = tc.privateJWKJWE
			}

			if tc.algorithm != "" {
				testElastic.Algorithm = tc.algorithm
			}

			_, err := impl.signWithMaterial(testContext, &testElastic, &testMaterial, []byte("payload"))
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}
