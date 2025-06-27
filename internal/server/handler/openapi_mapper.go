package handler

import (
	"errors"
	"fmt"
	"net/http"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
)

type openapiBusinessLogicMapper struct{}

func NewOpenapiBusinessLogicMapper() *openapiBusinessLogicMapper {
	return &openapiBusinessLogicMapper{}
}

func (m *openapiBusinessLogicMapper) toPostKeyResponse(err error, addedElasticKey *cryptoutilOpenapiModel.ElasticKey) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
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

func (m *openapiBusinessLogicMapper) toGetElastickeyElasticKeyIDResponse(err error, elasticKey *cryptoutilOpenapiModel.ElasticKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
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

func (m *openapiBusinessLogicMapper) toBusinessLogicModelPostEncryptQueryParams(openapiParams *cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptParams) *cryptoutilOpenapiModel.EncryptParams {
	filters := cryptoutilOpenapiModel.EncryptParams{
		Context: openapiParams.Context,
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

func (m *openapiBusinessLogicMapper) toPostElastickeyElasticKeyIDMaterialkeyResponse(err error, generateKeyInElasticKeyResponse *cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to generate Key by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey200JSONResponse(*generateKeyInElasticKeyResponse), err
}

func (m *openapiBusinessLogicMapper) toGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err error, key *cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID200JSONResponse(*key), err
}

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetElasticKeyMaterialKeysQueryParams(openapiParams *cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysParams) *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams {
	filters := cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
		MaterialKeyID:   openapiParams.MaterialKeyID,
		MinGenerateDate: openapiParams.MinGenerateDate,
		MaxGenerateDate: openapiParams.MaxGenerateDate,
		Sort:            openapiParams.Sort,
		Page:            openapiParams.Page,
		Size:            openapiParams.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetElastickeyElasticKeyIDMaterialkeysResponse(err error, keys []cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys200JSONResponse(keys), err
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

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetElasticKeyQueryParams(openapiParams *cryptoutilOpenapiServer.GetElastickeysParams) *cryptoutilOpenapiModel.ElasticKeysQueryParams {
	filters := cryptoutilOpenapiModel.ElasticKeysQueryParams{
		ElasticKeyID:      openapiParams.ElasticKeyID,
		Name:              openapiParams.Name,
		Provider:          openapiParams.Provider,
		Algorithm:         openapiParams.Algorithm,
		VersioningAllowed: openapiParams.VersioningAllowed,
		ImportAllowed:     openapiParams.ImportAllowed,
		ExportAllowed:     openapiParams.ExportAllowed,
		Status:            openapiParams.Status,
		Sort:              openapiParams.Sort,
		Page:              openapiParams.Page,
		Size:              openapiParams.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetElastickeysResponse(err error, elasticKeys []cryptoutilOpenapiModel.ElasticKey) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
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

func (m *openapiBusinessLogicMapper) toBusinessLogicModelGetMaterialKeysQueryParams(openapiParams *cryptoutilOpenapiServer.GetMaterialkeysParams) *cryptoutilOpenapiModel.MaterialKeysQueryParams {
	filters := cryptoutilOpenapiModel.MaterialKeysQueryParams{
		ElasticKeyID:    openapiParams.ElasticKeyID,
		MaterialKeyID:   openapiParams.MaterialKeyID,
		MinGenerateDate: openapiParams.MinGenerateDate,
		MaxGenerateDate: openapiParams.MaxGenerateDate,
		Sort:            openapiParams.Sort,
		Page:            openapiParams.Page,
		Size:            openapiParams.Size,
	}
	return &filters
}

func (m *openapiBusinessLogicMapper) toGetMaterialKeysResponse(err error, keys []cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.GetMaterialkeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetMaterialkeys400JSONResponse{HTTP400BadRequest: m.toHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetMaterialkeys404JSONResponse{HTTP404NotFound: m.toHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetMaterialkeys500JSONResponse{HTTP500InternalServerError: m.toHTTP500Response(appErr)}, nil
			}
		}
		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}
	return cryptoutilOpenapiServer.GetMaterialkeys200JSONResponse(keys), err
}

// Helper methods

func (m *openapiBusinessLogicMapper) toHTTP400Response(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTP400BadRequest {
	return cryptoutilOpenapiModel.HTTP400BadRequest(m.toHTTPErrorResponse(appErr))
}

func (m *openapiBusinessLogicMapper) toHTTP404Response(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTP404NotFound {
	return cryptoutilOpenapiModel.HTTP404NotFound(m.toHTTPErrorResponse(appErr))
}

func (m *openapiBusinessLogicMapper) toHTTP500Response(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTP500InternalServerError {
	return cryptoutilOpenapiModel.HTTP500InternalServerError(m.toHTTPErrorResponse(appErr))
}

func (*openapiBusinessLogicMapper) toHTTPErrorResponse(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTPError {
	return cryptoutilOpenapiModel.HTTPError{
		Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
		Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
	}
}
