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
type OamOasMapper struct{}

// NewOasOamMapper creates a new mapper between OpenAPI Model and OpenAPI Server types.
func NewOasOamMapper() *OamOasMapper {
	return &OamOasMapper{}
}

func (m *OamOasMapper) toOasPostKeyResponse(err error, addedElasticKey *cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.PostElastickeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickey400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickey500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to add ElasticKey: %w", err)
	}

	return cryptoutilKmsServer.PostElastickey200JSONResponse(*addedElasticKey), nil
}

func (m *OamOasMapper) toOasGetElastickeyElasticKeyIDResponse(err error, elasticKey *cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.GetElastickeyElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElastickeyElasticKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.GetElastickeyElasticKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElastickeyElasticKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetElastickeyElasticKeyID200JSONResponse(*elasticKey), err
}

func (m *OamOasMapper) toOasPostDecryptResponse(err error, decryptedBytes []byte) (cryptoutilKmsServer.PostElastickeyElasticKeyIDDecryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDDecrypt200TextResponse(decryptedBytes), err
}

func (m *OamOasMapper) toOamPostGenerateQueryParams(openapiParams *cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerateParams) *cryptoutilOpenapiModel.GenerateParams {
	// Convert Alg (server *string) to model *GenerateAlgorithm.
	var alg *cryptoutilOpenapiModel.GenerateAlgorithm

	if openapiParams.Alg != nil {
		a := cryptoutilOpenapiModel.GenerateAlgorithm(*openapiParams.Alg)
		alg = &a
	}

	// Convert Context (server *string) to model *EncryptContext.
	var context *cryptoutilOpenapiModel.EncryptContext

	if openapiParams.Context != nil {
		c := cryptoutilOpenapiModel.EncryptContext(*openapiParams.Context)
		context = &c
	}

	generateParams := cryptoutilOpenapiModel.GenerateParams{
		Context: context,
		Alg:     alg,
	}

	return &generateParams
}

func (m *OamOasMapper) toOasPostGenerateResponse(err error, encryptedNonPublicJWKBytes, clearPublicJWKBytes []byte) (cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerateResponseObject, error) {
	// clearPublicJWKBytes is intentionally unused in current implementation
	// but kept for potential future logging/debugging purposes
	_ = clearPublicJWKBytes

	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDGenerate200TextResponse(encryptedNonPublicJWKBytes), err
}

func (m *OamOasMapper) toOamPostEncryptQueryParams(openapiParams *cryptoutilKmsServer.PostElastickeyElasticKeyIDEncryptParams) *cryptoutilOpenapiModel.EncryptParams {
	encryptParams := cryptoutilOpenapiModel.EncryptParams{
		Context: openapiParams.Context,
	}

	return &encryptParams
}

func (m *OamOasMapper) toOasPostEncryptResponse(err error, encryptedBytes []byte) (cryptoutilKmsServer.PostElastickeyElasticKeyIDEncryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDEncrypt200TextResponse(encryptedBytes), err
}

func (m *OamOasMapper) toOasPostElastickeyElasticKeyIDMaterialkeyResponse(err error, generateKeyInElasticKeyResponse *cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkeyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkey400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkey404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkey500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to generate Key by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDMaterialkey200JSONResponse(*generateKeyInElasticKeyResponse), err
}

func (m *OamOasMapper) toOasGetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponse(err error, key *cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeyMaterialKeyID200JSONResponse(*key), err
}

func (m *OamOasMapper) toOamGetElasticKeyMaterialKeysQueryParams(openapiParams *cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeysParams) *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams {
	// Convert from server params to model query params
	// Note: KMS server uses different field names than model
	var materialKeyIDs *[]cryptoutilOpenapiModel.MaterialKeyID

	if openapiParams.MaterialKeyIds != nil {
		ids := make([]cryptoutilOpenapiModel.MaterialKeyID, len(*openapiParams.MaterialKeyIds))
		for i, id := range *openapiParams.MaterialKeyIds {
			ids[i] = cryptoutilOpenapiModel.MaterialKeyID(id)
		}

		materialKeyIDs = &ids
	}

	var pageNum *cryptoutilOpenapiModel.PageNumber

	if openapiParams.PageNumber != nil {
		pn := cryptoutilOpenapiModel.PageNumber(*openapiParams.PageNumber)
		pageNum = &pn
	}

	var pageSize *cryptoutilOpenapiModel.PageSize

	if openapiParams.PageSize != nil {
		ps := cryptoutilOpenapiModel.PageSize(*openapiParams.PageSize)
		pageSize = &ps
	}

	filters := cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
		MaterialKeyID:     materialKeyIDs,
		MinGenerateDate:   (*cryptoutilOpenapiModel.MaterialKeyGenerateDate)(openapiParams.MinGenerateDate),
		MaxGenerateDate:   (*cryptoutilOpenapiModel.MaterialKeyGenerateDate)(openapiParams.MaxGenerateDate),
		MinImportDate:     (*cryptoutilOpenapiModel.MaterialKeyImportDate)(openapiParams.MinImportDate),
		MaxImportDate:     (*cryptoutilOpenapiModel.MaterialKeyImportDate)(openapiParams.MaxImportDate),
		MinExpirationDate: (*cryptoutilOpenapiModel.MaterialKeyExpirationDate)(openapiParams.MinExpirationDate),
		MaxExpirationDate: (*cryptoutilOpenapiModel.MaterialKeyExpirationDate)(openapiParams.MaxExpirationDate),
		MinRevocationDate: (*cryptoutilOpenapiModel.MaterialKeyRevocationDate)(openapiParams.MinRevocationDate),
		MaxRevocationDate: (*cryptoutilOpenapiModel.MaterialKeyRevocationDate)(openapiParams.MaxRevocationDate),
		Page:              pageNum,
		Size:              pageSize,
		// Note: Sort not in KMS server params, leave nil
	}

	return &filters
}

