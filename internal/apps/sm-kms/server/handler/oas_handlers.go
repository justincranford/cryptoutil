// Copyright (c) 2025-2026 Justin Cranford.
//
//

package handler

import (
	"context"

	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm-kms/server/businesslogic"
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
func (s *StrictServer) PostElasticKeys(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysRequestObject) (cryptoutilKmsServer.PostElasticKeysResponseObject, error) {
	addedElasticKey, err := s.businessLogicService.AddElasticKey(ctx, request.Body)

	return s.oasOamMapper.toOasPostElasticKeysResponse(err, addedElasticKey)
}

// GetElastickeyElasticKeyID gets an Elastic Key.
// (GET /elastickey/{elasticKeyID}).
func (s *StrictServer) GetElasticKeysElasticKeyID(ctx context.Context, request cryptoutilKmsServer.GetElasticKeysElasticKeyIDRequestObject) (cryptoutilKmsServer.GetElasticKeysElasticKeyIDResponseObject, error) {
	elasticKey, err := s.businessLogicService.GetElasticKeyByElasticKeyID(ctx, &request.ElasticKeyID)

	return s.oasOamMapper.toOasGetElasticKeysElasticKeyIDResponse(err, elasticKey)
}

// PostElastickeyElasticKeyIDDecrypt decrypts ciphertext using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/decrypt).
func (s *StrictServer) PostElasticKeysElasticKeyIDDecrypt(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecryptRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecryptResponseObject, error) {
	encryptedBytes := []byte(*request.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptedBytes)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDDecryptResponse(err, decryptedBytes)
}

// PostElastickeyElasticKeyIDEncrypt encrypts clear text data using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/encrypt).
func (s *StrictServer) PostElasticKeysElasticKeyIDEncrypt(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncryptRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncryptResponseObject, error) {
	encryptParams := s.oasOamMapper.toOamPostElasticKeysElasticKeyIDEncryptQueryParams(&request.Params)
	clearBytes := []byte(*request.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptParams, clearBytes)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDEncryptResponse(err, encryptedBytes)
}

// PostElastickeyElasticKeyIDGenerate generates a random Secret Key, Key Pair, or other algorithm. It will be in JWK format, returned in encrypted form as a JWE message.
// (POST /elastickey/{elasticKeyID}/generate).
func (s *StrictServer) PostElasticKeysElasticKeyIDGenerate(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerateRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerateResponseObject, error) {
	generateParams := s.oasOamMapper.toOamPostElasticKeysElasticKeyIDGenerateQueryParams(&request.Params)
	encryptedNonPublicJWKBytes, _, clearPublicJWKBytes, err := s.businessLogicService.PostGenerateByElasticKeyID(ctx, &request.ElasticKeyID, generateParams)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDGenerateResponse(err, encryptedNonPublicJWKBytes, clearPublicJWKBytes)
}

// PostElastickeyElasticKeyIDMaterialkey generates a new Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey).
func (s *StrictServer) PostElasticKeysElasticKeyIDMaterialKeys(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysResponseObject, error) {
	key, err := s.businessLogicService.GenerateMaterialKeyInElasticKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDMaterialKeysResponse(err, key)
}

// GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID gets Material Key in Elastic Key.
// (GET /elastickey/{elasticKeyID}/materialkey/{materialKeyID}).
func (s *StrictServer) GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyID(ctx context.Context, request cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRequestObject) (cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponseObject, error) {
	key, err := s.businessLogicService.GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasGetElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponse(err, key)
}

// GetElastickeyElasticKeyIDMaterialkeys finds Material Keys in Elastic Key. Supports optional filtering, sorting, and paging.
// (GET /elastickey/{elasticKeyID}/materialkeys).
func (s *StrictServer) GetElasticKeysElasticKeyIDMaterialKeys(ctx context.Context, request cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysRequestObject) (cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysResponseObject, error) {
	elasticKeyMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeysElasticKeyIDMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeysForElasticKey(ctx, &request.ElasticKeyID, elasticKeyMaterialKeysQueryParams)

	return s.oasOamMapper.toOasGetElasticKeysElasticKeyIDMaterialKeysResponse(err, keys)
}

// PostElastickeyElasticKeyIDSign signs clear text using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/sign).
func (s *StrictServer) PostElasticKeysElasticKeyIDSign(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDSignRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDSignResponseObject, error) {
	clearBytes := []byte(*request.Body)
	signedBytes, err := s.businessLogicService.PostSignByElasticKeyID(ctx, &request.ElasticKeyID, clearBytes)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDSignResponse(err, signedBytes)
}

// PostElastickeyElasticKeyIDVerify verifies JWS message using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/verify).
func (s *StrictServer) PostElasticKeysElasticKeyIDVerify(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerifyRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerifyResponseObject, error) {
	signedBytes := []byte(*request.Body)
	_, err := s.businessLogicService.PostVerifyByElasticKeyID(ctx, &request.ElasticKeyID, signedBytes)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDVerifyResponse(err)
}

// GetElastickeys finds Elastic Keys. Supports optional filtering, sorting, and paging.
// (GET /elastickeys).
func (s *StrictServer) GetElasticKeys(ctx context.Context, request cryptoutilKmsServer.GetElasticKeysRequestObject) (cryptoutilKmsServer.GetElasticKeysResponseObject, error) {
	elasticMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyQueryParams(&request.Params)
	elasticKeys, err := s.businessLogicService.GetElasticKeys(ctx, elasticMaterialKeysQueryParams)

	return s.oasOamMapper.toOasGetElasticKeysResponse(err, elasticKeys)
}

// GetMaterialkeys finds Material Keys. Supports optional filtering, sorting, and paging.
// (GET /materialkeys).
func (s *StrictServer) GetMaterialKeys(ctx context.Context, request cryptoutilKmsServer.GetMaterialKeysRequestObject) (cryptoutilKmsServer.GetMaterialKeysResponseObject, error) {
	keysQueryParams := s.oasOamMapper.toOamGetMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeys(ctx, keysQueryParams)

	return s.oasOamMapper.toOasGetMaterialKeysResponse(err, keys)
}

// PutElastickeyElasticKeyID updates an Elastic Key.
// (PUT /elastickey/{elasticKeyID}).
func (s *StrictServer) PutElasticKeysElasticKeyID(ctx context.Context, request cryptoutilKmsServer.PutElasticKeysElasticKeyIDRequestObject) (cryptoutilKmsServer.PutElasticKeysElasticKeyIDResponseObject, error) {
	updatedElasticKey, err := s.businessLogicService.UpdateElasticKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPutElasticKeysElasticKeyIDResponse(err, updatedElasticKey)
}

// DeleteElastickeyElasticKeyID deletes an Elastic Key (soft delete).
// (DELETE /elastickey/{elasticKeyID}).
func (s *StrictServer) DeleteElasticKeysElasticKeyID(ctx context.Context, request cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDRequestObject) (cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDResponseObject, error) {
	err := s.businessLogicService.DeleteElasticKey(ctx, &request.ElasticKeyID)

	return s.oasOamMapper.toOasDeleteElasticKeysElasticKeyIDResponse(err)
}

// PostElastickeyElasticKeyIDImport imports a Material Key into an Elastic Key.
// (POST /elastickey/{elasticKeyID}/import).
func (s *StrictServer) PostElasticKeysElasticKeyIDImport(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDImportRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDImportResponseObject, error) {
	importedMaterialKey, err := s.businessLogicService.ImportMaterialKey(ctx, &request.ElasticKeyID, request.Body)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDImportResponse(err, importedMaterialKey)
}

// PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke revokes a Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey/{materialKeyID}/revoke).
func (s *StrictServer) PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevoke(ctx context.Context, request cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevokeRequestObject) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevokeResponseObject, error) {
	err := s.businessLogicService.RevokeMaterialKey(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasPostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevokeResponse(err)
}

// DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID deletes a Material Key from an Elastic Key.
// (DELETE /elastickey/{elasticKeyID}/materialkey/{materialKeyID}).
func (s *StrictServer) DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyID(ctx context.Context, request cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRequestObject) (cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponseObject, error) {
	err := s.businessLogicService.DeleteMaterialKey(ctx, &request.ElasticKeyID, &request.MaterialKeyID)

	return s.oasOamMapper.toOasDeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponse(err)
}
