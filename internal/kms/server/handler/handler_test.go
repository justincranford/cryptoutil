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
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/kms/server/businesslogic"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"

	googleUuid "github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

// Test constants for error messages and test data.
const (
	testKeyNotFound     = "key not found"
	testContext         = "test-context"
	testInvalidRequest  = "invalid request"
	testInternalError   = "internal error"
	testResourceNotFnd  = "resource not found"
	testInvalidPTText   = "invalid plaintext"
	testInvalidCTText   = "invalid ciphertext"
	testDecryptFailed   = "decryption failed"
	testEncryptFailed   = "encryption failed"
	testGenFailed       = "generation failed"
	testInvalidGenParam = "invalid generate params"
	testEKNotFound      = "elastic key not found"
)

// TestNewOpenapiStrictServer tests that NewOpenapiStrictServer creates a proper server instance.
func TestNewOpenapiStrictServer(t *testing.T) {
	t.Parallel()

	// Create a nil business logic service (this is just testing server construction)
	server := NewOpenapiStrictServer(nil)

	require.NotNil(t, server)
	require.NotNil(t, server.oasOamMapper)
}

// TestNewOasOamMapper tests that NewOasOamMapper creates a mapper instance.
func TestNewOasOamMapper(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	require.NotNil(t, mapper)
}

// TestOamOasMapper_ToOasPostKeyResponse_Success tests successful elastic key creation response.
func TestOamOasMapper_ToOasPostKeyResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	googleUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	uuid := openapiTypes.UUID(googleUUID)
	elasticKey := &cryptoutilOpenapiModel.ElasticKey{
		ElasticKeyID: &uuid,
	}

	resp, err := mapper.toOasPostKeyResponse(nil, elasticKey)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickey200JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.ElasticKeyID)
}

// TestOamOasMapper_ToOasPostKeyResponse_BadRequest tests 400 error response.
func TestOamOasMapper_ToOasPostKeyResponse_BadRequest(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInvalidRequest
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)

	resp, err := mapper.toOasPostKeyResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickey400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP400BadRequest)
}

// TestOamOasMapper_ToOasPostKeyResponse_NotFound tests 404 error response.
func TestOamOasMapper_ToOasPostKeyResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testResourceNotFnd
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasPostKeyResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickey404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP404NotFound)
}

// TestOamOasMapper_ToOasPostKeyResponse_InternalServerError tests 500 error response.
func TestOamOasMapper_ToOasPostKeyResponse_InternalServerError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInternalError
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)

	resp, err := mapper.toOasPostKeyResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickey500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP500InternalServerError)
}

// TestOamOasMapper_ToOasPostKeyResponse_UnknownError tests handling of unknown errors.
func TestOamOasMapper_ToOasPostKeyResponse_UnknownError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	unknownErr := errors.New("unknown error")

	resp, err := mapper.toOasPostKeyResponse(unknownErr, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to add ElasticKey")
}

// TestOamOasMapper_ToOasPostDecryptResponse_Success tests successful decrypt response.
func TestOamOasMapper_ToOasPostDecryptResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	decryptedData := []byte("decrypted plaintext")

	resp, err := mapper.toOasPostDecryptResponse(nil, decryptedData)
	require.NoError(t, err)
	require.NotNil(t, resp)

	textResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt200TextResponse)
	require.True(t, ok)
	require.Equal(t, decryptedData, []byte(textResp))
}

// TestOamOasMapper_ToOasPostDecryptResponse_BadRequest tests 400 error for decrypt.
func TestOamOasMapper_ToOasPostDecryptResponse_BadRequest(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInvalidCTText
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)

	resp, err := mapper.toOasPostDecryptResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP400BadRequest)
}

// TestOamOasMapper_ToOasPostDecryptResponse_NotFound tests 404 error for decrypt.
func TestOamOasMapper_ToOasPostDecryptResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testKeyNotFound
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasPostDecryptResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP404NotFound)
}

// TestOamOasMapper_ToOasPostDecryptResponse_InternalServerError tests 500 error for decrypt.
func TestOamOasMapper_ToOasPostDecryptResponse_InternalServerError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testDecryptFailed
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)

	resp, err := mapper.toOasPostDecryptResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP500InternalServerError)
}

