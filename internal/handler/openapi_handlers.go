package handler

import (
	"context"
	"fmt"

	cryptoutilServiceLogic "cryptoutil/internal/businesslogic"
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"

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
	if openapiGetKeypoolRequestObject.Params.Filter != nil && len(string(*openapiGetKeypoolRequestObject.Params.Filter)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolResponseError(fmt.Errorf("query parameter 'filter' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolRequestObject.Params.Sort != nil && len(string(*openapiGetKeypoolRequestObject.Params.Sort)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolResponseError(fmt.Errorf("query parameter 'sort' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolRequestObject.Params.Page != nil && len(string(*openapiGetKeypoolRequestObject.Params.Page)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolResponseError(fmt.Errorf("query parameter 'page' not supported yet: %w", fiber.ErrBadRequest))
	}
	keyPools, err := s.businessLogicService.ListKeyPools(ctx)
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
	if openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Filter != nil && len(string(*openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Filter)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolKeyPoolIDKeyResponseError(fmt.Errorf("query parameter 'filter' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Sort != nil && len(string(*openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Sort)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolKeyPoolIDKeyResponseError(fmt.Errorf("query parameter 'sort' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Page != nil && len(string(*openapiGetKeypoolKeyPoolIDKeyRequestObject.Params.Page)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolKeyPoolIDKeyResponseError(fmt.Errorf("query parameter 'page' not supported yet: %w", fiber.ErrBadRequest))
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
