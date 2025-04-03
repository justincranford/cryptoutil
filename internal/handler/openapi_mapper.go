package handler

import (
	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	"errors"
	"fmt"
	"net/http"
)

type openapiMapper struct{}

func NewMapper() *openapiMapper {
	return &openapiMapper{}
}

func (m *openapiMapper) toPostKeyResponse(err error, addedKeyPool *cryptoutilServiceModel.KeyPool) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypool400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypool404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to add KeyPool: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypool200JSONResponse(*addedKeyPool), nil
}

func (m *openapiMapper) toServiceModelGetKeyPoolQueryParams(openapiGetKeyPoolQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolsParams) *cryptoutilServiceModel.KeyPoolsQueryParams {
	filters := cryptoutilServiceModel.KeyPoolsQueryParams{
		Id:                openapiGetKeyPoolQueryParamsObject.Id,
		Name:              openapiGetKeyPoolQueryParamsObject.Name,
		Provider:          openapiGetKeyPoolQueryParamsObject.Provider,
		Algorithm:         openapiGetKeyPoolQueryParamsObject.Algorithm,
		VersioningAllowed: openapiGetKeyPoolQueryParamsObject.VersioningAllowed,
		ImportAllowed:     openapiGetKeyPoolQueryParamsObject.ImportAllowed,
		ExportAllowed:     openapiGetKeyPoolQueryParamsObject.ExportAllowed,
		Status:            openapiGetKeyPoolQueryParamsObject.Status,
		Sort:              openapiGetKeyPoolQueryParamsObject.Sort,
		Page:              openapiGetKeyPoolQueryParamsObject.Page,
		Size:              openapiGetKeyPoolQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toGetKeypoolsResponse(err error, keyPools []cryptoutilServiceModel.KeyPool) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypools400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypools404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypools500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get KeyPools: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypools200JSONResponse(keyPools), err
}

func (m *openapiMapper) toGetKeypoolKeyPoolIDResponse(err error, keyPool *cryptoutilServiceModel.KeyPool) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolID400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolID404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolID500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolID200JSONResponse(*keyPool), err
}

func (m *openapiMapper) toPostKeypoolKeyPoolIDKeyResponse(err error, generateKeyInKeyPoolResponse *cryptoutilServiceModel.Key) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to generate Key by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*generateKeyInKeyPoolResponse), err
}

func (m *openapiMapper) toGetKeypoolKeyPoolIDKeyKeyIDResponse(err error, key *cryptoutilServiceModel.Key) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get Keys by KeyPoolID and KeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID200JSONResponse(*key), err
}

func (m *openapiMapper) toServiceModelGetKeyPoolKeysQueryParams(openapiGetKeyQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysParams) *cryptoutilServiceModel.KeyPoolKeysQueryParams {
	filters := cryptoutilServiceModel.KeyPoolKeysQueryParams{
		Id:   openapiGetKeyQueryParamsObject.Id,
		Sort: openapiGetKeyQueryParamsObject.Sort,
		Page: openapiGetKeyQueryParamsObject.Page,
		Size: openapiGetKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toGetKeypoolKeyPoolIDKeysResponse(err error, keys []cryptoutilServiceModel.Key) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys200JSONResponse(keys), err
}

func (m *openapiMapper) toServiceModelGetKeysQueryParams(openapiGetKeyQueryParamsObject *cryptoutilOpenapiServer.GetKeysParams) *cryptoutilServiceModel.KeysQueryParams {
	filters := cryptoutilServiceModel.KeysQueryParams{
		Pool: openapiGetKeyQueryParamsObject.Pool,
		Id:   openapiGetKeyQueryParamsObject.Id,
		Sort: openapiGetKeyQueryParamsObject.Sort,
		Page: openapiGetKeyQueryParamsObject.Page,
		Size: openapiGetKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiMapper) toGetKeysResponse(err error, keys []cryptoutilServiceModel.Key) (cryptoutilOpenapiServer.GetKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeys400JSONResponse{HTTP400BadRequest: m.http400(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeys404JSONResponse{HTTP404NotFound: m.http404(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeys500JSONResponse{HTTP500InternalServerError: m.http500(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeys200JSONResponse(keys), err
}

func (m *openapiMapper) http400(appErr *cryptoutilAppErr.Error) cryptoutilServiceModel.HTTP400BadRequest {
	return cryptoutilServiceModel.HTTP400BadRequest{
		Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
		Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
	}
}

func (m *openapiMapper) http404(appErr *cryptoutilAppErr.Error) cryptoutilServiceModel.HTTP404NotFound {
	return cryptoutilServiceModel.HTTP404NotFound{
		Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
		Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
	}
}

func (m *openapiMapper) http500(appErr *cryptoutilAppErr.Error) cryptoutilServiceModel.HTTP500InternalServerError {
	return cryptoutilServiceModel.HTTP500InternalServerError{
		Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
		Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
	}
}