func (m *OamOasMapper) toOasGetElastickeyElasticKeyIDMaterialkeysResponse(err error, keys []cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeys404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeys200JSONResponse(keys), err
}

func (m *OamOasMapper) toOasPostSignResponse(err error, encryptedBytes []byte) (cryptoutilKmsServer.PostElastickeyElasticKeyIDSignResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDSign400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDSign404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDSign500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDSign200TextResponse(encryptedBytes), err
}

func (m *OamOasMapper) toOasPostVerifyResponse(err error) (cryptoutilKmsServer.PostElastickeyElasticKeyIDVerifyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDVerify400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDVerify404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElastickeyElasticKeyIDVerify500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to verify: %w", err)
	}

	return cryptoutilKmsServer.PostElastickeyElasticKeyIDVerify204Response{}, err
}

func (m *OamOasMapper) toOamGetElasticKeyQueryParams(openapiParams *cryptoutilKmsServer.GetElastickeysParams) *cryptoutilOpenapiModel.ElasticKeysQueryParams {
	// Convert from server params to model query params
	// Note: KMS server uses plural field names, model uses singular names for arrays
	var elasticKeyIDs *[]cryptoutilOpenapiModel.ElasticKeyID

	if openapiParams.ElasticKeyIds != nil {
		ids := make([]cryptoutilOpenapiModel.ElasticKeyID, len(*openapiParams.ElasticKeyIds))
		for i, id := range *openapiParams.ElasticKeyIds {
			ids[i] = cryptoutilOpenapiModel.ElasticKeyID(id)
		}

		elasticKeyIDs = &ids
	}

	var names *[]cryptoutilOpenapiModel.ElasticKeyName

	if openapiParams.Names != nil {
		n := make([]cryptoutilOpenapiModel.ElasticKeyName, len(*openapiParams.Names))
		for i, name := range *openapiParams.Names {
			n[i] = cryptoutilOpenapiModel.ElasticKeyName(name)
		}

		names = &n
	}

	var providers *[]cryptoutilOpenapiModel.ElasticKeyProvider

	if openapiParams.Providers != nil {
		p := make([]cryptoutilOpenapiModel.ElasticKeyProvider, len(*openapiParams.Providers))
		for i, prov := range *openapiParams.Providers {
			p[i] = cryptoutilOpenapiModel.ElasticKeyProvider(prov)
		}

		providers = &p
	}

	var algorithms *[]cryptoutilOpenapiModel.ElasticKeyAlgorithm

	if openapiParams.Algorithms != nil {
		a := make([]cryptoutilOpenapiModel.ElasticKeyAlgorithm, len(*openapiParams.Algorithms))
		for i, alg := range *openapiParams.Algorithms {
			a[i] = cryptoutilOpenapiModel.ElasticKeyAlgorithm(alg)
		}

		algorithms = &a
	}

	var statuses *[]cryptoutilOpenapiModel.ElasticKeyStatus

	if openapiParams.Statuses != nil {
		s := make([]cryptoutilOpenapiModel.ElasticKeyStatus, len(*openapiParams.Statuses))
		for i, status := range *openapiParams.Statuses {
			s[i] = cryptoutilOpenapiModel.ElasticKeyStatus(status)
		}

		statuses = &s
	}

	var sorts *[]cryptoutilOpenapiModel.ElasticKeySort

	if openapiParams.Sorts != nil {
		srt := make([]cryptoutilOpenapiModel.ElasticKeySort, len(*openapiParams.Sorts))
		for i, sort := range *openapiParams.Sorts {
			srt[i] = cryptoutilOpenapiModel.ElasticKeySort(sort)
		}

		sorts = &srt
	}

	var pageNum *cryptoutilOpenapiModel.PageNumber

	if openapiParams.PageNumber != nil {
		pn := cryptoutilOpenapiModel.PageNumber(*openapiParams.PageNumber)
		pageNum = &pn
	}

	var pageSize *cryptoutilOpenapiModel.PageSize

	if openapiParams.PageSize != nil {
		ps := cryptoutilOpenapiModel.PageSize(*openapiParams.PageSize)
		pageSize = &ps
	}

	filters := cryptoutilOpenapiModel.ElasticKeysQueryParams{
		ElasticKeyID:      elasticKeyIDs,
		Name:              names,
		Provider:          providers,
		Algorithm:         algorithms,
		VersioningAllowed: (*cryptoutilOpenapiModel.ElasticKeyVersioningAllowed)(openapiParams.VersioningAllowed),
		ImportAllowed:     (*cryptoutilOpenapiModel.ElasticKeyImportAllowed)(openapiParams.ImportAllowed),
		Status:            statuses,
		Sort:              sorts,
		Page:              pageNum,
		Size:              pageSize,
	}

	return &filters
}

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
