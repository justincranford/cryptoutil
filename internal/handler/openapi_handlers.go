package handler

import (
	"context"

	cryptoutilBusinessLogic "cryptoutil/internal/businesslogic"
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
)

// StrictServer implements cryptoutilOpenapiServer.StrictServerInterface
type StrictServer struct {
	businessLogicService *cryptoutilBusinessLogic.KeyPoolService
	openapiMapper        *openapiMapper
}

func NewOpenapiHandler(service *cryptoutilBusinessLogic.KeyPoolService) *StrictServer {
	return &StrictServer{businessLogicService: service, openapiMapper: &openapiMapper{}}
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilServiceModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	addedKeyPool, err := s.businessLogicService.AddKeyPool(ctx, &keyPoolCreate)
	return s.openapiMapper.toPostKeyResponse(err, addedKeyPool)
}

func (s *StrictServer) GetKeypools(ctx context.Context, openapiGetKeypoolRequestObject cryptoutilOpenapiServer.GetKeypoolsRequestObject) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	keyPoolsQueryParams := s.openapiMapper.toServiceModelGetKeyPoolQueryParams(&openapiGetKeypoolRequestObject.Params)
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
	keyGenerateRequest := cryptoutilServiceModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	key, err := s.businessLogicService.GenerateKeyInPoolKey(ctx, keyPoolID, &keyGenerateRequest)
	return s.openapiMapper.toPostKeypoolKeyPoolIDKeyResponse(err, key)
}

func (s *StrictServer) GetKeypoolKeyPoolIDKeys(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyPoolKeysQueryParams := s.openapiMapper.toServiceModelGetKeyPoolKeysQueryParams(&openapiGetKeypoolKeyPoolIDKeyRequestObject.Params)
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
	keysQueryParams := s.openapiMapper.toServiceModelGetKeysQueryParams(&openapiGetKeysRequestObject.Params)
	keys, err := s.businessLogicService.GetKeys(ctx, keysQueryParams)
	return s.openapiMapper.toGetKeysResponse(err, keys)
}
