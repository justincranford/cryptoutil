package handler

import (
	"errors"
	"fmt"
	"net/http"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
)

type oamOasMapper struct{}

func NewOasOamMapper() *oamOasMapper {
	return &oamOasMapper{}
}

func (m *oamOasMapper) toOasPostKeyResponse(err error, addedElasticKey *cryptoutilOpenapiModel.ElasticKey) (cryptoutilOpenapiServer.PostElastickeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickey400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickey404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickey500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to add ElasticKey: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickey200JSONResponse(*addedElasticKey), nil
}

func (m *oamOasMapper) toOasGetElastickeyElasticKeyIDResponse(err error, elasticKey *cryptoutilOpenapiModel.ElasticKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyID400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyID404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyID500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
	}

	return cryptoutilOpenapiServer.GetElastickeyElasticKeyID200JSONResponse(*elasticKey), err
}

func (m *oamOasMapper) toOasPostDecryptResponse(err error, decryptedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDDecrypt200TextResponse(decryptedBytes), err
}

func (m *oamOasMapper) toOamPostGenerateQueryParams(openapiParams *cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateParams) *cryptoutilOpenapiModel.GenerateParams {
	generateParams := cryptoutilOpenapiModel.GenerateParams{
		Context: openapiParams.Context,
		Alg:     openapiParams.Alg,
	}

	return &generateParams
}

func (m *oamOasMapper) toOasPostGenerateResponse(err error, encryptedNonPublicJWKBytes, clearPublicJWKBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerateResponseObject, error) {
	// clearPublicJWKBytes is intentionally unused in current implementation
	// but kept for potential future logging/debugging purposes
	_ = clearPublicJWKBytes

	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDGenerate200TextResponse(encryptedNonPublicJWKBytes), err
}

func (m *oamOasMapper) toOamPostEncryptQueryParams(openapiParams *cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptParams) *cryptoutilOpenapiModel.EncryptParams {
	encryptParams := cryptoutilOpenapiModel.EncryptParams{
		Context: openapiParams.Context,
	}

	return &encryptParams
}

func (m *oamOasMapper) toOasPostEncryptResponse(err error, encryptedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDEncrypt200TextResponse(encryptedBytes), err
}

func (m *oamOasMapper) toOasPostElastickeyElasticKeyIDMaterialkeyResponse(err error, generateKeyInElasticKeyResponse *cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to generate Key by ElasticKeyID: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDMaterialkey200JSONResponse(*generateKeyInElasticKeyResponse), err
}

func (m *oamOasMapper) toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err error, key *cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID200JSONResponse(*key), err
}

func (m *oamOasMapper) toOamGetElasticKeyMaterialKeysQueryParams(openapiParams *cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysParams) *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams {
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

func (m *oamOasMapper) toOasGetElastickeyElasticKeyIDMaterialkeysResponse(err error, keys []cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilOpenapiServer.GetElastickeyElasticKeyIDMaterialkeys200JSONResponse(keys), err
}

func (m *oamOasMapper) toOasPostSignResponse(err error, encryptedBytes []byte) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDSign200TextResponse(encryptedBytes), err
}

func (m *oamOasMapper) toOasPostVerifyResponse(err error) (cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to verify: %w", err)
	}

	return cryptoutilOpenapiServer.PostElastickeyElasticKeyIDVerify204Response{}, err
}

func (m *oamOasMapper) toOamGetElasticKeyQueryParams(openapiParams *cryptoutilOpenapiServer.GetElastickeysParams) *cryptoutilOpenapiModel.ElasticKeysQueryParams {
	filters := cryptoutilOpenapiModel.ElasticKeysQueryParams{
		ElasticKeyID:      openapiParams.ElasticKeyID,
		Name:              openapiParams.Name,
		Provider:          openapiParams.Provider,
		Algorithm:         openapiParams.Algorithm,
		VersioningAllowed: openapiParams.VersioningAllowed,
		ImportAllowed:     openapiParams.ImportAllowed,
		Status:            openapiParams.Status,
		Sort:              openapiParams.Sort,
		Page:              openapiParams.Page,
		Size:              openapiParams.Size,
	}

	return &filters
}

func (m *oamOasMapper) toOasGetElastickeysResponse(err error, elasticKeys []cryptoutilOpenapiModel.ElasticKey) (cryptoutilOpenapiServer.GetElastickeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetElastickeys400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetElastickeys404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetElastickeys500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to get ElasticKeys: %w", err)
	}

	return cryptoutilOpenapiServer.GetElastickeys200JSONResponse(elasticKeys), err
}

func (m *oamOasMapper) toOamGetMaterialKeysQueryParams(openapiParams *cryptoutilOpenapiServer.GetMaterialkeysParams) *cryptoutilOpenapiModel.MaterialKeysQueryParams {
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

func (m *oamOasMapper) toOasGetMaterialKeysResponse(err error, keys []cryptoutilOpenapiModel.MaterialKey) (cryptoutilOpenapiServer.GetMaterialkeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilAppErr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilOpenapiServer.GetMaterialkeys400JSONResponse{HTTP400BadRequest: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilOpenapiServer.GetMaterialkeys404JSONResponse{HTTP404NotFound: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilOpenapiServer.GetMaterialkeys500JSONResponse{HTTP500InternalServerError: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilOpenapiServer.GetMaterialkeys200JSONResponse(keys), err
}

// Helper methods

func (m *oamOasMapper) toOasHTTP400Response(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTP400BadRequest {
	return cryptoutilOpenapiModel.HTTP400BadRequest(m.toOasHTTPErrorResponse(appErr))
}

func (m *oamOasMapper) toOasHTTP404Response(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTP404NotFound {
	return cryptoutilOpenapiModel.HTTP404NotFound(m.toOasHTTPErrorResponse(appErr))
}

func (m *oamOasMapper) toOasHTTP500Response(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTP500InternalServerError {
	return cryptoutilOpenapiModel.HTTP500InternalServerError(m.toOasHTTPErrorResponse(appErr))
}

func (*oamOasMapper) toOasHTTPErrorResponse(appErr *cryptoutilAppErr.Error) cryptoutilOpenapiModel.HTTPError {
	return cryptoutilOpenapiModel.HTTPError{
		Error:   string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
		Status:  int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode),
	}
}
