package handlers

import (
	"context"
	"cryptoutil/api/openapi"
	"cryptoutil/database"
)

// StrictServer implements openapi.StrictServerInterface
type StrictServer struct {
	dbService *database.Service
}

func NewStrictServer(dbService *database.Service) *StrictServer {
	return &StrictServer{dbService: dbService}
}

// PostKek implements openapi.StrictServerInterface.PostKeK to handle POST /kek requests.
func (s *StrictServer) PostKek(_ context.Context, request openapi.PostKekRequestObject) (openapi.PostKekResponseObject, error) {
	kekCreate := *request.Body
	kek := openapi.KEK{
		Id:       new(openapi.KEKId),
		Name:     &kekCreate.Name,
		Provider: new(openapi.KEKProvider),
		Status:   new(openapi.KEKStatus),
	}
	return openapi.PostKek200JSONResponse(kek), nil
}

// GetKek implements openapi.StrictServerInterface.GetKeK to handle GET /kek requests.
func (s *StrictServer) GetKek(_ context.Context, request openapi.GetKekRequestObject) (openapi.GetKekResponseObject, error) {
	keks := []openapi.KEK{
		{
			Id:       new(openapi.KEKId),
			Name:     new(openapi.KEKName),
			Provider: new(openapi.KEKProvider),
			Status:   new(openapi.KEKStatus),
		},
	}
	return openapi.GetKek200JSONResponse(keks), nil
}
