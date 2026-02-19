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
