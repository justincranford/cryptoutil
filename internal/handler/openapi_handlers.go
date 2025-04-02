package handler

import (
	"context"
	"fmt"

	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilServiceLogic "cryptoutil/internal/servicelogic"

	"github.com/gofiber/fiber/v2"
)

// StrictServer implements cryptoutilOpenapiServer.StrictServerInterface
type StrictServer struct {
	businessLogicService *cryptoutilServiceLogic.KeyPoolService
	openapiMapper        *openapiMapper
}

func NewOpenapiHandler(service *cryptoutilServiceLogic.KeyPoolService) *StrictServer {
	return &StrictServer{businessLogicService: service, openapiMapper: &openapiMapper{}}
}

func (s *StrictServer) GetKeypool(ctx context.Context, openapiGetKeypoolRequestObject cryptoutilOpenapiServer.GetKeypoolRequestObject) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	keyPoolQueryParams := s.openapiMapper.toServiceModelGetKeyPoolQueryParams(&openapiGetKeypoolRequestObject.Params)
	keyPools, err := s.businessLogicService.ListKeyPools(ctx, keyPoolQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list KeyPools: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypool200JSONResponse(keyPools), err
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilServiceModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	addedKeyPool, err := s.businessLogicService.AddKeyPool(ctx, &keyPoolCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to add KeyPool: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypool200JSONResponse(*addedKeyPool), nil
}

func (s *StrictServer) GetKeypoolKeyPoolIDKey(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	if openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Sort != nil && len(*openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Sort) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolKeyPoolIDKeyResponseError(fmt.Errorf("query parameter 'sort' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Page != nil && *openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Page >= 0 {
		return s.openapiMapper.toOpenapiGetKeypoolKeyPoolIDKeyResponseError(fmt.Errorf("query parameter 'page' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Size != nil && *openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Size > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolKeyPoolIDKeyResponseError(fmt.Errorf("query parameter 'size' not supported yet: %w", fiber.ErrBadRequest))
	}
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keys, err := s.businessLogicService.ListKeysByKeyPool(ctx, keyPoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKey200JSONResponse(keys), err
}

func (s *StrictServer) PostKeypoolKeyPoolIDKey(ctx context.Context, openapiPostKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyGenerateRequest := cryptoutilServiceModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	generateKeyInKeyPoolResponse, err := s.businessLogicService.GenerateKeyInPoolKey(ctx, keyPoolID, &keyGenerateRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*generateKeyInKeyPoolResponse), err
}
