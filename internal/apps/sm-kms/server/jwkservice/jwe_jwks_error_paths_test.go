// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	jose "github.com/go-jose/go-jose/v4"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWEService_EncryptFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilAppsSmKmsModel.KeyUseEnc, 2)
	require.NoError(t, err)

	tests := []struct {
		name          string
		getError      bool
		wrongTenant   bool
		wrongUse      bool
		activeError   bool
		algorithm     string
		publicJWKJWE  string
		cancelContext bool
		wantError     string
	}{
		{name: "elastic repository failure", getError: true, wantError: testErrorFailedGetElastic},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: testErrorNotFound},
		{name: "signing key use", wrongUse: true, wantError: testErrorNotConfiguredEncrypt},
		{name: "active material failure", activeError: true, wantError: testErrorFailedGetActive},
		{name: "invalid base64 material", publicJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePublic},
		{name: "barrier decrypt failure", publicJWKJWE: base64.StdEncoding.EncodeToString([]byte("not encrypted")), wantError: testErrorFailedDecryptPublic},
		{name: "invalid public JWK JSON", publicJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePublic},
		{name: "unsupported JWE algorithm", algorithm: testErrorUnsupported, wantError: testErrorUnsupportedAlgorithmForJWE},
		{name: "algorithm and key mismatch", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA2048, wantError: testErrorFailedCreateEncrypter},
		{name: "canceled barrier context", cancelContext: true, wantError: testErrorFailedDecryptPublic},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			elastic := *elasticJWK
			activeMaterial := *material

			if tc.wrongTenant {
				elastic.TenantID = newTestUUID(t)
			}

			if tc.wrongUse {
				elastic.Use = cryptoutilAppsSmKmsModel.KeyUseSig
			}

			if tc.algorithm != "" {
				elastic.Algorithm = tc.algorithm
			}

			if tc.publicJWKJWE != "" {
				activeMaterial.PublicJWKJWE = tc.publicJWKJWE
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

					return &activeMaterial, nil
				},
			}

			ctx := testContext

			if tc.cancelContext {
				var cancel context.CancelFunc

				ctx, cancel = context.WithCancel(testContext)
				cancel()
			}

			_, err := NewJWEService(elasticRepo, materialRepo, testBarrierService).Encrypt(ctx, tenantID, elastic.ID, []byte("payload"))
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestJWEService_EncryptWithKIDFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilAppsSmKmsModel.KeyUseEnc, 2)
	require.NoError(t, err)

	tests := []struct {
		name          string
		getError      bool
		wrongTenant   bool
		wrongUse      bool
		materialError bool
		foreign       bool
		publicJWKJWE  string
		algorithm     string
		wantError     string
	}{
		{name: testErrorCaseElasticLookup, getError: true, wantError: testErrorFailedGetElastic},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: testErrorNotFound},
		{name: testErrorCaseWrongKeyUse, wrongUse: true, wantError: testErrorNotConfiguredEncrypt},
		{name: testErrorCaseMaterialLookup, materialError: true, wantError: testErrorFailedGetMaterialKID},
		{name: testErrorCaseForeignMaterial, foreign: true, wantError: testErrorMaterialNotBelong},
		{name: "invalid base64", publicJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePublic},
		{name: "barrier decrypt error", publicJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad envelope")), wantError: testErrorFailedDecryptPublic},
		{name: "JWK JSON error", publicJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePublic},
		{name: testErrorCaseUnsupportedAlg, algorithm: testErrorUnsupported, wantError: testErrorUnsupportedAlgorithmForJWE},
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
				elastic.Use = cryptoutilAppsSmKmsModel.KeyUseSig
			}

			if tc.foreign {
				testMaterial.ElasticJWKID = newTestUUID(t)
			}

			if tc.publicJWKJWE != "" {
				testMaterial.PublicJWKJWE = tc.publicJWKJWE
			}

			if tc.algorithm != "" {
				elastic.Algorithm = tc.algorithm
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

			materialRepo := &materialJWKRepositoryStub{}
			if tc.materialError {
				materialRepo.getByKIDFn = func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					return nil, errJWKServiceInjected
				}
			} else {
				materialRepo.getByKIDFn = func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
					return &testMaterial, nil
				}
			}

			_, err := NewJWEService(elasticRepo, materialRepo, testBarrierService).EncryptWithKID(testContext, tenantID, elastic.ID, material.MaterialKID, []byte("payload"))
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestJWEService_DecryptFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, material, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilAppsSmKmsModel.KeyUseEnc, 2)
	require.NoError(t, err)

	elasticRepo, materialRepo := testRepositories()
	service := NewJWEService(elasticRepo, materialRepo, testBarrierService)
	jweImpl, ok := service.(*jweServiceImpl)
	require.True(t, ok)

	compact, err := service.Encrypt(testContext, tenantID, elasticJWK.ID, []byte("decrypt failure tests"))
	require.NoError(t, err)

	_, err = service.Decrypt(testContext, tenantID, newTestUUID(t), compact)
	require.Error(t, err)
	_, err = service.Decrypt(testContext, tenantID, elasticJWK.ID, "malformed")
	require.ErrorContains(t, err, "failed to parse JWE")

	stubElasticRepo := &elasticJWKRepositoryStub{
		getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
			return elasticJWK, nil
		},
	}
	stubMaterialRepo := &materialJWKRepositoryStub{
		listByElasticJWKFn: func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
			return nil, 0, errJWKServiceInjected
		},
	}
	_, err = NewJWEService(stubElasticRepo, stubMaterialRepo, testBarrierService).Decrypt(testContext, tenantID, elasticJWK.ID, compact)
	require.ErrorContains(t, err, "failed to list materials")

	stubMaterialRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
		return []*cryptoutilAppsSmKmsModel.MaterialJWK{{PrivateJWKJWE: testErrorInvalidJWKEnvelope}}, 1, nil
	}
	_, err = NewJWEService(stubElasticRepo, stubMaterialRepo, testBarrierService).Decrypt(testContext, tenantID, elasticJWK.ID, compact)
	require.ErrorContains(t, err, "no matching key")

	for _, tc := range []struct {
		name          string
		privateJWKJWE string
		wantError     string
	}{
		{name: "decode failure", privateJWKJWE: testErrorInvalidJWKEnvelope, wantError: testErrorFailedDecodePrivate},
		{name: "decrypt failure", privateJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad envelope")), wantError: testErrorFailedDecryptPrivate},
		{name: "parse failure", privateJWKJWE: encryptTestValue(t, []byte("not-json")), wantError: testErrorFailedParsePrivate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			malformedMaterial := *material
			malformedMaterial.PrivateJWKJWE = tc.privateJWKJWE
			jweObject, parseErr := jose.ParseEncrypted(compact, []jose.KeyAlgorithm{jose.ECDH_ES_A256KW}, []jose.ContentEncryption{jose.A256GCM})
			require.NoError(t, parseErr)

			_, decryptErr := jweImpl.decryptWithMaterial(testContext, jweObject, &malformedMaterial)
			require.Error(t, decryptErr)
			require.ErrorContains(t, decryptErr, tc.wantError)
		})
	}
}

