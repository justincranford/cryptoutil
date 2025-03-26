package handler

import (
	"context"

	cryptoutilBusinessLogic "cryptoutil/internal/businesslogic"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
)

// StrictServer implements cryptoutilServer.StrictServerInterface
type StrictServer struct {
	service *cryptoutilBusinessLogic.KeyPoolService
}

func NewStrictServer(service *cryptoutilBusinessLogic.KeyPoolService) *StrictServer {
	return &StrictServer{service: service}
}

func (s *StrictServer) GetKeypool(ctx context.Context, _ cryptoutilOpenapiServer.GetKeypoolRequestObject) (cryptoutilOpenapiServer.GetKeypoolResponseObject, error) {
	return s.service.GetKeyPool(ctx)
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilOpenapiModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	return s.service.PostKeyPool(ctx, &keyPoolCreate)
}

func (s *StrictServer) GetKeypoolKeyPoolIDKey(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	return s.service.GetKeyPoolKeyPoolIDKey(ctx, &keyPoolID)
}

func (s *StrictServer) PostKeypoolKeyPoolIDKey(ctx context.Context, openapiPostKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyRequestObject) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyGenerate := cryptoutilOpenapiModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	return s.service.PostKeyPoolKeyPoolIDKey(ctx, &keyPoolID, &keyGenerate)
}
