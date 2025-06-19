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

func (m *openapiBusinessLogicMapper) toPostKeyResponse(err error, addedElasticKey *cryptoutilBusinessLogicModel.ElasticKey) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickey400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickey404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickey500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to add ElasticKey: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickey200JSONResponse(*addedElasticKey), nil
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetElasticKeyQueryParams(openapiGetElasticKeyQueryParamsObject *cryptoutilOpenapiServer.GetElastickeysParams) *cryptoutilBusinessLogicModel.ElasticKeysQueryParams {
	filters := cryptoutilBusinessLogicModel.ElasticKeysQueryParams{
		Id:                openapiGetElasticKeyQueryParamsObject.Id,
		Name:              openapiGetElasticKeyQueryParamsObject.Name,
		Provider:          openapiGetElasticKeyQueryParamsObject.Provider,
		Algorithm:         openapiGetElasticKeyQueryParamsObject.Algorithm,
		VersioningAllowed: openapiGetElasticKeyQueryParamsObject.VersioningAllowed,
		ImportAllowed:     openapiGetElasticKeyQueryParamsObject.ImportAllowed,
		ExportAllowed:     openapiGetElasticKeyQueryParamsObject.ExportAllowed,
		Status:            openapiGetElasticKeyQueryParamsObject.Status,
		Sort:              openapiGetElasticKeyQueryParamsObject.Sort,
		Page:              openapiGetElasticKeyQueryParamsObject.Page,
		Size:              openapiGetElasticKeyQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetElastickeysResponse(err error, elasticKeys []cryptoutilBusinessLogicModel.ElasticKey) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeys400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeys404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeys500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get ElasticKeys: %w", err)
	}
	return cryptoutilOpenapiServer.GetElastickeys200JSONResponse(elasticKeys), err
}

func (m *openapiBusinessLogicMapper) toGetElastickeyElasticKeyIDResponse(err error, elasticKey *cryptoutilBusinessLogicModel.ElasticKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyID400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyID404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyID500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetElastickeyElasticKeyID200JSONResponse(*elasticKey), err
}

func (m *openapiBusinessLogicMapper) toPostElastickeyElasticKeyIDKeyResponse(err error, generateKeyInElasticKeyResponse *cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKey400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKey404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKey500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to generate Key by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDKey200JSONResponse(*generateKeyInElasticKeyResponse), err
}

func (m *openapiBusinessLogicMapper) toGetElastickeyElasticKeyIDKeyKeyIDResponse(err error, key *cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyID400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyID404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyID500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to get Keys by ElasticKeyID and KeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeyKeyID200JSONResponse(*key), err
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetElasticKeyKeysQueryParams(openapiGetElasticKeyKeysQueryParamsObject *cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeysParams) *cryptoutilBusinessLogicModel.ElasticKeyKeysQueryParams {
	filters := cryptoutilBusinessLogicModel.ElasticKeyKeysQueryParams{
		Id:              openapiGetElasticKeyKeysQueryParamsObject.Id,
		MinGenerateDate: openapiGetElasticKeyKeysQueryParamsObject.MinGenerateDate,
		MaxGenerateDate: openapiGetElasticKeyKeysQueryParamsObject.MaxGenerateDate,
		Sort:            openapiGetElasticKeyKeysQueryParamsObject.Sort,
		Page:            openapiGetElasticKeyKeysQueryParamsObject.Page,
		Size:            openapiGetElasticKeyKeysQueryParamsObject.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetElastickeyElasticKeyIDKeysResponse(err error, keys []cryptoutilBusinessLogicModel.Key) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeys400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeys404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeys500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDKeys200JSONResponse(keys), err
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
		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetKeys200JSONResponse(keys), err
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelPostEncryptQueryParams(openapiPostElastickeyElasticKeyIDKeyKeyIDEncryptParamsObject *cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptParams) *cryptoutilBusinessLogicModel.EncryptParams {
	filters := cryptoutilBusinessLogicModel.EncryptParams{
		Context: openapiPostElastickeyElasticKeyIDKeyKeyIDEncryptParamsObject.Context,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toPostEncryptResponse(err error, encryptedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt200TextResponse(encryptedBytes), err
}

func (m *openapiBusinessLogicMapper) toPostDecryptResponse(err error, decryptedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt200TextResponse(decryptedBytes), err
}

func (m *openapiBusinessLogicMapper) toPostSignResponse(err error, encryptedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign200TextResponse(encryptedBytes), err
}

func (m *openapiBusinessLogicMapper) toPostVerifyResponse(err error, verifiedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to verify: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify204Response{}, err
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
