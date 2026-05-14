// Copyright (c) 2025-2026 Justin Cranford.
//
//

package handler

import (
	"errors"
	"fmt"
	http "net/http"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// OamOasMapper maps between OpenAPI Model and OpenAPI Server types.
func (m *OamOasMapper) toOasGetElasticKeysResponse(err error, elasticKeys []cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.GetElasticKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElasticKeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElasticKeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to get ElasticKeys: %w", err)
	}

	return cryptoutilKmsServer.GetElasticKeys200JSONResponse(elasticKeys), err
}

func (m *OamOasMapper) toOamGetMaterialKeysQueryParams(openapiParams *cryptoutilKmsServer.GetMaterialKeysParams) *cryptoutilOpenapiModel.MaterialKeysQueryParams {
	// Convert ElasticKeyIds (server) to ElasticKeyID (model) with type conversion.
	var elasticKeyIDs *[]cryptoutilOpenapiModel.ElasticKeyID

	if openapiParams.ElasticKeyIDS != nil {
		ids := make([]cryptoutilOpenapiModel.ElasticKeyID, len(*openapiParams.ElasticKeyIDS))
		for i, id := range *openapiParams.ElasticKeyIDS {
			ids[i] = cryptoutilOpenapiModel.ElasticKeyID(id)
		}

		elasticKeyIDs = &ids
	}

	// Convert MaterialKeyIds (server) to MaterialKeyID (model) with type conversion.
	var materialKeyIDs *[]cryptoutilOpenapiModel.MaterialKeyID

	if openapiParams.MaterialKeyIDS != nil {
		ids := make([]cryptoutilOpenapiModel.MaterialKeyID, len(*openapiParams.MaterialKeyIDS))
		for i, id := range *openapiParams.MaterialKeyIDS {
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

func (m *OamOasMapper) toOasGetMaterialKeysResponse(err error, keys []cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.GetMaterialKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetMaterialKeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetMaterialKeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetMaterialKeys200JSONResponse(keys), err
}

func (m *OamOasMapper) toOasPutElasticKeysElasticKeyIDResponse(err error, updatedElasticKey *cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.PutElasticKeysElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PutElasticKeysElasticKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PutElasticKeysElasticKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusConflict:
				return cryptoutilKmsServer.PutElasticKeysElasticKeyID409JSONResponse{ConflictJSONResponse: m.toOasHTTP409Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PutElasticKeysElasticKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to update ElasticKey: %w", err)
	}

	return cryptoutilKmsServer.PutElasticKeysElasticKeyID200JSONResponse(*updatedElasticKey), nil
}

func (m *OamOasMapper) toOasDeleteElasticKeysElasticKeyIDResponse(err error) (cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.DeleteElasticKeysElasticKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.DeleteElasticKeysElasticKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.DeleteElasticKeysElasticKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to delete ElasticKey: %w", err)
	}

	return cryptoutilKmsServer.DeleteElasticKeysElasticKeyID204Response{}, nil
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDImportResponse(err error, importedMaterialKey *cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDImportResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDImport400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDImport404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDImport500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to import MaterialKey: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDImport200JSONResponse(*importedMaterialKey), nil
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevokeResponse(err error) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevokeResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevoke400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevoke404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevoke500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to revoke MaterialKey: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDRevoke204Response{}, nil
}

func (m *OamOasMapper) toOasDeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponse(err error) (cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to delete MaterialKey: %w", err)
	}

	return cryptoutilKmsServer.DeleteElasticKeysElasticKeyIDMaterialKeysMaterialKeyID204Response{}, nil
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
