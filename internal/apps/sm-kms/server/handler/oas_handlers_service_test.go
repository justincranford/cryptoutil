// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	"context"
	"errors"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	googleUuid "github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

const (
	testStrictServerPayload      = "strict-server payload"
	testStrictServerSuccess      = "strict server returns a response"
	testStrictServerServiceError = "injected business logic failure"
)

type strictServerBusinessLogicStub struct {
	err         error
	elasticKey  *cryptoutilKmsServer.ElasticKey
	materialKey *cryptoutilKmsServer.MaterialKey
}

func newStrictServerBusinessLogicStub(t *testing.T, serviceErr error) *strictServerBusinessLogicStub {
	t.Helper()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	apiID := openapiTypes.UUID(id)

	return &strictServerBusinessLogicStub{
		err: serviceErr,
		elasticKey: &cryptoutilKmsServer.ElasticKey{
			ElasticKeyID: &apiID,
		},
		materialKey: &cryptoutilKmsServer.MaterialKey{
			MaterialKeyID: &apiID,
			ElasticKeyID:  &apiID,
		},
	}
}

func (s *strictServerBusinessLogicStub) AddElasticKey(context.Context, *cryptoutilKmsServer.ElasticKeyCreate) (*cryptoutilKmsServer.ElasticKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.elasticKey, nil
}

func (s *strictServerBusinessLogicStub) GetElasticKeyByElasticKeyID(context.Context, *googleUuid.UUID) (*cryptoutilKmsServer.ElasticKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.elasticKey, nil
}

func (s *strictServerBusinessLogicStub) PostDecryptByElasticKeyID(context.Context, *googleUuid.UUID, []byte) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []byte(testStrictServerPayload), nil
}

func (s *strictServerBusinessLogicStub) PostEncryptByElasticKeyID(context.Context, *googleUuid.UUID, *cryptoutilOpenapiModel.EncryptParams, []byte) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []byte(testStrictServerPayload), nil
}

func (s *strictServerBusinessLogicStub) PostGenerateByElasticKeyID(context.Context, *googleUuid.UUID, *cryptoutilOpenapiModel.GenerateParams) ([]byte, []byte, []byte, error) {
	if s.err != nil {
		return nil, nil, nil, s.err
	}

	payload := []byte(testStrictServerPayload)

	return payload, payload, payload, nil
}

func (s *strictServerBusinessLogicStub) GenerateMaterialKeyInElasticKey(context.Context, *googleUuid.UUID, *cryptoutilKmsServer.MaterialKeyGenerate) (*cryptoutilKmsServer.MaterialKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.materialKey, nil
}

func (s *strictServerBusinessLogicStub) GetMaterialKeyByElasticKeyAndMaterialKeyID(context.Context, *googleUuid.UUID, *googleUuid.UUID) (*cryptoutilKmsServer.MaterialKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.materialKey, nil
}

func (s *strictServerBusinessLogicStub) GetMaterialKeysForElasticKey(context.Context, *googleUuid.UUID, *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams) ([]cryptoutilKmsServer.MaterialKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []cryptoutilKmsServer.MaterialKey{*s.materialKey}, nil
}

func (s *strictServerBusinessLogicStub) PostSignByElasticKeyID(context.Context, *googleUuid.UUID, []byte) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []byte(testStrictServerPayload), nil
}

func (s *strictServerBusinessLogicStub) PostVerifyByElasticKeyID(context.Context, *googleUuid.UUID, []byte) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []byte(testStrictServerPayload), nil
}

func (s *strictServerBusinessLogicStub) GetElasticKeys(context.Context, *cryptoutilOpenapiModel.ElasticKeysQueryParams) ([]cryptoutilKmsServer.ElasticKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []cryptoutilKmsServer.ElasticKey{*s.elasticKey}, nil
}

func (s *strictServerBusinessLogicStub) GetMaterialKeys(context.Context, *cryptoutilOpenapiModel.MaterialKeysQueryParams) ([]cryptoutilKmsServer.MaterialKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return []cryptoutilKmsServer.MaterialKey{*s.materialKey}, nil
}

func (s *strictServerBusinessLogicStub) UpdateElasticKey(context.Context, *googleUuid.UUID, *cryptoutilKmsServer.ElasticKeyUpdate) (*cryptoutilKmsServer.ElasticKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.elasticKey, nil
}

func (s *strictServerBusinessLogicStub) DeleteElasticKey(context.Context, *googleUuid.UUID) error {
	return s.err
}

func (s *strictServerBusinessLogicStub) ImportMaterialKey(context.Context, *googleUuid.UUID, *cryptoutilKmsServer.MaterialKeyImport) (*cryptoutilKmsServer.MaterialKey, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.materialKey, nil
}

