//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

// Package client provides KMS client functionality for cryptoutil.
// NOTE: This file uses legacy OpenAPI client types and is only used for integration tests.
package client

import (
	"fmt"
	http "net/http"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"
)

// OamOacMapper maps between OpenAPI Model and OpenAPI Client types.
type OamOacMapper struct{}

// NewOamOacMapper creates a new mapper between OpenAPI Model and OpenAPI Client types.
func NewOamOacMapper() *OamOacMapper {
	return &OamOacMapper{}
}

func (m *OamOacMapper) toOamElasticKeyCreate(name, description, algorithm, provider *string, importAllowed, versioningAllowed *bool) (*cryptoutilOpenapiModel.ElasticKeyCreate, error) {
	elasticKeyAlgorithm, err := cryptoutilSharedCryptoJose.ToElasticKeyAlgorithm(algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to map Elastic Key: %w", err)
	}

	elasticKeyProvider := cryptoutilOpenapiModel.ElasticKeyProvider(*provider)

	return &cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:              *name,
		Description:       *description,
		Algorithm:         elasticKeyAlgorithm,
		Provider:          &elasticKeyProvider,
		ImportAllowed:     importAllowed,
		VersioningAllowed: versioningAllowed,
	}, nil
}

func (m *OamOacMapper) toOamElasticKey(openapiCreateElasticKeyResponse *cryptoutilOpenapiClient.PostElastickeyResponse) (*cryptoutilOpenapiModel.ElasticKey, error) {
	if openapiCreateElasticKeyResponse == nil {
		return nil, fmt.Errorf("failed to create Elastic Key, response is nil")
	} else if openapiCreateElasticKeyResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to create Elastic Key, HTTP response is nil")
	}

	switch openapiCreateElasticKeyResponse.HTTPResponse.StatusCode {
	case http.StatusOK:
		if openapiCreateElasticKeyResponse.Body == nil {
			return nil, fmt.Errorf("failed to create Elastic Key, body is nil")
		} else if openapiCreateElasticKeyResponse.JSON200 == nil {
			return nil, fmt.Errorf("failed to create Elastic Key, JSON200 is nil")
		}

		elasticKey := openapiCreateElasticKeyResponse.JSON200

		// According to OpenAPI spec, all ElasticKey fields are optional, so don't validate they are non-nil
		return elasticKey, nil
	default:
		return nil, fmt.Errorf("failed to create Elastic Key, Status: %d, Message: %s, Body: %s", openapiCreateElasticKeyResponse.HTTPResponse.StatusCode, openapiCreateElasticKeyResponse.HTTPResponse.Status, openapiCreateElasticKeyResponse.Body)
	}
}

func (m *OamOacMapper) toOamMaterialKeyGenerate(openapiMaterialKeyGenerateResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDMaterialkeyResponse) (*cryptoutilOpenapiModel.MaterialKey, error) {
	if openapiMaterialKeyGenerateResponse == nil {
		return nil, fmt.Errorf("failed to generate key, response is nil")
	} else if openapiMaterialKeyGenerateResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to generate key, HTTP response is nil")
	}

	switch openapiMaterialKeyGenerateResponse.HTTPResponse.StatusCode {
	case http.StatusOK:
		if openapiMaterialKeyGenerateResponse.Body == nil {
			return nil, fmt.Errorf("failed to generate key, body is nil")
		} else if openapiMaterialKeyGenerateResponse.JSON200 == nil {
			return nil, fmt.Errorf("failed to generate key, JSON200 is nil")
		}

		key := openapiMaterialKeyGenerateResponse.JSON200

		if key.ElasticKeyID == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.ElasticKeyID is zero")
		} else if key.MaterialKeyID == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.MaterialKeyID is zero")
		} else if key.GenerateDate == nil && key.ImportDate == nil {
			return nil, fmt.Errorf("failed to generate key, both GenerateDate and ImportDate are nil")
		}

		return key, nil
	default:
		return nil, fmt.Errorf("failed to generate key, nextElasticKeyName(), Status: %d, Message: %s, Body: %s", openapiMaterialKeyGenerateResponse.HTTPResponse.StatusCode, openapiMaterialKeyGenerateResponse.HTTPResponse.Status, openapiMaterialKeyGenerateResponse.Body)
	}
}

func (m *OamOacMapper) toOacGenerateParams(generateParams *cryptoutilOpenapiModel.GenerateParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDGenerateParams {
	elastickeyElasticKeyIDGenerateParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDGenerateParams{}
	if generateParams != nil {
		elastickeyElasticKeyIDGenerateParams.Context = generateParams.Context
		elastickeyElasticKeyIDGenerateParams.Alg = generateParams.Alg
	}

	return elastickeyElasticKeyIDGenerateParams
}

// func toOamGenerateRequest(cleartext *string) *cryptoutilOpenapiModel.GenerateRequest {
// 	encryptRequest := cryptoutilOpenapiModel.GenerateRequest(*cleartext)
// 	return &encryptRequest
// }

func (m *OamOacMapper) toPlainGenerateResponse(openapiGenerateResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDGenerateResponse) (*string, error) {
	if openapiGenerateResponse == nil {
		return nil, fmt.Errorf("failed to encrypt, response is nil")
	} else if openapiGenerateResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to encrypt, HTTP response is nil")
	}

	switch openapiGenerateResponse.HTTPResponse.StatusCode {
	case http.StatusOK:
		if openapiGenerateResponse.Body == nil {
			return nil, fmt.Errorf("failed to encrypt, body is nil")
		}

		ciphertext := string(openapiGenerateResponse.Body)

		return &ciphertext, nil
	default:
		return nil, fmt.Errorf("failed to encrypt, nextElasticKeyName(), Status: %d, Message: %s, Body: %s", openapiGenerateResponse.HTTPResponse.StatusCode, openapiGenerateResponse.HTTPResponse.Status, openapiGenerateResponse.Body)
	}
}

