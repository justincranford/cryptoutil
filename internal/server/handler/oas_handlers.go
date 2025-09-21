package handler

import (
	"context"

	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
)

// StrictServer implements cryptoutilOpenapiServer.StrictServerInterface
type StrictServer struct {
	businessLogicService *cryptoutilBusinessLogic.BusinessLogicService
	oasOamMapper         *oamOasMapper
}

func NewOpenapiStrictServer(service *cryptoutilBusinessLogic.BusinessLogicService) *StrictServer {
	return &StrictServer{businessLogicService: service, oasOamMapper: &oamOasMapper{}}
}

// Create a new Elastic Key.
// (POST /elastickey)
func (s *StrictServer) PostElastickey(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
	addedElasticKey, err := s.businessLogicService.AddElasticKey(ctx, request.Body)
	return s.oasOamMapper.toOasPostKeyResponse(err, addedElasticKey)
}

// Get an Elastic Key.
// (GET /elastickey/{elasticKeyID})
func (s *StrictServer) GetElastickeyElasticKeyID(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
	elasticKey, err := s.businessLogicService.GetElasticKeyByElasticKeyID(ctx, &request.ElasticKeyID)
	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDResponse(err, elasticKey)
}

// Decrypt ciphertext using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/decrypt)
func (s *StrictServer) PostElastickeyElasticKeyIDDecrypt(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	encryptedBytes := []byte(*request.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptedBytes)
	return s.oasOamMapper.toOasPostDecryptResponse(err, decryptedBytes)
}

// Encrypt clear text data using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/encrypt)
func (s *StrictServer) PostElastickeyElasticKeyIDEncrypt(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	encryptParams := s.oasOamMapper.toOamPostEncryptQueryParams(&request.Params)
	clearBytes := []byte(*request.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByElasticKeyID(ctx, &request.ElasticKeyID, encryptParams, clearBytes)
	return s.oasOamMapper.toOasPostEncryptResponse(err, encryptedBytes)
}

// Generate a random Secret Key, Key Pair, or other algorithm. It will be in JWK format, returned in encrypted form as a JWE message.
// (POST /elastickey/{elasticKeyID}/generate)
func (s *StrictServer) PostElastickeyElasticKeyIDGenerate(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateResponseObject, error) {
	generateParams := s.oasOamMapper.toOamPostGenerateQueryParams(&request.Params)
	encryptedNonPublicJwkBytes, clearNonPublicJwkBytes, clearPublicJwkBytes, err := s.businessLogicService.PostGenerateByElasticKeyID(ctx, &request.ElasticKeyID, generateParams)
	return s.oasOamMapper.toOasPostGenerateResponse(err, encryptedNonPublicJwkBytes, clearNonPublicJwkBytes, clearPublicJwkBytes)
}

// Generate a new Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey)
func (s *StrictServer) PostElastickeyElasticKeyIDMaterialkey(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	key, err := s.businessLogicService.GenerateMaterialKeyInElasticKey(ctx, &request.ElasticKeyID, request.Body)
	return s.oasOamMapper.toOasPostElastickeyElasticKeyIDMaterialkeyResponse(err, key)
}

// Get Material Key in Elastic Key.
// (GET /elastickey/{elasticKeyID}/materialkey/{materialKeyID})
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	key, err := s.businessLogicService.GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx, &request.ElasticKeyID, &request.MaterialKeyID)
	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err, key)
}

// Find Material Keys in Elastic Key. Supports optional filtering, sorting, and paging.
// (GET /elastickey/{elasticKeyID}/materialkeys)
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeys(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	elasticKeyMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeysForElasticKey(ctx, &request.ElasticKeyID, elasticKeyMaterialKeysQueryParams)
	return s.oasOamMapper.toOasGetElastickeyElasticKeyIDMaterialkeysResponse(err, keys)
}

// Sign clear text using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/sign)
func (s *StrictServer) PostElastickeyElasticKeyIDSign(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	clearBytes := []byte(*request.Body)
	signedBytes, err := s.businessLogicService.PostSignByElasticKeyID(ctx, &request.ElasticKeyID, clearBytes)
	return s.oasOamMapper.toOasPostSignResponse(err, signedBytes)
}

// Verify JWS message using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/verify)
func (s *StrictServer) PostElastickeyElasticKeyIDVerify(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	signedBytes := []byte(*request.Body)
	verifiedBytes, err := s.businessLogicService.PostVerifyByElasticKeyID(ctx, &request.ElasticKeyID, signedBytes)
	return s.oasOamMapper.toOasPostVerifyResponse(err, verifiedBytes)
}

// Find Elastic Keys. Supports optional filtering, sorting, and paging.
// (GET /elastickeys)
func (s *StrictServer) GetElastickeys(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeysRequestObject) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
	elasticMaterialKeysQueryParams := s.oasOamMapper.toOamGetElasticKeyQueryParams(&request.Params)
	elasticKeys, err := s.businessLogicService.GetElasticKeys(ctx, elasticMaterialKeysQueryParams)
	return s.oasOamMapper.toOasGetElastickeysResponse(err, elasticKeys)
}

// Find Material Keys. Supports optional filtering, sorting, and paging.
// (GET /materialkeys)
func (s *StrictServer) GetMaterialkeys(ctx context.Context, request cryptoutilOpenapiServer.GetMaterialkeysRequestObject) (cryptoutilOpenapiServer.GetMaterialkeysResponseObject, error) {
	keysQueryParams := s.oasOamMapper.toOamGetMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeys(ctx, keysQueryParams)
	return s.oasOamMapper.toOasGetMaterialKeysResponse(err, keys)
}
