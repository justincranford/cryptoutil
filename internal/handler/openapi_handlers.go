package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"cryptoutil/internal/apperr"
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilServiceLogic "cryptoutil/internal/servicelogic"
)

// StrictServer implements cryptoutilOpenapiServer.StrictServerInterface
type StrictServer struct {
	businessLogicService *cryptoutilServiceLogic.KeyPoolService
	openapiMapper        *openapiMapper
}

func NewOpenapiHandler(service *cryptoutilServiceLogic.KeyPoolService) *StrictServer {
	return &StrictServer{businessLogicService: service, openapiMapper: &openapiMapper{}}
}

func (s *StrictServer) PostKeypool(ctx context.Context, openapiPostKeypoolRequestObject cryptoutilOpenapiServer.PostKeypoolRequestObject) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	keyPoolCreate := cryptoutilServiceModel.KeyPoolCreate(*openapiPostKeypoolRequestObject.Body)
	addedKeyPool, err := s.businessLogicService.AddKeyPool(ctx, &keyPoolCreate)
	if err != nil {
		var appErr *apperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypool400JSONResponse{
					HTTP400BadRequest: cryptoutilServiceModel.HTTP400BadRequest{
						Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
						Message: appErr.Error(),
						Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
					},
				}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypool404JSONResponse{
					HTTP404NotFound: cryptoutilServiceModel.HTTP404NotFound{
						Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
						Message: appErr.Error(),
						Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
					},
				}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypool500JSONResponse{
					HTTP500InternalServerError: cryptoutilServiceModel.HTTP500InternalServerError{
						Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
						Message: appErr.Error(),
						Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
					},
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to add KeyPool: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypool200JSONResponse(*addedKeyPool), nil
}

func (s *StrictServer) GetKeypoolKeyPoolID(ctx context.Context, openapiGetKeypoolKeyPoolIDRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDRequestObject.KeyPoolID
	keyPool, err := s.businessLogicService.GetKeyPoolByKeyPoolID(ctx, keyPoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolID200JSONResponse(*keyPool), err
}

func (s *StrictServer) GetKeypools(ctx context.Context, openapiGetKeypoolRequestObject cryptoutilOpenapiServer.GetKeypoolsRequestObject) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	keyPoolsQueryParams := s.openapiMapper.toServiceModelGetKeyPoolQueryParams(&openapiGetKeypoolRequestObject.Params)
	keyPools, err := s.businessLogicService.GetKeyPools(ctx, keyPoolsQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get KeyPools: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypools200JSONResponse(keyPools), err
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

func (s *StrictServer) GetKeypoolKeyPoolIDKeyKeyID(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyKeyIDRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyIDRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyIDResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyKeyIDRequestObject.KeyPoolID
	keyID := openapiGetKeypoolKeyPoolIDKeyKeyIDRequestObject.KeyID
	key, err := s.businessLogicService.GetKeyByKeyPoolAndKeyID(ctx, keyPoolID, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Keys by KeyPoolID and KeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID200JSONResponse(*key), err
}

func (s *StrictServer) GetKeypoolKeyPoolIDKeys(ctx context.Context, openapiGetKeypoolKeyPoolIDKeyRequestObject cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysRequestObject) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	keyPoolID := openapiGetKeypoolKeyPoolIDKeyRequestObject.KeyPoolID
	keyPoolKeysQueryParams := s.openapiMapper.toServiceModelGetKeyPoolKeysQueryParams(&openapiGetKeypoolKeyPoolIDKeyRequestObject.Params)
	keys, err := s.businessLogicService.GetKeysByKeyPool(ctx, keyPoolID, keyPoolKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys200JSONResponse(keys), err
}
func (s *StrictServer) GetKeys(ctx context.Context, openapiGetKeysRequestObject cryptoutilOpenapiServer.GetKeysRequestObject) (cryptoutilOpenapiServer.GetKeysResponseObject, error) {
	keysQueryParams := s.openapiMapper.toServiceModelGetKeysQueryParams(&openapiGetKeysRequestObject.Params)
	keys, err := s.businessLogicService.GetKeys(ctx, keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeys200JSONResponse(keys), err
}
