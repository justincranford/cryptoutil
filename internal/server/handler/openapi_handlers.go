package handler

import (
	"context"

	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
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

func (s *StrictServer) PostElastickey(ctx context.Context, openapiPostElastickeyRequestObject cryptoutilOpenapiServer.PostElastickeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
	elasticKeyCreate := cryptoutilBusinessLogicModel.ElasticKeyCreate(*openapiPostElastickeyRequestObject.Body)
	addedElasticKey, err := s.businessLogicService.AddElasticKey(ctx, &elasticKeyCreate)
	return s.openapiMapper.toPostKeyResponse(err, addedElasticKey)
}

func (s *StrictServer) GetElastickeys(ctx context.Context, openapiGetElastickeyRequestObject cryptoutilOpenapiServer.GetElastickeysRequestObject) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
	elasticKeysQueryParams := s.openapiMapper.toBusinessLogicModelGetElasticKeyQueryParams(&openapiGetElastickeyRequestObject.Params)
	elasticKeys, err := s.businessLogicService.GetElasticKeys(ctx, elasticKeysQueryParams)
	return s.openapiMapper.toGetElastickeysResponse(err, elasticKeys)
}

func (s *StrictServer) GetElastickeyElasticKeyID(ctx context.Context, openapiGetElastickeyElasticKeyIDRequestObject cryptoutilOpenapiServer.GetElastickeyElasticKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
	elasticKeyID := openapiGetElastickeyElasticKeyIDRequestObject.ElasticKeyID
	elasticKey, err := s.businessLogicService.GetElasticKeyByElasticKeyID(ctx, elasticKeyID)
	return s.openapiMapper.toGetElastickeyElasticKeyIDResponse(err, elasticKey)
}

func (s *StrictServer) PostElastickeyElasticKeyIDKey(ctx context.Context, openapiPostElastickeyElasticKeyIDKeyRequestObject cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKeyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKeyResponseObject, error) {
	elasticKeyID := openapiPostElastickeyElasticKeyIDKeyRequestObject.ElasticKeyID
	keyGenerateRequest := cryptoutilBusinessLogicModel.KeyGenerate(*openapiPostElastickeyElasticKeyIDKeyRequestObject.Body)
	key, err := s.businessLogicService.GenerateKeyInPoolKey(ctx, elasticKeyID, &keyGenerateRequest)
	return s.openapiMapper.toPostElastickeyElasticKeyIDKeyResponse(err, key)
}

func (s *StrictServer) GetElastickeyElasticKeyIDKeys(ctx context.Context, openapiGetElastickeyElasticKeyIDKeyRequestObject cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeysRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeysResponseObject, error) {
	elasticKeyID := openapiGetElastickeyElasticKeyIDKeyRequestObject.ElasticKeyID
	elasticKeyKeysQueryParams := s.openapiMapper.toBusinessLogicModelGetElasticKeyKeysQueryParams(&openapiGetElastickeyElasticKeyIDKeyRequestObject.Params)
	keys, err := s.businessLogicService.GetKeysByElasticKey(ctx, elasticKeyID, elasticKeyKeysQueryParams)
	return s.openapiMapper.toGetElastickeyElasticKeyIDKeysResponse(err, keys)
}

func (s *StrictServer) GetElastickeyElasticKeyIDKeyKeyID(ctx context.Context, openapiGetElastickeyElasticKeyIDKeyKeyIDRequestObject cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyIDRequestObject) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyIDResponseObject, error) {
	elasticKeyID := openapiGetElastickeyElasticKeyIDKeyKeyIDRequestObject.ElasticKeyID
	keyID := openapiGetElastickeyElasticKeyIDKeyKeyIDRequestObject.KeyID
	key, err := s.businessLogicService.GetKeyByElasticKeyAndKeyID(ctx, elasticKeyID, keyID)
	return s.openapiMapper.toGetElastickeyElasticKeyIDKeyKeyIDResponse(err, key)
}

func (s *StrictServer) GetKeys(ctx context.Context, openapiGetKeysRequestObject cryptoutilOpenapiServer.GetKeysRequestObject) (cryptoutilOpenapiServer.GetKeysResponseObject, error) {
	keysQueryParams := s.openapiMapper.toBusinessLogicModelGetKeysQueryParams(&openapiGetKeysRequestObject.Params)
	keys, err := s.businessLogicService.GetKeys(ctx, keysQueryParams)
	return s.openapiMapper.toGetKeysResponse(err, keys)
}

func (s *StrictServer) PostElastickeyElasticKeyIDEncrypt(ctx context.Context, openapiPostElastickeyElasticKeyIDEncryptRequestObject cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	elasticKeyID := openapiPostElastickeyElasticKeyIDEncryptRequestObject.ElasticKeyID
	encryptParams := s.openapiMapper.toBusinessLogicModelPostEncryptQueryParams(&openapiPostElastickeyElasticKeyIDEncryptRequestObject.Params)
	clearBytes := []byte(*openapiPostElastickeyElasticKeyIDEncryptRequestObject.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByElasticKeyID(ctx, elasticKeyID, encryptParams, clearBytes)
	return s.openapiMapper.toPostEncryptResponse(err, encryptedBytes)
}

func (s *StrictServer) PostElastickeyElasticKeyIDDecrypt(ctx context.Context, openapiPostElastickeyElasticKeyIDDecryptRequestObject cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	elasticKeyID := openapiPostElastickeyElasticKeyIDDecryptRequestObject.ElasticKeyID
	encryptedBytes := []byte(*openapiPostElastickeyElasticKeyIDDecryptRequestObject.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByElasticKeyID(ctx, elasticKeyID, encryptedBytes)
	return s.openapiMapper.toPostDecryptResponse(err, decryptedBytes)
}

func (s *StrictServer) PostElastickeyElasticKeyIDSign(ctx context.Context, openapiPostElastickeyElasticKeyIDSignRequestObject cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	elasticKeyID := openapiPostElastickeyElasticKeyIDSignRequestObject.ElasticKeyID
	clearBytes := []byte(*openapiPostElastickeyElasticKeyIDSignRequestObject.Body)
	signedBytes, err := s.businessLogicService.PostSignByElasticKeyID(ctx, elasticKeyID, clearBytes)
	return s.openapiMapper.toPostSignResponse(err, signedBytes)
}

func (s *StrictServer) PostElastickeyElasticKeyIDVerify(ctx context.Context, openapiPostElastickeyElasticKeyIDVerifyRequestObject cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyRequestObject) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	elasticKeyID := openapiPostElastickeyElasticKeyIDVerifyRequestObject.ElasticKeyID
	signedBytes := []byte(*openapiPostElastickeyElasticKeyIDVerifyRequestObject.Body)
	verifiedBytes, err := s.businessLogicService.PostVerifyByElasticKeyID(ctx, elasticKeyID, signedBytes)
	return s.openapiMapper.toPostVerifyResponse(err, verifiedBytes)
}
