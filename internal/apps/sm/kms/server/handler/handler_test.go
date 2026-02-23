// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package handler

import (
	"errors"
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
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
	elasticKey := &cryptoutilKmsServer.ElasticKey{
		ElasticKeyID: &uuid,
	}

	resp, err := mapper.toOasPostKeyResponse(nil, elasticKey)
	require.NoError(t, err)
	require.NotNil(t, resp)

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickey200JSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickey400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.BadRequestJSONResponse)
}

// TestOamOasMapper_ToOasPostKeyResponse_NotFound tests that 404 falls through to error (PostElastickey has no 404 response).
func TestOamOasMapper_ToOasPostKeyResponse_NotFound(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	summary := testResourceNotFnd
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)

	resp, err := mapper.toOasPostKeyResponse(appErr, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to add ElasticKey")
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickey500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.InternalServerErrorJSONResponse)
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

	textResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt200TextResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.BadRequestJSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.NotFoundJSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.InternalServerErrorJSONResponse)
}

// TestOamOasMapper_ToOasPostEncryptResponse_Success tests successful encrypt response.
func TestOamOasMapper_ToOasPostEncryptResponse_Success(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	encryptedData := []byte("encrypted ciphertext")

	resp, err := mapper.toOasPostEncryptResponse(nil, encryptedData)
	require.NoError(t, err)
	require.NotNil(t, resp)

	textResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt200TextResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.BadRequestJSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.NotFoundJSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.InternalServerErrorJSONResponse)
}

// TestOamOasMapper_ToOamPostGenerateQueryParams tests parameter mapping for generate endpoint.
func TestOamOasMapper_ToOamPostGenerateQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	context := testContext
	alg := "RSA-OAEP"
	openapiParams := &cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerateParams{
		Context: &context,
		Alg:     &alg,
	}

	generateParams := mapper.toOamPostGenerateQueryParams(openapiParams)
	require.NotNil(t, generateParams)
	require.NotNil(t, generateParams.Context)
	require.NotNil(t, generateParams.Alg)
}

// TestOamOasMapper_ToOamPostEncryptQueryParams tests parameter mapping for encrypt endpoint.
func TestOamOasMapper_ToOamPostEncryptQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()
	context := testContext
	openapiParams := &cryptoutilKmsServer.PostElastickeyElasticKeyIDEncryptParams{
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

	textResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate200TextResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate400JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.BadRequestJSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate404JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.NotFoundJSONResponse)
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

	jsonResp, ok := resp.(cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate500JSONResponse)
	require.True(t, ok)
	require.NotNil(t, jsonResp.InternalServerErrorJSONResponse)
}

// TestStrictServer_HandlerMethodsExist verifies that all handler methods are implemented.
