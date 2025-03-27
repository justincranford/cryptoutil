package handler

import (
	"context"
	"fmt"

	cryptoutilBusinessLogic "cryptoutil/internal/businesslogic"
	cryptoutilBusinessModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"

	"github.com/gofiber/fiber/v2"
)

// StrictServer implements cryptoutilServer.StrictServerInterface
type StrictServer struct {
	businessLogicService *cryptoutilBusinessLogic.KeyPoolService
	openapiMapper        *openapiMapper
}

func NewOpenapiHandler(service *cryptoutilBusinessLogic.KeyPoolService) *StrictServer {
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
	listKeyPoolsResponse, err := s.businessLogicService.ListKeyPools(ctx)
	return listKeyPoolsResponse, err
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreateRequest := cryptoutilBusinessModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	keyPoolCreateResponse, err := s.businessLogicService.AddKeyPool(ctx, &keyPoolCreateRequest)
	return keyPoolCreateResponse, err
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
	listKeysResponse, err := s.businessLogicService.ListKeysByKeyPool(ctx, keyPoolID)
	return listKeysResponse, err
}

func (s *StrictServer) PostKeypoolKeyPoolIDKey(ctx context.Context, openapiPostKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyGenerateRequest := cryptoutilBusinessModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	generateKeyInKeyPoolResponse, err := s.businessLogicService.GenerateKeyInPoolKey(ctx, keyPoolID, &keyGenerateRequest)
	return generateKeyInKeyPoolResponse, err
}
