package handler

import (
	"context"

	"cryptoutil/internal/openapi/model"
	"cryptoutil/internal/openapi/server"
	"cryptoutil/service"
)

// StrictServer implements server.StrictServerInterface
type StrictServer struct {
	service *service.KEKPoolService
}

func NewStrictServer(service *service.KEKPoolService) *StrictServer {
	return &StrictServer{service: service}
}

func (s *StrictServer) GetKekpool(ctx context.Context, _ server.GetKekpoolRequestObject) (server.GetKekpoolResponseObject, error) {
	return s.service.GetKEKPool(ctx)
}

func (s *StrictServer) PostKekpool(ctx context.Context, openapiPostKekpoolRequestObject server.PostKekpoolRequestObject) (server.PostKekpoolResponseObject, error) {
	kekPoolCreate := model.KEKPoolCreate(*openapiPostKekpoolRequestObject.Body)
	return s.service.PostKEKPool(ctx, &kekPoolCreate)
}

func (s *StrictServer) GetKekpoolKekPoolIDKek(ctx context.Context, openapiGetKekpoolKekPoolIDKekRequestObject server.GetKekpoolKekPoolIDKekRequestObject) (server.GetKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID := openapiGetKekpoolKekPoolIDKekRequestObject.KekPoolID
	return s.service.GetKEKPoolKEKPoolIDKEK(ctx, &kekPoolID)
}

func (s *StrictServer) PostKekpoolKekPoolIDKek(ctx context.Context, openapiPostKekpoolKekPoolIDKekRequestObject server.PostKekpoolKekPoolIDKekRequestObject) (server.PostKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID := openapiPostKekpoolKekPoolIDKekRequestObject.KekPoolID
	kekGenerate := model.KEKGenerate(*openapiPostKekpoolKekPoolIDKekRequestObject.Body)
	return s.service.PostKEKPoolKEKPoolIDKEK(ctx, &kekPoolID, &kekGenerate)
}
