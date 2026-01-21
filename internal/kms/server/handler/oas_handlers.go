// Copyright (c) 2025 Justin Cranford
//
//

package handler

import (
	"context"

	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilBusinessLogic "cryptoutil/internal/kms/server/businesslogic"
)

// StrictServer implements cryptoutilOpenapiServer.StrictServerInterface.
type StrictServer struct {
	businessLogicService *cryptoutilBusinessLogic.BusinessLogicService
	oasOamMapper         *OamOasMapper
}

// NewOpenapiStrictServer creates a new OpenAPI strict server handler.
func NewOpenapiStrictServer(service *cryptoutilBusinessLogic.BusinessLogicService) *StrictServer {
	return &StrictServer{businessLogicService: service, oasOamMapper: &OamOasMapper{}}
}

// PostElastickey creates a new Elastic Key.
// (POST /elastickey).
func (s *StrictServer) PostElastickey(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
	addedElasticKey, err := s.businessLogicService.AddElasticKey(ctx, request.Body)

	return s.oasOamMapper.toOasPostKeyResponse(err, addedElasticKey)
}

// GetElastickeyElasticKeyID gets an Elastic Key.
// (GET /elastickey/{elasticKeyID}).
func (s *StrictServer) GetElastickeyElasticKeyID(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
	elasticKey, err := s.businessLogicService.GetElasticKeyByElasticKeyID(ctx, &request.ElasticKeyID)

	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDResponse(err, elasticKey)
}

// PostElastickeyElasticKeyIDDecrypt decrypts ciphertext using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/decrypt).
func (s *StrictServer) PostElastickeyElasticKeyIDDecrypt(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	encryptedBytes := []byte(*request.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptedBytes)

	return s.oasOamMapper.toOasPostDecryptResponse(err, decryptedBytes)
}

// PostElastickeyElasticKeyIDEncrypt encrypts clear text data using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/encrypt).
func (s *StrictServer) PostElastickeyElasticKeyIDEncrypt(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	encryptParams := s.oasOamMapper.toOamPostEncryptQueryParams(&request.Params)
	clearBytes := []byte(*request.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptParams, clearBytes)

	return s.oasOamMapper.toOasPostEncryptResponse(err, encryptedBytes)
}

// PostElastickeyElasticKeyIDGenerate generates a random Secret Key, Key Pair, or other algorithm. It will be in JWK format, returned in encrypted form as a JWE message.
// (POST /elastickey/{elasticKeyID}/generate).
func (s *StrictServer) PostElastickeyElasticKeyIDGenerate(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateResponseObject, error) {
	generateParams := s.oasOamMapper.toOamPostGenerateQueryParams(&request.Params)
	encryptedNonPublicJWKBytes, _, clearPublicJWKBytes, err := s.businessLogicService.PostGenerateByElasticKeyID(ctx, &request.ElasticKeyID, generateParams)

	return s.oasOamMapper.toOasPostGenerateResponse(err, encryptedNonPublicJWKBytes, clearPublicJWKBytes)
}

// PostElastickeyElasticKeyIDMaterialkey generates a new Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey).
func (s *StrictServer) PostElastickeyElasticKeyIDMaterialkey(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	key, err := s.businessLogicService.GenerateMaterialKeyInElasticKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(err, key)
}

// GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID gets Material Key in Elastic Key.
// (GET /elastickey/{elasticKeyID}/materialkey/{materialKeyID}).
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	key, err := s.businessLogicService.GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err, key)
}

// GetElastickeyElasticKeyIDMaterialkeys finds Material Keys in Elastic Key. Supports optional filtering, sorting, and paging.
// (GET /elastickey/{elasticKeyID}/materialkeys).
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeys(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	elasticKeyMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeysForElasticKey(ctx, &request.ElasticKeyID, elasticKeyMaterialKeysQueryParams)

	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDMaterialkeysResponse(err, keys)
}

// PostElastickeyElasticKeyIDSign signs clear text using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/sign).
func (s *StrictServer) PostElastickeyElasticKeyIDSign(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	clearBytes := []byte(*request.Body)
	signedBytes, err := s.businessLogicService.PostSignByElasticKeyID(ctx, &request.ElasticKeyID, clearBytes)

	return s.oasOamMapper.toOasPostSignResponse(err, signedBytes)
}

// PostElastickeyElasticKeyIDVerify verifies JWS message using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/verify).
func (s *StrictServer) PostElastickeyElasticKeyIDVerify(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	signedBytes := []byte(*request.Body)
	_, err := s.businessLogicService.PostVerifyByElasticKeyID(ctx, &request.ElasticKeyID, signedBytes)

	return s.oasOamMapper.toOasPostVerifyResponse(err)
}

// GetElastickeys finds Elastic Keys. Supports optional filtering, sorting, and paging.
// (GET /elastickeys).
func (s *StrictServer) GetElastickeys(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeysRequestObject) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
	elasticMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyQueryParams(&request.Params)
	elasticKeys, err := s.businessLogicService.GetElasticKeys(ctx, elasticMaterialKeysQueryParams)

	return s.oasOamMapper.toOasGetElastickeysResponse(err, elasticKeys)
}

// GetMaterialkeys finds Material Keys. Supports optional filtering, sorting, and paging.
// (GET /materialkeys).
func (s *StrictServer) GetMaterialkeys(ctx context.Context, request cryptoutilOpenapiServer.GetMaterialkeysRequestObject) (cryptoutilOpenapiServer.GetMaterialkeysResponseObject, error) {
	keysQueryParams := s.oasOamMapper.toOamGetMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeys(ctx, keysQueryParams)

	return s.oasOamMapper.toOasGetMaterialKeysResponse(err, keys)
}

// PutElastickeyElasticKeyID updates an Elastic Key.
// (PUT /elastickey/{elasticKeyID}).
func (s *StrictServer) PutElastickeyElasticKeyID(ctx context.Context, request cryptoutilOpenapiServer.PutElastickeyElasticKeyIDRequestObject) (cryptoutilOpenapiServer.PutElastickeyElasticKeyIDResponseObject, error) {
	updatedElasticKey, err := s.businessLogicService.UpdateElasticKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPutElastickeyElasticKeyIDResponse(err, updatedElasticKey)
}

// DeleteElastickeyElasticKeyID deletes an Elastic Key (soft delete).
// (DELETE /elastickey/{elasticKeyID}).
func (s *StrictServer) DeleteElastickeyElasticKeyID(ctx context.Context, request cryptoutilOpenapiServer.DeleteElastickeyElasticKeyIDRequestObject) (cryptoutilOpenapiServer.DeleteElastickeyElasticKeyIDResponseObject, error) {
	err := s.businessLogicService.DeleteElasticKey(ctx, &request.ElasticKeyID)

	return s.oasOamMapper.toOasDeleteElastickeyElasticKeyIDResponse(err)
}

// PostElastickeyElasticKeyIDImport imports a Material Key into an Elastic Key.
// (POST /elastickey/{elasticKeyID}/import).
func (s *StrictServer) PostElastickeyElasticKeyIDImport(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDImportRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDImportResponseObject, error) {
	importedMaterialKey, err := s.businessLogicService.ImportMaterialKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDImportResponse(err, importedMaterialKey)
}

// PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke revokes a Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey/{materialKeyID}/revoke).
func (s *StrictServer) PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponseObject, error) {
	err := s.businessLogicService.RevokeMaterialKey(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(err)
}
