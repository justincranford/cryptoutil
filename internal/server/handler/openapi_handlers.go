package handler

import (
	"context"

	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
)

// StrictServer implements cryptoutilOpenapiServer.StrictServerInterface
type StrictServer struct {
	businessLogicService *cryptoutilBusinessLogic.BusinessLogicService
	openapiMapper        *openapiBusinessLogicMapper
}

func NewOpenapiStrictServer(service *cryptoutilBusinessLogic.BusinessLogicService) *StrictServer {
	return &StrictServer{businessLogicService: service, openapiMapper: &openapiBusinessLogicMapper{}}
}

// Create a new Elastic Key.
// (POST /elastickey)
func (s *StrictServer) PostElastickey(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
	elasticKeyCreate := cryptoutilOpenapiModel.ElasticKeyCreate(*request.Body)
	addedElasticKey, err := s.businessLogicService.AddElasticKey(ctx, &elasticKeyCreate)
	return s.openapiMapper.toPostKeyResponse(err, addedElasticKey)
}

// Get an Elastic Key.
// (GET /elastickey/{elasticKeyID})
func (s *StrictServer) GetElastickeyElasticKeyID(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	elasticKey, err := s.businessLogicService.GetElasticKeyByElasticKeyID(ctx, elasticKeyID)
	return s.openapiMapper.toGetElastickeyElasticKeyIDResponse(err, elasticKey)
}

// Decrypt ciphertext using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/decrypt)
func (s *StrictServer) PostElastickeyElasticKeyIDDecrypt(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	encryptedBytes := []byte(*request.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByElasticKeyID(ctx, elasticKeyID, encryptedBytes)
	return s.openapiMapper.toPostDecryptResponse(err, decryptedBytes)
}

// Encrypt clear text data using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWE message kid header.
// (POST /elastickey/{elasticKeyID}/encrypt)
func (s *StrictServer) PostElastickeyElasticKeyIDEncrypt(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	encryptParams := s.openapiMapper.toBusinessLogicModelPostEncryptQueryParams(&request.Params)
	clearBytes := []byte(*request.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByElasticKeyID(ctx, elasticKeyID, encryptParams, clearBytes)
	return s.openapiMapper.toPostEncryptResponse(err, encryptedBytes)
}

// Generate a new Material Key in an Elastic Key.
// (POST /elastickey/{elasticKeyID}/materialkey)
func (s *StrictServer) PostElastickeyElasticKeyIDMaterialkey(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	keyGenerateRequest := cryptoutilOpenapiModel.MaterialKeyGenerate(*request.Body)
	key, err := s.businessLogicService.GenerateKeyInPoolKey(ctx, elasticKeyID, &keyGenerateRequest)
	return s.openapiMapper.toPostElastickeyElasticKeyIDMaterialkeyResponse(err, key)
}

// Get Material Key in Elastic Key.
// (GET /elastickey/{elasticKeyID}/materialkey/{materialKeyID})
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	materialKeyID := request.MaterialKeyID
	key, err := s.businessLogicService.GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx, elasticKeyID, materialKeyID)
	return s.openapiMapper.toGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err, key)
}

// Find Material Keys in Elastic Key. Supports optional filtering, sorting, and paging.
// (GET /elastickey/{elasticKeyID}/materialkeys)
func (s *StrictServer) GetElastickeyElasticKeyIDMaterialkeys(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	elasticKeyMaterialKeysQueryParams := s.openapiMapper.toBusinessLogicModelGetElasticKeyMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeysForElasticKey(ctx, elasticKeyID, elasticKeyMaterialKeysQueryParams)
	return s.openapiMapper.toGetElastickeyElasticKeyIDMaterialkeysResponse(err, keys)
}

// Sign clear text using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/sign)
func (s *StrictServer) PostElastickeyElasticKeyIDSign(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	clearBytes := []byte(*request.Body)
	signedBytes, err := s.businessLogicService.PostSignByElasticKeyID(ctx, elasticKeyID, clearBytes)
	return s.openapiMapper.toPostSignResponse(err, signedBytes)
}

// Verify JWS message using a specific Elastic Key. The Material Key in the Elastic Key is identified by the JWS message kid header.
// (POST /elastickey/{elasticKeyID}/verify)
func (s *StrictServer) PostElastickeyElasticKeyIDVerify(ctx context.Context, request cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	elasticKeyID := request.ElasticKeyID
	signedBytes := []byte(*request.Body)
	verifiedBytes, err := s.businessLogicService.PostVerifyByElasticKeyID(ctx, elasticKeyID, signedBytes)
	return s.openapiMapper.toPostVerifyResponse(err, verifiedBytes)
}

// Find Elastic Keys. Supports optional filtering, sorting, and paging.
// (GET /elastickeys)
func (s *StrictServer) GetElastickeys(ctx context.Context, request cryptoutilOpenapiServer.GetElastickeysRequestObject) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
	elasticMaterialKeysQueryParams := s.openapiMapper.toBusinessLogicModelGetElasticKeyQueryParams(&request.Params)
	elasticKeys, err := s.businessLogicService.GetElasticKeys(ctx, elasticMaterialKeysQueryParams)
	return s.openapiMapper.toGetElastickeysResponse(err, elasticKeys)
}

// Find Material Keys. Supports optional filtering, sorting, and paging.
// (GET /materialkeys)
func (s *StrictServer) GetMaterialkeys(ctx context.Context, request cryptoutilOpenapiServer.GetMaterialkeysRequestObject) (cryptoutilOpenapiServer.GetMaterialkeysResponseObject, error) {
	keysQueryParams := s.openapiMapper.toBusinessLogicModelGetMaterialKeysQueryParams(&request.Params)
	keys, err := s.businessLogicService.GetMaterialKeys(ctx, keysQueryParams)
	return s.openapiMapper.toGetMaterialKeysResponse(err, keys)
}
