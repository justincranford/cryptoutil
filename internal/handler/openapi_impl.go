package handler

import (
	"context"

	cryptoutilModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/openapi/server"
	cryptoutilService "cryptoutil/internal/service"
)

// StrictServer implements cryptoutilServer.StrictServerInterface
type StrictServer struct {
	service *cryptoutilService.KeyPoolService
}

func NewStrictServer(service *cryptoutilService.KeyPoolService) *StrictServer {
	return &StrictServer{service: service}
}

func (s *StrictServer) GetKeypool(ctx context.Context, _ cryptoutilServer.GetKeypoolRequestObject) (cryptoutilServer.GetKeypoolResponseObject, error) {
	return s.service.GetKeyPool(ctx)
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilServer.PostKeypoolRequestObject) (cryptoutilServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	return s.service.PostKeyPool(ctx, &keyPoolCreate)
}

func (s *StrictServer) GetKeypoolKeyPoolIDKey(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyRequestObject cryptoutilServer.GetKeypoolKeyPoolIDKeyRequestObject) (cryptoutilServer.GetKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	return s.service.GetKeyPoolKeyPoolIDKey(ctx, &keyPoolID)
}

func (s *StrictServer) PostKeypoolKeyPoolIDKey(ctx context.Context, openapiPostKeypoolKeyPoolIDKeyRequestObject cryptoutilServer.PostKeypoolKeyPoolIDKeyRequestObject) (cryptoutilServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	keyPoolID := openapiPostKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyGenerate := cryptoutilModel.KeyGenerate(*openapiPostKeypoolKeyPoolIDKeyRequestObject.Body)
	return s.service.PostKeyPoolKeyPoolIDKey(ctx, &keyPoolID, &keyGenerate)
}
