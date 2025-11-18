//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package handler

import (
	"errors"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"

	googleUuid "github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

// Test constants for error messages and test data.
const (
	testKeyNotFound = "key not found"
	testContext     = "test-context"
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
	summary := "invalid request"
	appErr := cryptoutilAppErr.NewHTTP400BadRequest(&summary, nil)

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
	summary := "resource not found"
	appErr := cryptoutilAppErr.NewHTTP404NotFound(&summary, nil)

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
	summary := "internal error"
	appErr := cryptoutilAppErr.NewHTTP500InternalServerError(&summary, nil)

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
	summary := "invalid ciphertext"
	appErr := cryptoutilAppErr.NewHTTP400BadRequest(&summary, nil)

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
	appErr := cryptoutilAppErr.NewHTTP404NotFound(&summary, nil)

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
	summary := "decryption failed"
	appErr := cryptoutilAppErr.NewHTTP500InternalServerError(&summary, nil)

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
	summary := "invalid plaintext"
	appErr := cryptoutilAppErr.NewHTTP400BadRequest(&summary, nil)

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
	appErr := cryptoutilAppErr.NewHTTP404NotFound(&summary, nil)

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
	summary := "encryption failed"
	appErr := cryptoutilAppErr.NewHTTP500InternalServerError(&summary, nil)

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
	summary := "invalid generate params"
	appErr := cryptoutilAppErr.NewHTTP400BadRequest(&summary, nil)

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
	summary := "elastic key not found"
	appErr := cryptoutilAppErr.NewHTTP404NotFound(&summary, nil)

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
	summary := "generation failed"
	appErr := cryptoutilAppErr.NewHTTP500InternalServerError(&summary, nil)

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
	mockService := &cryptoutilBusinessLogic.BusinessLogicService{}
	server := NewOpenapiStrictServer(mockService)

	// Verify server is a valid implementation
	var _ cryptoutilOpenapiServer.StrictServerInterface = server

	require.NotNil(t, server)
}
