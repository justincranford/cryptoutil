package handler

import (
	"errors"
	"fmt"
	"net/http"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
)

type openapiBusinessLogicMapper struct{}

func NewOpenapiBusinessLogicMapper() *openapiBusinessLogicMapper {
	return &openapiBusinessLogicMapper{}
}

func (m *openapiBusinessLogicMapper) toPostKeyResponse(err error, addedKeyPool *cryptoutilBusinessLogicModel.KeyPool) (cryptoutilOpenapiServer.PostKeypoolResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypool400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypool404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypool500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to add KeyPool: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypool200JSONResponse(*addedKeyPool), nil
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetKeyPoolQueryParams(openapiGetKeyPoolQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolsParams) *cryptoutilBusinessLogicModel.KeyPoolsQueryParams {
	filters := cryptoutilBusinessLogicModel.KeyPoolsQueryParams{
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

func (m *openapiBusinessLogicMapper) toGetKeypoolsResponse(err error, keyPools []cryptoutilBusinessLogicModel.KeyPool) (cryptoutilOpenapiServer.GetKeypoolsResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypools400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypools404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypools500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get KeyPools: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypools200JSONResponse(keyPools), err
}

func (m *openapiBusinessLogicMapper) toGetKeypoolKeyPoolIDResponse(err error, keyPool *cryptoutilBusinessLogicModel.KeyPool) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolID400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolID404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolID500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolID200JSONResponse(*keyPool), err
}

func (m *openapiBusinessLogicMapper) toPostKeypoolKeyPoolIDKeyResponse(err error, generateKeyInKeyPoolResponse *cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to generate Key by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDKey200JSONResponse(*generateKeyInKeyPoolResponse), err
}

func (m *openapiBusinessLogicMapper) toGetKeypoolKeyPoolIDKeyKeyIDResponse(err error, key *cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get Keys by KeyPoolID and KeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeyKeyID200JSONResponse(*key), err
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetKeyPoolKeysQueryParams(openapiGetKeyPoolKeysQueryParamsObject *cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysParams) *cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams {
	filters := cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams{
		Id:              openapiGetKeyPoolKeysQueryParamsObject.Id,
		MinGenerateDate: openapiGetKeyPoolKeysQueryParamsObject.MinGenerateDate,
		MaxGenerateDate: openapiGetKeyPoolKeysQueryParamsObject.MaxGenerateDate,
		Sort:            openapiGetKeyPoolKeysQueryParamsObject.Sort,
		Page:            openapiGetKeyPoolKeysQueryParamsObject.Page,
		Size:            openapiGetKeyPoolKeysQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetKeypoolKeyPoolIDKeysResponse(err error, keys []cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeypoolKeyPoolIDKeys200JSONResponse(keys), err
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetKeysQueryParams(openapiGetKeyQueryParamsObject *cryptoutilOpenapiServer.GetKeysParams) *cryptoutilBusinessLogicModel.KeysQueryParams {
	filters := cryptoutilBusinessLogicModel.KeysQueryParams{
		Pool:            openapiGetKeyQueryParamsObject.Pool,
		Id:              openapiGetKeyQueryParamsObject.Id,
		MinGenerateDate: openapiGetKeyQueryParamsObject.MinGenerateDate,
		MaxGenerateDate: openapiGetKeyQueryParamsObject.MaxGenerateDate,
		Sort:            openapiGetKeyQueryParamsObject.Sort,
		Page:            openapiGetKeyQueryParamsObject.Page,
		Size:            openapiGetKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetKeysResponse(err error, keys []cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.GetKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetKeys400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetKeys404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetKeys500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeys200JSONResponse(keys), err
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelPostEncryptQueryParams(openapiPostKeypoolKeyPoolIDKeyKeyIDEncryptParamsObject *cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncryptParams) *cryptoutilBusinessLogicModel.EncryptParams {
	filters := cryptoutilBusinessLogicModel.EncryptParams{
		Context: openapiPostKeypoolKeyPoolIDKeyKeyIDEncryptParamsObject.Context,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toPostEncryptResponse(err error, encryptedBytes []byte) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncrypt400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncrypt404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncrypt500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDEncrypt200TextResponse(encryptedBytes), err
}

func (m *openapiBusinessLogicMapper) toPostDecryptResponse(err error, decryptedBytes []byte) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecrypt400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecrypt404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecrypt500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDDecrypt200TextResponse(decryptedBytes), err
}

func (m *openapiBusinessLogicMapper) toPostSignResponse(err error, encryptedBytes []byte) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSignResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSign400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSign404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSign500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDSign200TextResponse(encryptedBytes), err
}

func (m *openapiBusinessLogicMapper) toPostVerifyResponse(err error) (cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerifyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerify400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerify404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerify500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to verify: %w", err)
	}
	return cryptoutilOpenapiServer.PostKeypoolKeyPoolIDVerify204Response{}, err
}

// Helper methods

func (m *openapiBusinessLogicMapper) toHTTP400Response(appErr *cryptoutilAppErr.Error) cryptoutilBusinessLogicModel.HTTP400BadRequest {
	return cryptoutilBusinessLogicModel.HTTP400BadRequest(m.toHTTPErrorResponse(appErr))
}

func (m *openapiBusinessLogicMapper) toHTTP404Response(appErr *cryptoutilAppErr.Error) cryptoutilBusinessLogicModel.HTTP404NotFound {
	return cryptoutilBusinessLogicModel.HTTP404NotFound(m.toHTTPErrorResponse(appErr))
}

func (m *openapiBusinessLogicMapper) toHTTP500Response(appErr *cryptoutilAppErr.Error) cryptoutilBusinessLogicModel.HTTP500InternalServerError {
	return cryptoutilBusinessLogicModel.HTTP500InternalServerError(m.toHTTPErrorResponse(appErr))
}

func (*openapiBusinessLogicMapper) toHTTPErrorResponse(appErr *cryptoutilAppErr.Error) cryptoutilBusinessLogicModel.HTTPError {
	return cryptoutilBusinessLogicModel.HTTPError{
		Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
		Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
	}
}
