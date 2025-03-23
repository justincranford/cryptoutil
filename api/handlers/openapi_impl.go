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

func (s *StrictServer) GetKekpool(ctx context.Context, request openapi.GetKekpoolRequestObject) (openapi.GetKekpoolResponseObject, error) {
	return s.service.GetKEKPool(ctx, request)
}

func (s *StrictServer) PostKekpool(ctx context.Context, request openapi.PostKekpoolRequestObject) (openapi.PostKekpoolResponseObject, error) {
	return s.service.PostKEKPool(ctx, request)
}

func (s *StrictServer) GetKekpoolKekPoolIDKek(ctx context.Context, request openapi.GetKekpoolKekPoolIDKekRequestObject) (openapi.GetKekpoolKekPoolIDKekResponseObject, error) {
	return s.service.GetKEKPoolKEKPoolIDKEK(ctx, request)
}

func (s *StrictServer) PostKekpoolKekPoolIDKek(ctx context.Context, request openapi.PostKekpoolKekPoolIDKekRequestObject) (openapi.PostKekpoolKekPoolIDKekResponseObject, error) {
	return s.service.PostKEKPoolKEKPoolIDKEK(ctx, request)
}