func TestJWKSService_RepositoryAndSkippedMaterialPaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, validMaterial, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 3)
	require.NoError(t, err)

	elasticRepo := &elasticJWKRepositoryStub{}
	materialRepo := &materialJWKRepositoryStub{}
	service := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	elasticRepo.listFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
		return nil, 0, errJWKServiceInjected
	}
	_, err = service.GetJWKS(testContext, tenantID)
	require.ErrorContains(t, err, "failed to list elastic JWKs")

	noMaterial := &cryptoutilAppsSmKmsModel.ElasticJWK{ID: newTestUUID(t), TenantID: tenantID}
	badBase64Key := &cryptoutilAppsSmKmsModel.ElasticJWK{ID: newTestUUID(t), TenantID: tenantID}
	badEnvelopeKey := &cryptoutilAppsSmKmsModel.ElasticJWK{ID: newTestUUID(t), TenantID: tenantID}
	badJSONKey := &cryptoutilAppsSmKmsModel.ElasticJWK{ID: newTestUUID(t), TenantID: tenantID}
	validKey := *elasticJWK
	elasticRepo.listFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
		return []*cryptoutilAppsSmKmsModel.ElasticJWK{noMaterial, badBase64Key, badEnvelopeKey, badJSONKey, &validKey}, 5, nil
	}
	materialRepo.getActiveFn = func(_ context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
		switch id {
		case noMaterial.ID:
			return nil, errJWKServiceInjected
		case badBase64Key.ID:
			return &cryptoutilAppsSmKmsModel.MaterialJWK{PublicJWKJWE: testErrorInvalidJWKEnvelope}, nil
		case badEnvelopeKey.ID:
			return &cryptoutilAppsSmKmsModel.MaterialJWK{PublicJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad"))}, nil
		case badJSONKey.ID:
			return &cryptoutilAppsSmKmsModel.MaterialJWK{PublicJWKJWE: encryptTestValue(t, []byte("not-json"))}, nil
		default:
			return validMaterial, nil
		}
	}
	jwks, err := service.GetJWKS(testContext, tenantID)
	require.NoError(t, err)
	require.Len(t, jwks.Keys, 1)

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return nil, errJWKServiceInjected
	}
	_, err = service.GetJWKSForElasticKey(testContext, tenantID, elasticJWK.ID)
	require.ErrorContains(t, err, testErrorFailedGetElastic)

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		foreign := *elasticJWK
		foreign.TenantID = newTestUUID(t)

		return &foreign, nil
	}
	_, err = service.GetJWKSForElasticKey(testContext, tenantID, elasticJWK.ID)
	require.ErrorContains(t, err, testErrorNotFound)

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return elasticJWK, nil
	}
	materialRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
		return nil, 0, errJWKServiceInjected
	}
	_, err = service.GetJWKSForElasticKey(testContext, tenantID, elasticJWK.ID)
	require.ErrorContains(t, err, "failed to list materials")

	retiredAt := time.Now().UTC()
	materialRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
		retired := *validMaterial
		retired.RetiredAt = &retiredAt

		return []*cryptoutilAppsSmKmsModel.MaterialJWK{
			&retired,
			{PublicJWKJWE: testErrorInvalidJWKEnvelope},
			{PublicJWKJWE: base64.StdEncoding.EncodeToString([]byte("bad"))},
			{PublicJWKJWE: encryptTestValue(t, []byte("not-json"))},
			validMaterial,
		}, 5, nil
	}
	keySet, err := service.GetJWKSForElasticKey(testContext, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, keySet.Keys, 1)

	materialRepo.getByKIDFn = func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
		return nil, errJWKServiceInjected
	}
	_, err = service.GetPublicJWK(testContext, tenantID, validMaterial.MaterialKID)
	require.ErrorContains(t, err, testErrorNotFound)

	materialRepo.getByKIDFn = func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
		return validMaterial, nil
	}
	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return nil, errJWKServiceInjected
	}
	_, err = service.GetPublicJWK(testContext, tenantID, validMaterial.MaterialKID)
	require.ErrorContains(t, err, testErrorFailedGetElastic)

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return elasticJWK, nil
	}
	validMaterialCopy := *validMaterial
	validMaterialCopy.PublicJWKJWE = testErrorInvalidJWKEnvelope
	materialRepo.getByKIDFn = func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
		return &validMaterialCopy, nil
	}
	_, err = service.GetPublicJWK(testContext, tenantID, validMaterial.MaterialKID)
	require.ErrorContains(t, err, testErrorFailedDecodePublic)

	validMaterialCopy.PublicJWKJWE = base64.StdEncoding.EncodeToString([]byte("bad"))
	_, err = service.GetPublicJWK(testContext, tenantID, validMaterial.MaterialKID)
	require.ErrorContains(t, err, testErrorFailedDecryptPublic)
	validMaterialCopy.PublicJWKJWE = encryptTestValue(t, []byte("not-json"))
	_, err = service.GetPublicJWK(testContext, tenantID, validMaterial.MaterialKID)
	require.ErrorContains(t, err, testErrorFailedParsePublic)
}
