// Copyright (c) 2025 Justin Cranford
//
//

package handler

import (
	"context"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm/kms/server/businesslogic"
)

// StrictServer implements cryptoutilKmsServer.StrictServerInterface.
type StrictServer struct {
	businessLogicService *cryptoutilKmsServerBusinesslogic.BusinessLogicService
	oasOamMapper         *OamOasMapper
}

// NewOpenapiStrictServer creates a new OpenAPI strict server handler.
func NewOpenapiStrictServer(service *cryptoutilKmsServerBusinesslogic.BusinessLogicService) *StrictServer {
	return &StrictServer{businessLogicService: service, oasOamMapper: &OamOasMapper{}}
}

// PostElastickey creates a new Elastic Key.
// (POST /elastickey).
func (s *StrictServer) PostElastickey(ctx context.Context, request cryptoutilKmsServer.PostElastickeyRequestObject) (cryptoutilKmsServer.PostElastickeyResponseObject, error) {
	addedElasticKey, err := s.businessLogicService.AddElasticKey(ctx, request.Body)

	return s.oasOamMapper.toOasPostKeyResponse(err, addedElasticKey)
}

// GetElastickeyElasticKeyID gets an Elastic Key.
// (GET /elastickey/{elasticKeyID}).
func (s *StrictServer) GetElastickeyElasticKeyID(ctx context.Context, request cryptoutilKmsServer.GetElastickeyElasticKeyIDRequestObject) (cryptoutilKmsServer.GetElastickeyElasticKeyIDResponseObject, error) {
	elasticKey, err := s.businessLogicService.GetElasticKeyByElasticKeyID(ctx, &request.ElasticKeyID)

	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDResponse(err, elasticKey)
}

// PostElastickeyElasticKeyIDDecrypt decrypts ciphertext using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/decrypt).
func (s *StrictServer) PostElastickeyElasticKeyIDDecrypt(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDDecryptRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	encryptedBytes := []byte(*request.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptedBytes)

	return s.oasOamMapper.toOasPostDecryptResponse(err, decryptedBytes)
}

// PostElastickeyElasticKeyIDEncrypt encrypts clear text data using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/encrypt).
func (s *StrictServer) PostElastickeyElasticKeyIDEncrypt(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDEncryptRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	encryptParams := s.oasOamMapper.toOamPostEncryptQueryParams(&request.Params)
	clearBytes := []byte(*request.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptParams, clearBytes)

	return s.oasOamMapper.toOasPostEncryptResponse(err, encryptedBytes)
}

// PostElastickeyElasticKeyIDGenerate generates a random Secret Key, Key Pair, or other algorithm. It will be in JWK format, returned in encrypted form as a JWE message.
// (POST /elastickey/{elasticKeyID}/generate).
func (s *StrictServer) PostElastickeyElasticKeyIDGenerate(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerateRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerateResponseObject, error) {
	generateParams := s.oasOamMapper.toOamPostGenerateQueryParams(&request.Params)
	encryptedNonPublicJWKBytes, _, clearPublicJWKBytes, err := s.businessLogicService.PostGenerateByElasticKeyID(ctx, &request.ElasticKeyID, generateParams)

	return s.oasOamMapper.toOasPostGenerateResponse(err, encryptedNonPublicJWKBytes, clearPublicJWKBytes)
}

// PostElastickeyElasticKeyIDMaterialkey generates a new Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey).
func (s *StrictServer) PostElastickeyElasticKeyIDMaterialkey(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	key, err := s.businessLogicService.GenerateMaterialKeyInElasticKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(err, key)
}

// GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID gets Material Key in Elastic Key.
// (GET /elastickey/{elasticKeyID}/materialkey/{materialKeyID}).
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID(ctx context.Context, request cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRequestObject) (cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	key, err := s.businessLogicService.GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err, key)
}

// GetElastickeyElasticKeyIDMaterialkeys finds Material Keys in Elastic Key. Supports optional filtering, sorting, and paging.
// (GET /elastickey/{elasticKeyID}/materialkeys).
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeys(ctx context.Context, request cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeysRequestObject) (cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	elasticKeyMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeysForElasticKey(ctx, &request.ElasticKeyID, elasticKeyMaterialKeysQueryParams)

	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDMaterialkeysResponse(err, keys)
}

