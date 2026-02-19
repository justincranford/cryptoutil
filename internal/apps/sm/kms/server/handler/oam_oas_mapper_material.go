// Copyright (c) 2025 Justin Cranford
//
//

package handler

import (
	"errors"
	"fmt"
	http "net/http"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// OamOasMapper maps between OpenAPI Model and OpenAPI Server types.
func (m *OamOasMapper) toOasGetElastickeysResponse(err error, elasticKeys []cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.GetElastickeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElastickeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElastickeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to get ElasticKeys: %w", err)
	}

	return cryptoutilKmsServer.GetElastickeys200JSONResponse(elasticKeys), err
}

func (m *OamOasMapper) toOamGetMaterialKeysQueryParams(openapiParams *cryptoutilKmsServer.GetMaterialkeysParams) *cryptoutilOpenapiModel.MaterialKeysQueryParams {
	// Convert ElasticKeyIds (server) to ElasticKeyID (model) with type conversion.
	var elasticKeyIDs *[]cryptoutilOpenapiModel.ElasticKeyID

	if openapiParams.ElasticKeyIds != nil {
		ids := make([]cryptoutilOpenapiModel.ElasticKeyID, len(*openapiParams.ElasticKeyIds))
		for i, id := range *openapiParams.ElasticKeyIds {
			ids[i] = cryptoutilOpenapiModel.ElasticKeyID(id)
		}

		elasticKeyIDs = &ids
	}

	// Convert MaterialKeyIds (server) to MaterialKeyID (model) with type conversion.
	var materialKeyIDs *[]cryptoutilOpenapiModel.MaterialKeyID

	if openapiParams.MaterialKeyIds != nil {
		ids := make([]cryptoutilOpenapiModel.MaterialKeyID, len(*openapiParams.MaterialKeyIds))
		for i, id := range *openapiParams.MaterialKeyIds {
			ids[i] = cryptoutilOpenapiModel.MaterialKeyID(id)
		}

		materialKeyIDs = &ids
	}

	// Convert date filters with type casts.
	var minGenerateDate *cryptoutilOpenapiModel.MaterialKeyGenerateDate

	if openapiParams.MinGenerateDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyGenerateDate(*openapiParams.MinGenerateDate)
		minGenerateDate = &d
	}

	var maxGenerateDate *cryptoutilOpenapiModel.MaterialKeyGenerateDate

	if openapiParams.MaxGenerateDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyGenerateDate(*openapiParams.MaxGenerateDate)
		maxGenerateDate = &d
	}

	var minImportDate *cryptoutilOpenapiModel.MaterialKeyImportDate

	if openapiParams.MinImportDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyImportDate(*openapiParams.MinImportDate)
		minImportDate = &d
	}

	var maxImportDate *cryptoutilOpenapiModel.MaterialKeyImportDate

	if openapiParams.MaxImportDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyImportDate(*openapiParams.MaxImportDate)
		maxImportDate = &d
	}

	var minExpirationDate *cryptoutilOpenapiModel.MaterialKeyExpirationDate

	if openapiParams.MinExpirationDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyExpirationDate(*openapiParams.MinExpirationDate)
		minExpirationDate = &d
	}

	var maxExpirationDate *cryptoutilOpenapiModel.MaterialKeyExpirationDate

	if openapiParams.MaxExpirationDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyExpirationDate(*openapiParams.MaxExpirationDate)
		maxExpirationDate = &d
	}

	var minRevocationDate *cryptoutilOpenapiModel.MaterialKeyRevocationDate

	if openapiParams.MinRevocationDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyRevocationDate(*openapiParams.MinRevocationDate)
		minRevocationDate = &d
	}

	var maxRevocationDate *cryptoutilOpenapiModel.MaterialKeyRevocationDate

	if openapiParams.MaxRevocationDate != nil {
		d := cryptoutilOpenapiModel.MaterialKeyRevocationDate(*openapiParams.MaxRevocationDate)
		maxRevocationDate = &d
	}

	// Convert PageNumber (server) to Page (model).
	var page *cryptoutilOpenapiModel.PageNumber

	if openapiParams.PageNumber != nil {
		p := cryptoutilOpenapiModel.PageNumber(*openapiParams.PageNumber)
		page = &p
	}

	// Convert PageSize (server) to Size (model).
	var size *cryptoutilOpenapiModel.PageSize

	if openapiParams.PageSize != nil {
		s := cryptoutilOpenapiModel.PageSize(*openapiParams.PageSize)
		size = &s
	}

	filters := cryptoutilOpenapiModel.MaterialKeysQueryParams{
		ElasticKeyID:      elasticKeyIDs,
		MaterialKeyID:     materialKeyIDs,
		MinGenerateDate:   minGenerateDate,
		MaxGenerateDate:   maxGenerateDate,
		MinImportDate:     minImportDate,
		MaxImportDate:     maxImportDate,
		MinExpirationDate: minExpirationDate,
		MaxExpirationDate: maxExpirationDate,
		MinRevocationDate: minRevocationDate,
		MaxRevocationDate: maxRevocationDate,
		Page:              page,
		Size:              size,
		Sort:              nil, // Sort not in KMS server params.
	}

	return &filters
}