func (s *strictServerBusinessLogicStub) RevokeMaterialKey(context.Context, *googleUuid.UUID, *googleUuid.UUID) error {
	return s.err
}

func (s *strictServerBusinessLogicStub) DeleteMaterialKey(context.Context, *googleUuid.UUID, *googleUuid.UUID) error {
	return s.err
}

func TestStrictServer_BusinessLogicResponseMappings(t *testing.T) {
	t.Parallel()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	calls := []struct {
		name   string
		invoke func(*testing.T, *StrictServer, googleUuid.UUID) (bool, error)
	}{
		{
			name: "create elastic key",
			invoke: func(_ *testing.T, server *StrictServer, _ googleUuid.UUID) (bool, error) {
				response, err := server.PostElasticKeys(context.Background(), cryptoutilKmsServer.PostElasticKeysRequestObject{})

				return response != nil, err
			},
		},
		{
			name: "get elastic key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.GetElasticKeysElasticKeyID(context.Background(), cryptoutilKmsServer.GetElasticKeysElasticKeyIDRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "decrypt",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				body := cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecryptTextRequestBody(testStrictServerPayload)
				response, err := server.PostElasticKeysElasticKeyIDDecrypt(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecryptRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), Body: &body})

				return response != nil, err
			},
		},
		{
			name: "encrypt",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				body := cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncryptTextRequestBody(testStrictServerPayload)
				response, err := server.PostElasticKeysElasticKeyIDEncrypt(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncryptRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), Body: &body})

				return response != nil, err
			},
		},
		{
			name: "generate",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.PostElasticKeysElasticKeyIDGenerate(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerateRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "generate material key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.PostElasticKeysElasticKeyIDMaterialKeys(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "get material key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyID(context.Background(), cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), MaterialKeyID: cryptoutilKmsServer.MaterialKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "list elastic key material keys",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.GetElasticKeysElasticKeyIDMaterialKeys(context.Background(), cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "sign",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				body := cryptoutilKmsServer.PostElasticKeysElasticKeyIDSignTextRequestBody(testStrictServerPayload)
				response, err := server.PostElasticKeysElasticKeyIDSign(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDSignRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), Body: &body})

				return response != nil, err
			},
		},
		{
			name: "verify",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				body := cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerifyTextRequestBody(testStrictServerPayload)
				response, err := server.PostElasticKeysElasticKeyIDVerify(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerifyRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), Body: &body})

				return response != nil, err
			},
		},
		{
			name: "list elastic keys",
			invoke: func(_ *testing.T, server *StrictServer, _ googleUuid.UUID) (bool, error) {
				response, err := server.GetElasticKeys(context.Background(), cryptoutilKmsServer.GetElasticKeysRequestObject{})

				return response != nil, err
			},
		},
		{
			name: "list material keys",
			invoke: func(_ *testing.T, server *StrictServer, _ googleUuid.UUID) (bool, error) {
				response, err := server.GetMaterialKeys(context.Background(), cryptoutilKmsServer.GetMaterialKeysRequestObject{})

				return response != nil, err
			},
		},
		{
			name: "update elastic key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.PutElasticKeysElasticKeyID(context.Background(), cryptoutilKmsServer.PutElasticKeysElasticKeyIDRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "delete elastic key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.DeleteElasticKeysElasticKeyID(context.Background(), cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "import material key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.PostElasticKeysElasticKeyIDImport(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDImportRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "revoke material key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevoke(context.Background(), cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevokeRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), MaterialKeyID: cryptoutilKmsServer.MaterialKeyID(id)})

				return response != nil, err
			},
		},
		{
			name: "delete material key",
			invoke: func(_ *testing.T, server *StrictServer, id googleUuid.UUID) (bool, error) {
				response, err := server.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyID(context.Background(), cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRequestObject{ElasticKeyID: cryptoutilKmsServer.ElasticKeyID(id), MaterialKeyID: cryptoutilKmsServer.MaterialKeyID(id)})

				return response != nil, err
			},
		},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, path := range []struct {
				name string
				err  error
			}{
				{name: testStrictServerSuccess},
				{name: "business_logic_error", err: errors.New(testStrictServerServiceError)},
			} {
				t.Run(path.name, func(t *testing.T) {
					t.Parallel()

					service := newStrictServerBusinessLogicStub(t, path.err)
					handler := NewOpenapiStrictServer(service)
					hasResponse, err := tc.invoke(t, handler, id)

					if path.err != nil {
						require.ErrorContains(t, err, testStrictServerServiceError)
						require.False(t, hasResponse)
					} else {
						require.NoError(t, err)
						require.True(t, hasResponse)
					}
				})
			}
		})
	}
}
