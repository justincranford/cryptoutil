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
	service       *cryptoutilBusinessLogic.KeyPoolService
	openapiMapper *openapiMapper
}

func NewStrictServer(service *cryptoutilBusinessLogic.KeyPoolService) *StrictServer {
	return &StrictServer{service: service, openapiMapper: &openapiMapper{}}
}

func (s *StrictServer) GetKeypool(ctx context.Context, openapiGetKeypoolRequestObject cryptoutilOpenapiServer.GetKeypoolRequestObject) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	if openapiGetKeypoolRequestObject.Params.Filter != nil && len(string(*openapiGetKeypoolRequestObject.Params.Filter)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolResponseError(fmt.Errorf("query parameter 'filter' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolRequestObject.Params.Sort != nil && len(string(*openapiGetKeypoolRequestObject.Params.Sort)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolResponseError(fmt.Errorf("query parameter 'sort' not supported yet: %w", fiber.ErrBadRequest))
	} else if openapiGetKeypoolRequestObject.Params.Page != nil && len(string(*openapiGetKeypoolRequestObject.Params.Page)) > 0 {
		return s.openapiMapper.toOpenapiGetKeypoolResponseError(fmt.Errorf("query parameter 'page' not supported yet: %w", fiber.ErrBadRequest))
	}
	return s.service.GetKeyPool(ctx)
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilBusinessModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	return s.service.PostKeyPool(ctx, &keyPoolCreate)
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
	return s.service.GetKeyPoolKeyPoolIDKey(ctx, &keyPoolID)
}

func (s *StrictServer) PostKeypoolKeyPoolIDKey(ctx context.Context, openapiPostKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyGenerate := cryptoutilBusinessModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	return s.service.PostKeyPoolKeyPoolIDKey(ctx, &keyPoolID, &keyGenerate)
}