// PostElastickeyElasticKeyIDSign signs clear text using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/sign).
func (s *StrictServer) PostElastickeyElasticKeyIDSign(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDSignRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	clearBytes := []byte(*request.Body)
	signedBytes, err := s.businessLogicService.PostSignByElasticKeyID(ctx, &request.ElasticKeyID, clearBytes)

	return s.oasOamMapper.toOasPostSignResponse(err, signedBytes)
}

// PostElastickeyElasticKeyIDVerify verifies JWS message using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/verify).
func (s *StrictServer) PostElastickeyElasticKeyIDVerify(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDVerifyRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	signedBytes := []byte(*request.Body)
	_, err := s.businessLogicService.PostVerifyByElasticKeyID(ctx, &request.ElasticKeyID, signedBytes)

	return s.oasOamMapper.toOasPostVerifyResponse(err)
}

// GetElastickeys finds Elastic Keys. Supports optional filtering, sorting, and paging.
// (GET /elastickeys).
func (s *StrictServer) GetElastickeys(ctx context.Context, request cryptoutilKmsServer.GetElastickeysRequestObject) (cryptoutilKmsServer.GetElastickeysResponseObject, error) {
	elasticMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyQueryParams(&request.Params)
	elasticKeys, err := s.businessLogicService.GetElasticKeys(ctx, elasticMaterialKeysQueryParams)

	return s.oasOamMapper.toOasGetElastickeysResponse(err, elasticKeys)
}

// GetMaterialkeys finds Material Keys. Supports optional filtering, sorting, and paging.
// (GET /materialkeys).
func (s *StrictServer) GetMaterialkeys(ctx context.Context, request cryptoutilKmsServer.GetMaterialkeysRequestObject) (cryptoutilKmsServer.GetMaterialkeysResponseObject, error) {
	keysQueryParams := s.oasOamMapper.toOamGetMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeys(ctx, keysQueryParams)

	return s.oasOamMapper.toOasGetMaterialKeysResponse(err, keys)
}

// PutElastickeyElasticKeyID updates an Elastic Key.
// (PUT /elastickey/{elasticKeyID}).
func (s *StrictServer) PutElastickeyElasticKeyID(ctx context.Context, request cryptoutilKmsServer.PutElastickeyElasticKeyIDRequestObject) (cryptoutilKmsServer.PutElastickeyElasticKeyIDResponseObject, error) {
	updatedElasticKey, err := s.businessLogicService.UpdateElasticKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPutElastickeyElasticKeyIDResponse(err, updatedElasticKey)
}

// DeleteElastickeyElasticKeyID deletes an Elastic Key (soft delete).
// (DELETE /elastickey/{elasticKeyID}).
func (s *StrictServer) DeleteElastickeyElasticKeyID(ctx context.Context, request cryptoutilKmsServer.DeleteElastickeyElasticKeyIDRequestObject) (cryptoutilKmsServer.DeleteElastickeyElasticKeyIDResponseObject, error) {
	err := s.businessLogicService.DeleteElasticKey(ctx, &request.ElasticKeyID)

	return s.oasOamMapper.toOasDeleteElastickeyElasticKeyIDResponse(err)
}

// PostElastickeyElasticKeyIDImport imports a Material Key into an Elastic Key.
// (POST /elastickey/{elasticKeyID}/import).
func (s *StrictServer) PostElastickeyElasticKeyIDImport(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDImportRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDImportResponseObject, error) {
	importedMaterialKey, err := s.businessLogicService.ImportMaterialKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDImportResponse(err, importedMaterialKey)
}

// PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke revokes a Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey/{materialKeyID}/revoke).
func (s *StrictServer) PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke(ctx context.Context, request cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeRequestObject) (cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponseObject, error) {
	err := s.businessLogicService.RevokeMaterialKey(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(err)
}

// DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID deletes a Material Key from an Elastic Key.
// (DELETE /elastickey/{elasticKeyID}/materialkey/{materialKeyID}).
func (s *StrictServer) DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID(ctx context.Context, request cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRequestObject) (cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	err := s.businessLogicService.DeleteMaterialKey(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasDeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err)
}
