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

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilBusinessLogicModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	addedKeyPool, err := s.businessLogicService.AddKeyPool(ctx, &keyPoolCreate)
	return s.openapiMapper.toPostKeyResponse(err, addedKeyPool)
}

func (s *StrictServer) GetKeypools(ctx context.Context, openapiGetKeypoolRequestObject cryptoutilOpenapiServer.GetKeypoolsRequestObject) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	keyPoolsQueryParams := s.openapiMapper.toBusinessLogicModelGetKeyPoolQueryParams(&openapiGetKeypoolRequestObject.Params)
	keyPools, err := s.businessLogicService.GetKeyPools(ctx, keyPoolsQueryParams)
	return s.openapiMapper.toGetKeypoolsResponse(err, keyPools)
}

func (s *StrictServer) GetKeypoolKeyPoolID(ctx context.Context, openapiGetKeypoolKeyPoolIDRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDRequestObject.KeyPoolID
	keyPool, err := s.businessLogicService.GetKeyPoolByKeyPoolID(ctx, keyPoolID)
	return s.openapiMapper.toGetKeypoolKeyPoolIDResponse(err, keyPool)
}

func (s *StrictServer) PostKeypoolKeyPoolIDKey(ctx context.Context, openapiPostKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyGenerateRequest := cryptoutilBusinessLogicModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	key, err := s.businessLogicService.GenerateKeyInPoolKey(ctx, keyPoolID, &keyGenerateRequest)
	return s.openapiMapper.toPostKeypoolKeyPoolIDKeyResponse(err, key)
}

func (s *StrictServer) GetKeypoolKeyPoolIDKeys(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyPoolKeysQueryParams := s.openapiMapper.toBusinessLogicModelGetKeyPoolKeysQueryParams(&openapiGetKeypoolKeyPoolIDKeyRequestObject.Params)
	keys, err := s.businessLogicService.GetKeysByKeyPool(ctx, keyPoolID, keyPoolKeysQueryParams)
	return s.openapiMapper.toGetKeypoolKeyPoolIDKeysResponse(err, keys)
}

func (s *StrictServer) GetKeypoolKeyPoolIDKeyKeyID(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyKeyIDRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyIDRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyIDResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyKeyIDRequestObject.KeyPoolID
	keyID := openapiGetKeypoolKeyPoolIDKeyKeyIDRequestObject.KeyID
	key, err := s.businessLogicService.GetKeyByKeyPoolAndKeyID(ctx, keyPoolID, keyID)
	return s.openapiMapper.toGetKeypoolKeyPoolIDKeyKeyIDResponse(err, key)
}

func (s *StrictServer) GetKeys(ctx context.Context, openapiGetKeysRequestObject cryptoutilOpenapiServer.GetKeysRequestObject) (cryptoutilOpenapiServer.GetKeysResponseObject, error) {
	keysQueryParams := s.openapiMapper.toBusinessLogicModelGetKeysQueryParams(&openapiGetKeysRequestObject.Params)
	keys, err := s.businessLogicService.GetKeys(ctx, keysQueryParams)
	return s.openapiMapper.toGetKeysResponse(err, keys)
}

func (s *StrictServer) PostKeypoolKeyPoolIDEncrypt(ctx context.Context, openapiPostKeypoolKeyPoolIDEncryptRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncryptRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncryptResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDEncryptRequestObject.KeyPoolID
	encryptParams := s.openapiMapper.toBusinessLogicModelPostEncryptQueryParams(&openapiPostKeypoolKeyPoolIDEncryptRequestObject.Params)
	clearBytes := []byte(*openapiPostKeypoolKeyPoolIDEncryptRequestObject.Body)
	encryptedBytes, err := s.businessLogicService.PostEncryptByKeyPoolID(ctx, keyPoolID, encryptParams, clearBytes)
	return s.openapiMapper.toPostEncryptResponse(err, encryptedBytes)
}

func (s *StrictServer) PostKeypoolKeyPoolIDDecrypt(ctx context.Context, openapiPostKeypoolKeyPoolIDDecryptRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecryptRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecryptResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDDecryptRequestObject.KeyPoolID
	encryptedBytes := []byte(*openapiPostKeypoolKeyPoolIDDecryptRequestObject.Body)
	decryptedBytes, err := s.businessLogicService.PostDecryptByKeyPoolID(ctx, keyPoolID, encryptedBytes)
	return s.openapiMapper.toPostDecryptResponse(err, decryptedBytes)
}

func (s *StrictServer) PostKeypoolKeyPoolIDSign(ctx context.Context, openapiPostKeypoolKeyPoolIDSignRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSignRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSignResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDSignRequestObject.KeyPoolID
	clearBytes := []byte(*openapiPostKeypoolKeyPoolIDSignRequestObject.Body)
	signedBytes, err := s.businessLogicService.PostSignByKeyPoolID(ctx, keyPoolID, clearBytes)
	return s.openapiMapper.toPostSignResponse(err, signedBytes)
}

func (s *StrictServer) PostKeypoolKeyPoolIDVerify(ctx context.Context, openapiPostKeypoolKeyPoolIDVerifyRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerifyRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerifyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDVerifyRequestObject.KeyPoolID
	signedBytes := []byte(*openapiPostKeypoolKeyPoolIDVerifyRequestObject.Body)
	verifiedBytes, err := s.businessLogicService.PostVerifyByKeyPoolID(ctx, keyPoolID, signedBytes)
	return s.openapiMapper.toPostVerifyResponse(err, verifiedBytes)
}