func (m *OamOacMapper) toOacEncryptParams(encryptParams *cryptoutilOpenapiModel.EncryptParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptParams {
	elastickeyElasticKeyIDEncryptParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptParams{}
	if encryptParams != nil {
		elastickeyElasticKeyIDEncryptParams.Context = encryptParams.Context
	}

	return elastickeyElasticKeyIDEncryptParams
}

func (m *OamOacMapper) toOamEncryptRequest(cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	return cleartext
}

func (m *OamOacMapper) toPlainEncryptResponse(openapiEncryptResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptResponse) (*string, error) {
	if openapiEncryptResponse == nil {
		return nil, fmt.Errorf("failed to encrypt, response is nil")
	} else if openapiEncryptResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to encrypt, HTTP response is nil")
	}

	switch openapiEncryptResponse.HTTPResponse.StatusCode {
	case http.StatusOK:
		if openapiEncryptResponse.Body == nil {
			return nil, fmt.Errorf("failed to encrypt, body is nil")
		}

		ciphertext := string(openapiEncryptResponse.Body)

		return &ciphertext, nil
	default:
		return nil, fmt.Errorf("failed to encrypt, nextElasticKeyName(), Status: %d, Message: %s, Body: %s", openapiEncryptResponse.HTTPResponse.StatusCode, openapiEncryptResponse.HTTPResponse.Status, openapiEncryptResponse.Body)
	}
}

func (m *OamOacMapper) toOamDecryptRequest(ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	return ciphertext
}

func (m *OamOacMapper) toPlainDecryptResponse(openapiDecryptResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDDecryptResponse) (*string, error) {
	if openapiDecryptResponse == nil {
		return nil, fmt.Errorf("failed to decrypt, response is nil")
	} else if openapiDecryptResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to decrypt, HTTP response is nil")
	}

	switch openapiDecryptResponse.HTTPResponse.StatusCode {
	case http.StatusOK:
		if openapiDecryptResponse.Body == nil {
			return nil, fmt.Errorf("failed to decrypt, body is nil")
		}

		decrypted := string(openapiDecryptResponse.Body)

		return &decrypted, nil
	default:
		return nil, fmt.Errorf("failed to decrypt, nextElasticKeyName(), Status: %d, Message: %s, Body: %s", openapiDecryptResponse.HTTPResponse.StatusCode, openapiDecryptResponse.HTTPResponse.Status, openapiDecryptResponse.Body)
	}
}

func (m *OamOacMapper) toOacSignParams(signParams *cryptoutilOpenapiModel.SignParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignParams {
	elastickeyElasticKeyIDSignParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignParams{}
	if signParams != nil {
		elastickeyElasticKeyIDSignParams.Context = signParams.Context
	}

	return elastickeyElasticKeyIDSignParams
}

func (m *OamOacMapper) toOamSignRequest(cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	return cleartext
}

func (m *OamOacMapper) toPlainSignResponse(openapiSignResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignResponse) (*string, error) {
	if openapiSignResponse == nil {
		return nil, fmt.Errorf("failed to sign, response is nil")
	} else if openapiSignResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to sign, HTTP response is nil")
	}

	switch openapiSignResponse.HTTPResponse.StatusCode {
	case http.StatusOK:
		if openapiSignResponse.Body == nil {
			return nil, fmt.Errorf("failed to sign, body is nil")
		}

		ciphertext := string(openapiSignResponse.Body)

		return &ciphertext, nil
	default:
		return nil, fmt.Errorf("failed to sign, nextElasticKeyName(), Status: %d, Message: %s, Body: %s", openapiSignResponse.HTTPResponse.StatusCode, openapiSignResponse.HTTPResponse.Status, openapiSignResponse.Body)
	}
}

func (m *OamOacMapper) toOamVerifyRequest(signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	return signedtext
}

func (m *OamOacMapper) toPlainVerifyResponse(openapiVerifyResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDVerifyResponse) (*string, error) {
	if openapiVerifyResponse == nil {
		return nil, fmt.Errorf("failed to verify, response is nil")
	} else if openapiVerifyResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to verify, HTTP response is nil")
	}

	switch openapiVerifyResponse.HTTPResponse.StatusCode {
	case http.StatusNoContent:
		// 204 No Content means verification succeeded, return empty string
		empty := ""

		return &empty, nil
	default:
		return nil, fmt.Errorf("failed to verify, Status: %d, Message: %s, Body: %s", openapiVerifyResponse.HTTPResponse.StatusCode, openapiVerifyResponse.HTTPResponse.Status, openapiVerifyResponse.Body)
	}
}