// TestOamOasMapper_ToOasPostEncryptResponse_Success tests successful encrypt response.
func TestOamOasMapper_ToOasPostEncryptResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	encryptedData := []byte("encrypted ciphertext")

	resp, err := mapper.toOasPostEncryptResponse(nil, encryptedData)
	require.NoError(t, err)
	require.NotNil(t, resp)

	textResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt200TextResponse)
	require.True(t, ok)
	require.Equal(t, encryptedData, []byte(textResp))
}

// TestOamOasMapper_ToOasPostEncryptResponse_BadRequest tests 400 error for encrypt.
func TestOamOasMapper_ToOasPostEncryptResponse_BadRequest(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInvalidPTText
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)

	resp, err := mapper.toOasPostEncryptResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP400BadRequest)
}

// TestOamOasMapper_ToOasPostEncryptResponse_NotFound tests 404 error for encrypt.
func TestOamOasMapper_ToOasPostEncryptResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testKeyNotFound
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasPostEncryptResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP404NotFound)
}

// TestOamOasMapper_ToOasPostEncryptResponse_InternalServerError tests 500 error for encrypt.
func TestOamOasMapper_ToOasPostEncryptResponse_InternalServerError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testEncryptFailed
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)

	resp, err := mapper.toOasPostEncryptResponse(appErr, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP500InternalServerError)
}

// TestOamOasMapper_ToOamPostGenerateQueryParams tests parameter mapping for generate endpoint.
func TestOamOasMapper_ToOamPostGenerateQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	context := testContext
	alg := cryptoutilOpenapiModel.GenerateAlgorithm("RSA-OAEP")
	openapiParams := &cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateParams{
		Context: &context,
		Alg:     &alg,
	}

	generateParams := mapper.toOamPostGenerateQueryParams(openapiParams)
	require.NotNil(t, generateParams)
	require.Equal(t, &context, generateParams.Context)
	require.Equal(t, &alg, generateParams.Alg)
}

// TestOamOasMapper_ToOamPostEncryptQueryParams tests parameter mapping for encrypt endpoint.
func TestOamOasMapper_ToOamPostEncryptQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	context := testContext
	openapiParams := &cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptParams{
		Context: &context,
	}

	encryptParams := mapper.toOamPostEncryptQueryParams(openapiParams)
	require.NotNil(t, encryptParams)
	require.Equal(t, &context, encryptParams.Context)
}

// TestOamOasMapper_ToOasPostGenerateResponse_Success tests successful generate response.
func TestOamOasMapper_ToOasPostGenerateResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	encryptedJWK := []byte("encrypted-jwk-data")
	publicJWK := []byte("public-jwk-data")

	resp, err := mapper.toOasPostGenerateResponse(nil, encryptedJWK, publicJWK)
	require.NoError(t, err)
	require.NotNil(t, resp)

	textResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate200TextResponse)
	require.True(t, ok)
	require.Equal(t, encryptedJWK, []byte(textResp))
}

// TestOamOasMapper_ToOasPostGenerateResponse_BadRequest tests 400 error for generate.
func TestOamOasMapper_ToOasPostGenerateResponse_BadRequest(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testInvalidGenParam
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)

	resp, err := mapper.toOasPostGenerateResponse(appErr, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP400BadRequest)
}

// TestOamOasMapper_ToOasPostGenerateResponse_NotFound tests 404 error for generate.
func TestOamOasMapper_ToOasPostGenerateResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testEKNotFound
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasPostGenerateResponse(appErr, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP404NotFound)
}

// TestOamOasMapper_ToOasPostGenerateResponse_InternalServerError tests 500 error for generate.
func TestOamOasMapper_ToOasPostGenerateResponse_InternalServerError(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testGenFailed
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)

	resp, err := mapper.toOasPostGenerateResponse(appErr, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.HTTP500InternalServerError)
}

