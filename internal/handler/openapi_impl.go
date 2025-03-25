package handler

import (
	"context"

	cryptoutilModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/openapi/server"
	cryptoutilService "cryptoutil/internal/service"
)

// StrictServer implements cryptoutilServer.StrictServerInterface
type StrictServer struct {
	service *cryptoutilService.KEKPoolService
}

func NewStrictServer(service *cryptoutilService.KEKPoolService) *StrictServer {
	return &StrictServer{service: service}
}

func (s *StrictServer) GetKekpool(ctx context.Context, _ cryptoutilServer.GetKekpoolRequestObject) (cryptoutilServer.GetKekpoolResponseObject, error) {
	return s.service.GetKEKPool(ctx)
}

func (s *StrictServer) PostKekpool(ctx context.Context, openapiPostKekpoolRequestObject cryptoutilServer.PostKekpoolRequestObject) (cryptoutilServer.PostKekpoolResponseObject, error) {
	kekPoolCreate := cryptoutilModel.KEKPoolCreate(*openapiPostKekpoolRequestObject.Body)
	return s.service.PostKEKPool(ctx, &kekPoolCreate)
}

func (s *StrictServer) GetKekpoolKekPoolIDKek(ctx context.Context, openapiGetKekpoolKekPoolIDKekRequestObject cryptoutilServer.GetKekpoolKekPoolIDKekRequestObject) (cryptoutilServer.GetKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID := openapiGetKekpoolKekPoolIDKekRequestObject.KekPoolID
	return s.service.GetKEKPoolKEKPoolIDKEK(ctx, &kekPoolID)
}

func (s *StrictServer) PostKekpoolKekPoolIDKek(ctx context.Context, openapiPostKekpoolKekPoolIDKekRequestObject cryptoutilServer.PostKekpoolKekPoolIDKekRequestObject) (cryptoutilServer.PostKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID := openapiPostKekpoolKekPoolIDKekRequestObject.KekPoolID
	kekGenerate := cryptoutilModel.KEKGenerate(*openapiPostKekpoolKekPoolIDKekRequestObject.Body)
	return s.service.PostKEKPoolKEKPoolIDKEK(ctx, &kekPoolID, &kekGenerate)
}
