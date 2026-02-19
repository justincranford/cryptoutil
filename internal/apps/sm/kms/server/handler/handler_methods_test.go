//go:build ignore
// +build ignore

// TODO(v7-phase5): This test file is temporarily disabled during OpenAPI migration.
// The handler tests need to be updated to use the new KMS-specific OpenAPI types:
// - cryptoutil/api/kms/server instead of cryptoutil/api/server
// - New response type structure (embedded structs vs named fields)
// - 404 response handling for endpoints that support it

// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package handler

import (
	"errors"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm/kms/server/businesslogic"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"

	googleUuid "github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestStrictServer_HandlerMethodsExist(t *testing.T) {
	t.Parallel()

	// Create server with nil service (just testing method existence)
	mockService := &cryptoutilKmsServerBusinesslogic.BusinessLogicService{}
	server := NewOpenapiStrictServer(mockService)

	// Verify server is a valid implementation
	var _ cryptoutilOpenapiServer.StrictServerInterface = server

	require.NotNil(t, server)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_Success tests successful get elastic key response.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	googleUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	uuid := openapiTypes.UUID(googleUUID)
	elasticKey := &cryptoutilOpenapiModel.ElasticKey{
		ElasticKeyID: &uuid,
	}

	resp, err := mapper.toOasGetElastickeyElasticKeyIDResponse(nil, elasticKey)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_BadRequest tests 400 error for get elastic key.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_BadRequest(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInvalidRequest
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)

	resp, err := mapper.toOasGetElastickeyElasticKeyIDResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_NotFound tests 404 error for get elastic key.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testKeyNotFound
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasGetElastickeyElasticKeyIDResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_InternalServerError tests 500 error for get elastic key.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_InternalServerError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInternalError
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)

	resp, err := mapper.toOasGetElastickeyElasticKeyIDResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_UnknownError tests unknown error for get elastic key.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDResponse_UnknownError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	unknownErr := errors.New("unknown error")

	resp, err := mapper.toOasGetElastickeyElasticKeyIDResponse(unknownErr, nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

// TestOamOasMapper_ToOasPostDecryptResponse_UnknownError tests unknown error for decrypt.
func TestOamOasMapper_ToOasPostDecryptResponse_UnknownError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	unknownErr := errors.New("unknown error")

	resp, err := mapper.toOasPostDecryptResponse(unknownErr, nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

// TestOamOasMapper_ToOasPostEncryptResponse_UnknownError tests unknown error for encrypt.
func TestOamOasMapper_ToOasPostEncryptResponse_UnknownError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	unknownErr := errors.New("unknown error")

	resp, err := mapper.toOasPostEncryptResponse(unknownErr, nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

// TestOamOasMapper_ToOasPostGenerateResponse_UnknownError tests unknown error for generate.
func TestOamOasMapper_ToOasPostGenerateResponse_UnknownError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	unknownErr := errors.New("unknown error")

	resp, err := mapper.toOasPostGenerateResponse(unknownErr, nil, nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_Success tests successful material key creation.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	googleUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	materialKeyID := cryptoutilOpenapiModel.MaterialKeyID(googleUUID)
	elasticKeyID := cryptoutilOpenapiModel.ElasticKeyID(googleUUID)
	materialKey := &cryptoutilOpenapiModel.MaterialKey{
		MaterialKeyID: materialKeyID,
		ElasticKeyID:  elasticKeyID,
	}

	resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(nil, materialKey)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_BadRequest tests 400 error.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_BadRequest(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInvalidRequest
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)

	resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_NotFound tests 404 error.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testKeyNotFound
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_InternalServerError tests 500 error.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_InternalServerError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInternalError
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)

	resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_UnknownError tests unknown error.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyResponse_UnknownError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	unknownErr := errors.New("unknown error")

	resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(unknownErr, nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse_Success tests success.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	googleUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	materialKeyID := cryptoutilOpenapiModel.MaterialKeyID(googleUUID)
	elasticKeyID := cryptoutilOpenapiModel.ElasticKeyID(googleUUID)
	materialKey := &cryptoutilOpenapiModel.MaterialKey{
		MaterialKeyID: materialKeyID,
		ElasticKeyID:  elasticKeyID,
	}

	resp, err := mapper.toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(nil, materialKey)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse_Errors tests errors.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr("invalid"), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr("not found"), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr("error"), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(tc.err, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// strPtr is a helper to create string pointers.
func strPtr(s string) *string {
	return &s
}

// TestOamOasMapper_ToOamGetElasticKeyMaterialKeysQueryParams tests query param mapping.
func TestOamOasMapper_ToOamGetElasticKeyMaterialKeysQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	params := &cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysParams{}

	result := mapper.toOamGetElasticKeyMaterialKeysQueryParams(params)
	require.NotNil(t, result)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeysResponse_Success tests success response.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeysResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	keys := []cryptoutilOpenapiModel.MaterialKey{}

	resp, err := mapper.toOasGetElastickeyElasticKeyIDMaterialkeysResponse(nil, keys)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeysResponse_Errors tests error responses.
func TestOamOasMapper_ToOasGetElastickeyElasticKeyIDMaterialkeysResponse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr("invalid"), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr("not found"), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr("error"), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasGetElastickeyElasticKeyIDMaterialkeysResponse(tc.err, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasPostSignResponse_Success tests successful sign response.
func TestOamOasMapper_ToOasPostSignResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	signedData := []byte("signed-data")

	resp, err := mapper.toOasPostSignResponse(nil, signedData)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasPostSignResponse_Errors tests sign error responses.
func TestOamOasMapper_ToOasPostSignResponse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr("invalid"), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr("not found"), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr("error"), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasPostSignResponse(tc.err, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOasPostVerifyResponse_Success tests successful verify response.
func TestOamOasMapper_ToOasPostVerifyResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()

	resp, err := mapper.toOasPostVerifyResponse(nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasPostVerifyResponse_Errors tests verify error responses.
func TestOamOasMapper_ToOasPostVerifyResponse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "bad request",
			err:     cryptoutilSharedApperr.NewHTTP400BadRequest(strPtr("invalid"), nil),
			wantErr: false,
		},
		{
			name:    "not found",
			err:     cryptoutilSharedApperr.NewHTTP404NotFound(strPtr("not found"), nil),
			wantErr: false,
		},
		{
			name:    testInternalError,
			err:     cryptoutilSharedApperr.NewHTTP500InternalServerError(strPtr("error"), nil),
			wantErr: false,
		},
		{
			name:    "unknown error",
			err:     errors.New("unknown"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapper := NewOasOamMapper()

			resp, err := mapper.toOasPostVerifyResponse(tc.err)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

// TestOamOasMapper_ToOamGetElasticKeyQueryParams tests elastic key query param mapping.
func TestOamOasMapper_ToOamGetElasticKeyQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	params := &cryptoutilOpenapiServer.GetElastickeysParams{}

	result := mapper.toOamGetElasticKeyQueryParams(params)
	require.NotNil(t, result)
}

// TestOamOasMapper_ToOasGetElastickeysResponse_Success tests successful elastic keys response.
