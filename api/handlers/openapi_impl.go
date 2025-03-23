package handlers

import (
	"context"

	"cryptoutil/api/openapi"
	"cryptoutil/service"
)

// StrictServer implements openapi.StrictServerInterface
type StrictServer struct {
	service *service.KEKPoolService
}

func NewStrictServer(service *service.KEKPoolService) *StrictServer {
	return &StrictServer{service: service}
}

func (s *StrictServer) GetKekpool(ctx context.Context, _ openapi.GetKekpoolRequestObject) (openapi.GetKekpoolResponseObject, error) {
	return s.service.GetKEKPool(ctx)
}

func (s *StrictServer) PostKekpool(ctx context.Context, openapiPostKekpoolRequestObject openapi.PostKekpoolRequestObject) (openapi.PostKekpoolResponseObject, error) {
	kekPoolCreate := openapi.KEKPoolCreate(*openapiPostKekpoolRequestObject.Body)
	return s.service.PostKEKPool(ctx, &kekPoolCreate)
}

func (s *StrictServer) GetKekpoolKekPoolIDKek(ctx context.Context, openapiGetKekpoolKekPoolIDKekRequestObject openapi.GetKekpoolKekPoolIDKekRequestObject) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID := openapiGetKekpoolKekPoolIDKekRequestObject.KekPoolID
	return s.service.GetKEKPoolKEKPoolIDKEK(ctx, &kekPoolID)
}

func (s *StrictServer) PostKekpoolKekPoolIDKek(ctx context.Context, openapiPostKekpoolKekPoolIDKekRequestObject openapi.PostKekpoolKekPoolIDKekRequestObject) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	kekPoolID := openapiPostKekpoolKekPoolIDKekRequestObject.KekPoolID
	return s.service.PostKEKPoolKEKPoolIDKEK(ctx, &kekPoolID)
}