func (m *OamOasMapper) toOasGetMaterialKeysResponse(err error, keys []cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.GetMaterialkeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetMaterialkeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetMaterialkeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetMaterialkeys200JSONResponse(keys), err
}

func (m *OamOasMapper) toOasPutElastickeyElasticKeyIDResponse(err error, updatedElasticKey *cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.PutElastickeyElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PutElastickeyElasticKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PutElastickeyElasticKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusConflict:
				return cryptoutilKmsServer.PutElastickeyElasticKeyID409JSONResponse{ConflictJSONResponse: m.toOasHTTP409Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PutElastickeyElasticKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to update ElasticKey: %w", err)
	}

	return cryptoutilKmsServer.PutElastickeyElasticKeyID200JSONResponse(*updatedElasticKey), nil
}

func (m *OamOasMapper) toOasDeleteElastickeyElasticKeyIDResponse(err error) (cryptoutilKmsServer.DeleteElastickeyElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.DeleteElastickeyElasticKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.DeleteElastickeyElasticKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.DeleteElastickeyElasticKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to delete ElasticKey: %w", err)
	}

	return cryptoutilKmsServer.DeleteElastickeyElasticKeyID204Response{}, nil
}

func (m *OamOasMapper) toOasPostElastickeyElasticKeyIDImportResponse(err error, importedMaterialKey *cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.PostElastickeyElasticKeyIDImportResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDImport400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDImport404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDImport500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to import MaterialKey: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDImport200JSONResponse(*importedMaterialKey), nil
}

func (m *OamOasMapper) toOasPostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponse(err error) (cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevokeResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to revoke MaterialKey: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyMaterialKeyIDRevoke204Response{}, nil
}

func (m *OamOasMapper) toOasDeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err error) (cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to delete MaterialKey: %w", err)
	}

	return cryptoutilKmsServer.DeleteElastickeyElasticKeyIDMaterialkeyMaterialKeyID204Response{}, nil
}

// Helper methods

func (m *OamOasMapper) toOasHTTP400Response(appErr *cryptoutilSharedApperr.Error) cryptoutilKmsServer.BadRequestJSONResponse {
	return cryptoutilKmsServer.BadRequestJSONResponse(m.toOasHTTPErrorResponse(appErr))
}

func (m *OamOasMapper) toOasHTTP404Response(appErr *cryptoutilSharedApperr.Error) cryptoutilKmsServer.NotFoundJSONResponse {
	return cryptoutilKmsServer.NotFoundJSONResponse(m.toOasHTTPErrorResponse(appErr))
}

func (m *OamOasMapper) toOasHTTP409Response(appErr *cryptoutilSharedApperr.Error) cryptoutilKmsServer.ConflictJSONResponse {
	return cryptoutilKmsServer.ConflictJSONResponse(m.toOasHTTPErrorResponse(appErr))
}

func (m *OamOasMapper) toOasHTTP500Response(appErr *cryptoutilSharedApperr.Error) cryptoutilKmsServer.InternalServerErrorJSONResponse {
	return cryptoutilKmsServer.InternalServerErrorJSONResponse(m.toOasHTTPErrorResponse(appErr))
}

func (*OamOasMapper) toOasHTTPErrorResponse(appErr *cryptoutilSharedApperr.Error) cryptoutilKmsServer.Error {
	return cryptoutilKmsServer.Error{
		Code:    string(appErr.HTTPStatusLineAndCode.StatusLine.ReasonPhrase),
		Message: appErr.Error(),
	}
}