// TestStrictServer_HandlerMethodsExist verifies that all handler methods are implemented.
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
func TestOamOasMapper_ToOasGetElastickeysResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	elasticKeys := []cryptoutilOpenapiModel.ElasticKey{}

	resp, err := mapper.toOasGetElastickeysResponse(nil, elasticKeys)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetElastickeysResponse_Errors tests elastic keys error responses.
func TestOamOasMapper_ToOasGetElastickeysResponse_Errors(t *testing.T) {
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

			resp, err := mapper.toOasGetElastickeysResponse(tc.err, nil)
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

// TestOamOasMapper_ToOamGetMaterialKeysQueryParams tests material keys query param mapping.
func TestOamOasMapper_ToOamGetMaterialKeysQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	params := &cryptoutilOpenapiServer.GetMaterialkeysParams{}

	result := mapper.toOamGetMaterialKeysQueryParams(params)
	require.NotNil(t, result)
}

// TestOamOasMapper_ToOasGetMaterialKeysResponse_Success tests successful material keys response.
func TestOamOasMapper_ToOasGetMaterialKeysResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	keys := []cryptoutilOpenapiModel.MaterialKey{}

	resp, err := mapper.toOasGetMaterialKeysResponse(nil, keys)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestOamOasMapper_ToOasGetMaterialKeysResponse_Errors tests material keys error responses.
func TestOamOasMapper_ToOasGetMaterialKeysResponse_Errors(t *testing.T) {
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

			resp, err := mapper.toOasGetMaterialKeysResponse(tc.err, nil)
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

// TestOamOasMapper_ToOasPutElastickeyElasticKeyIDResponse tests PUT ElasticKey responses.
func TestOamOasMapper_ToOasPutElastickeyElasticKeyIDResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name: "bad request",
			err: func() error {
				summary := testInvalidRequest

				return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "not found",
			err: func() error {
				summary := testEKNotFound

				return cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "conflict",
			err: func() error {
				summary := "name already exists"

				return cryptoutilSharedApperr.NewHTTP409Conflict(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "internal error",
			err: func() error {
				summary := testInternalError

				return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)
			}(),
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

			googleUUID, uuidErr := googleUuid.NewV7()
			require.NoError(t, uuidErr)

			uuid := openapiTypes.UUID(googleUUID)
			elasticKey := &cryptoutilOpenapiModel.ElasticKey{
				ElasticKeyID: &uuid,
			}

			resp, err := mapper.toOasPutElastickeyElasticKeyIDResponse(tc.err, elasticKey)
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

// TestOamOasMapper_ToOasDeleteElastickeyElasticKeyIDResponse tests DELETE ElasticKey responses.
func TestOamOasMapper_ToOasDeleteElastickeyElasticKeyIDResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name: "bad request",
			err: func() error {
				summary := testInvalidRequest

				return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "not found",
			err: func() error {
				summary := testEKNotFound

				return cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "internal error",
			err: func() error {
				summary := testInternalError

				return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)
			}(),
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

			resp, err := mapper.toOasDeleteElastickeyElasticKeyIDResponse(tc.err)
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

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDImportResponse tests POST Import MaterialKey responses.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDImportResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name: "bad request",
			err: func() error {
				summary := testInvalidRequest

				return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "not found",
			err: func() error {
				summary := testEKNotFound

				return cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "internal error",
			err: func() error {
				summary := testInternalError

				return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)
			}(),
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

			googleUUID, uuidErr := googleUuid.NewV7()
			require.NoError(t, uuidErr)

			uuid := openapiTypes.UUID(googleUUID)
			materialKey := &cryptoutilOpenapiModel.MaterialKey{
				MaterialKeyID: uuid,
			}

			resp, err := mapper.toOasPostElastickeyElasticKeyIDImportResponse(tc.err, materialKey)
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

// TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse tests POST Revoke MaterialKey responses.
func TestOamOasMapper_ToOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name: "bad request",
			err: func() error {
				summary := testInvalidRequest

				return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "not found",
			err: func() error {
				summary := testEKNotFound

				return cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)
			}(),
			wantErr: false,
		},
		{
			name: "internal error",
			err: func() error {
				summary := testInternalError

				return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)
			}(),
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

			resp, err := mapper.toOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(tc.err)
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
