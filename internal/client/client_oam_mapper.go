package client

import (
	"errors"
	"fmt"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	googleUuid "github.com/google/uuid"
)

type oamOacMapper struct{}

// TODO Change all func to method
func NewOamOacMapper() *oamOacMapper {
	return &oamOacMapper{}
}

func toOamElasticKeyCreate(name *string, description *string, algorithm *string, provider *string, exportAllowed *bool, importAllowed *bool, versioningAllowed *bool) (*cryptoutilOpenapiModel.ElasticKeyCreate, error) {
	elasticKeyName := cryptoutilOpenapiModel.ElasticKeyName(*name)
	elasticKeyDescription := cryptoutilOpenapiModel.ElasticKeyDescription(*description)
	elasticKeyAlgorithm, err := cryptoutilJose.ToElasticKeyAlgorithm(algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to map elastic Key: %v", errors.Join(err))
	}
	elasticKeyProvider := cryptoutilOpenapiModel.ElasticKeyProvider(*provider)
	elasticKeyElasticKeyImportAllowed := cryptoutilOpenapiModel.ElasticKeyImportAllowed(*importAllowed)
	elasticKeyElasticKeyExportAllowed := cryptoutilOpenapiModel.ElasticKeyExportAllowed(*exportAllowed)
	elasticKeyElasticKeyVersioningAllowed := cryptoutilOpenapiModel.ElasticKeyVersioningAllowed(*versioningAllowed)
	return &cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:              elasticKeyName,
		Description:       elasticKeyDescription,
		Algorithm:         (*cryptoutilOpenapiModel.ElasticKeyAlgorithm)(elasticKeyAlgorithm),
		Provider:          &elasticKeyProvider,
		ImportAllowed:     &elasticKeyElasticKeyImportAllowed,
		ExportAllowed:     &elasticKeyElasticKeyExportAllowed,
		VersioningAllowed: &elasticKeyElasticKeyVersioningAllowed,
	}, nil
}

func toOamElasticKey(openapiCreateElasticKeyResponse *cryptoutilOpenapiClient.PostElastickeyResponse) (*cryptoutilOpenapiModel.ElasticKey, error) {
	if openapiCreateElasticKeyResponse == nil {
		return nil, fmt.Errorf("failed to create elastic Key, response is nil")
	} else if openapiCreateElasticKeyResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to create elastic Key, HTTP response is nil")
	}
	switch openapiCreateElasticKeyResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiCreateElasticKeyResponse.Body == nil {
			return nil, fmt.Errorf("failed to create elastic Key, body is nil")
		} else if openapiCreateElasticKeyResponse.JSON200 == nil {
			return nil, fmt.Errorf("failed to create elastic Key, JSON200 is nil")
		}
		elasticKey := openapiCreateElasticKeyResponse.JSON200
		if elasticKey.ElasticKeyID == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.Id is nil")
		} else if elasticKey.Description == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.Description is nil")
		} else if elasticKey.Algorithm == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.Algorithm is nil")
		} else if elasticKey.Provider == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.Provider is nil")
		} else if elasticKey.ExportAllowed == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.ExportAllowed is nil")
		} else if elasticKey.ImportAllowed == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.ImportAllowed is nil")
		} else if elasticKey.VersioningAllowed == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.VersioningAllowed is nil")
		} else if elasticKey.Status == nil {
			return nil, fmt.Errorf("failed to create elastic Key, elasticKey.Status is nil")
		}
		return elasticKey, nil
	default:
		return nil, fmt.Errorf("failed to create elastic Key, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiCreateElasticKeyResponse.HTTPResponse.StatusCode, openapiCreateElasticKeyResponse.HTTPResponse.Status, openapiCreateElasticKeyResponse.Body)
	}
}

func toOamMaterialKeyGenerate(openapiMaterialKeyGenerateResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDMaterialkeyResponse) (*cryptoutilOpenapiModel.MaterialKey, error) {
	if openapiMaterialKeyGenerateResponse == nil {
		return nil, fmt.Errorf("failed to generate key, response is nil")
	} else if openapiMaterialKeyGenerateResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to generate key, HTTP response is nil")
	}
	switch openapiMaterialKeyGenerateResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiMaterialKeyGenerateResponse.Body == nil {
			return nil, fmt.Errorf("failed to generate key, body is nil")
		} else if openapiMaterialKeyGenerateResponse.JSON200 == nil {
			return nil, fmt.Errorf("failed to generate key, JSON200 is nil")
		}
		key := openapiMaterialKeyGenerateResponse.JSON200
		if key.ElasticKeyID == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.Pool is zero")
		} else if key.MaterialKeyID == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.Id is zero")
		} else if key.GenerateDate == nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.GenerateDate is nil") // TODO nil allowed if import not nil
		}
		return key, nil
	default:
		return nil, fmt.Errorf("failed to generate key, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiMaterialKeyGenerateResponse.HTTPResponse.StatusCode, openapiMaterialKeyGenerateResponse.HTTPResponse.Status, openapiMaterialKeyGenerateResponse.Body)
	}
}

