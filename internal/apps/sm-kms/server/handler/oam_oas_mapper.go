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
type OamOasMapper struct{}

// NewOasOamMapper creates a new mapper between OpenAPI Model and OpenAPI Server types.
func NewOasOamMapper() *OamOasMapper {
	return &OamOasMapper{}
}

func (m *OamOasMapper) toOasPostElasticKeysResponse(err error, addedElasticKey *cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.PostElasticKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to add ElasticKey: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeys200JSONResponse(*addedElasticKey), nil
}

func (m *OamOasMapper) toOasGetElasticKeysElasticKeyIDResponse(err error, elasticKey *cryptoutilKmsServer.ElasticKey) (cryptoutilKmsServer.GetElasticKeysElasticKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetElasticKeysElasticKeyID200JSONResponse(*elasticKey), err
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDDecryptResponse(err error, decryptedBytes []byte) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecrypt400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecrypt404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecrypt500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDDecrypt200TextResponse(decryptedBytes), err
}

func (m *OamOasMapper) toOamPostElasticKeysElasticKeyIDGenerateQueryParams(openapiParams *cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerateParams) *cryptoutilOpenapiModel.GenerateParams {
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

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDGenerateResponse(err error, encryptedNonPublicJWKBytes, clearPublicJWKBytes []byte) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerateResponseObject, error) {
	// clearPublicJWKBytes is intentionally unused in current implementation
	// but kept for potential future logging/debugging purposes
	_ = clearPublicJWKBytes

	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerate400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerate404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerate500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDGenerate200TextResponse(encryptedNonPublicJWKBytes), err
}

func (m *OamOasMapper) toOamPostElasticKeysElasticKeyIDEncryptQueryParams(openapiParams *cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncryptParams) *cryptoutilOpenapiModel.EncryptParams {
	encryptParams := cryptoutilOpenapiModel.EncryptParams{
		Context: openapiParams.Context,
	}

	return &encryptParams
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDEncryptResponse(err error, encryptedBytes []byte) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncryptResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncrypt400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncrypt404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncrypt500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDEncrypt200TextResponse(encryptedBytes), err
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDMaterialKeysResponse(err error, generateKeyInElasticKeyResponse *cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeys404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to generate Key by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDMaterialKeys200JSONResponse(*generateKeyInElasticKeyResponse), err
}

func (m *OamOasMapper) toOasGetElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponse(err error, key *cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyIDResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyID400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyID404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyID500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysMaterialKeyID200JSONResponse(*key), err
}

func (m *OamOasMapper) toOamGetElasticKeysElasticKeyIDMaterialKeysQueryParams(openapiParams *cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysParams) *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams {
	// Convert from server params to model query params
	// Note: KMS server uses different field names than model
	var materialKeyIDs *[]cryptoutilOpenapiModel.MaterialKeyID

	if openapiParams.MaterialKeyIDS != nil {
		ids := make([]cryptoutilOpenapiModel.MaterialKeyID, len(*openapiParams.MaterialKeyIDS))
		for i, id := range *openapiParams.MaterialKeyIDS {
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

func (m *OamOasMapper) toOasGetElasticKeysElasticKeyIDMaterialKeysResponse(err error, keys []cryptoutilKmsServer.MaterialKey) (cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeysResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeys400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeys404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeys500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
	}

	return cryptoutilKmsServer.GetElasticKeysElasticKeyIDMaterialKeys200JSONResponse(keys), err
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDSignResponse(err error, encryptedBytes []byte) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDSignResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDSign400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDSign404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDSign500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDSign200TextResponse(encryptedBytes), err
}

func (m *OamOasMapper) toOasPostElasticKeysElasticKeyIDVerifyResponse(err error) (cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerifyResponseObject, error) {
	if err != nil {
		var appErr *cryptoutilSharedApperr.Error
		if errors.As(err, &appErr) {
			switch appErr.HTTPStatusLineAndCode.StatusLine.StatusCode {
			case http.StatusBadRequest:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerify400JSONResponse{BadRequestJSONResponse: m.toOasHTTP400Response(appErr)}, nil
			case http.StatusNotFound:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerify404JSONResponse{NotFoundJSONResponse: m.toOasHTTP404Response(appErr)}, nil
			case http.StatusInternalServerError:
				return cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerify500JSONResponse{InternalServerErrorJSONResponse: m.toOasHTTP500Response(appErr)}, nil
			}
		}

		return nil, fmt.Errorf("failed to verify: %w", err)
	}

	return cryptoutilKmsServer.PostElasticKeysElasticKeyIDVerify204Response{}, err
}

func (m *OamOasMapper) toOamGetElasticKeyQueryParams(openapiParams *cryptoutilKmsServer.GetElasticKeysParams) *cryptoutilOpenapiModel.ElasticKeysQueryParams {
	// Convert from server params to model query params
	// Note: KMS server uses plural field names, model uses singular names for arrays
	var elasticKeyIDs *[]cryptoutilOpenapiModel.ElasticKeyID

	if openapiParams.ElasticKeyIDS != nil {
		ids := make([]cryptoutilOpenapiModel.ElasticKeyID, len(*openapiParams.ElasticKeyIDS))
		for i, id := range *openapiParams.ElasticKeyIDS {
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