func toOacEncryptParams(encryptParams *cryptoutilOpenapiModel.EncryptParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptParams {
	elastickeyElasticKeyIDEncryptParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptParams{}
	if encryptParams != nil {
		elastickeyElasticKeyIDEncryptParams.Context = encryptParams.Context
	}
	return elastickeyElasticKeyIDEncryptParams
}

func toOamEncryptRequest(cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	encryptRequest := cryptoutilOpenapiModel.EncryptRequest(*cleartext)
	return &encryptRequest
}

func toPlainEncryptResponse(openapiEncryptResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptResponse) (*string, error) {
	if openapiEncryptResponse == nil {
		return nil, fmt.Errorf("failed to encrypt, response is nil")
	} else if openapiEncryptResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to encrypt, HTTP response is nil")
	}
	switch openapiEncryptResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiEncryptResponse.Body == nil {
			return nil, fmt.Errorf("failed to encrypt, body is nil")
		}
		ciphertext := string(openapiEncryptResponse.Body)
		return &ciphertext, nil
	default:
		return nil, fmt.Errorf("failed to encrypt, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiEncryptResponse.HTTPResponse.StatusCode, openapiEncryptResponse.HTTPResponse.Status, openapiEncryptResponse.Body)
	}
}

func toOamDecryptRequest(ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	decryptRequest := cryptoutilOpenapiModel.DecryptRequest(*ciphertext)
	return &decryptRequest
}

func toPlainDecryptResponse(openapiDecryptResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDDecryptResponse) (*string, error) {
	if openapiDecryptResponse == nil {
		return nil, fmt.Errorf("failed to decrypt, response is nil")
	} else if openapiDecryptResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to decrypt, HTTP response is nil")
	}
	switch openapiDecryptResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiDecryptResponse.Body == nil {
			return nil, fmt.Errorf("failed to decrypt, body is nil")
		}
		decrypted := string(openapiDecryptResponse.Body)
		return &decrypted, nil
	default:
		return nil, fmt.Errorf("failed to decrypt, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiDecryptResponse.HTTPResponse.StatusCode, openapiDecryptResponse.HTTPResponse.Status, openapiDecryptResponse.Body)
	}
}

func toOacSignParams(signParams *cryptoutilOpenapiModel.SignParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignParams {
	elastickeyElasticKeyIDSignParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignParams{}
	if signParams != nil {
		elastickeyElasticKeyIDSignParams.Context = signParams.Context
	}
	return elastickeyElasticKeyIDSignParams
}

func toOamSignRequest(cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	signRequest := cryptoutilOpenapiModel.SignRequest(*cleartext)
	return &signRequest
}

func toPlainSignResponse(openapiSignResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignResponse) (*string, error) {
	if openapiSignResponse == nil {
		return nil, fmt.Errorf("failed to sign, response is nil")
	} else if openapiSignResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to sign, HTTP response is nil")
	}
	switch openapiSignResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiSignResponse.Body == nil {
			return nil, fmt.Errorf("failed to sign, body is nil")
		}
		ciphertext := string(openapiSignResponse.Body)
		return &ciphertext, nil
	default:
		return nil, fmt.Errorf("failed to sign, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiSignResponse.HTTPResponse.StatusCode, openapiSignResponse.HTTPResponse.Status, openapiSignResponse.Body)
	}
}

func toOamVerifyRequest(signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	verifyRequest := cryptoutilOpenapiModel.VerifyRequest(*signedtext)
	return &verifyRequest
}

func toPlainVerifyResponse(openapiVerifyResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDVerifyResponse) (*string, error) {
	if openapiVerifyResponse == nil {
		return nil, fmt.Errorf("failed to verify, response is nil")
	} else if openapiVerifyResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to verify, HTTP response is nil")
	}
	switch openapiVerifyResponse.HTTPResponse.StatusCode {
	case 204:
		if openapiVerifyResponse.Body == nil {
			return nil, fmt.Errorf("failed to verify, body is nil")
		}
		verified := string(openapiVerifyResponse.Body)
		return &verified, nil
	default:
		return nil, fmt.Errorf("failed to verify, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiVerifyResponse.HTTPResponse.StatusCode, openapiVerifyResponse.HTTPResponse.Status, openapiVerifyResponse.Body)
	}
}

func toOacMaterialKeyGenerater() (*cryptoutilOpenapiClient.PostElastickeyElasticKeyIDMaterialkeyJSONRequestBody, error) {
	return &cryptoutilOpenapiModel.MaterialKeyGenerate{}, nil
}
