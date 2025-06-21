package client

import (
	"errors"
	"fmt"
	"strings"

	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	googleUuid "github.com/google/uuid"
)

type ClientMapper struct{}

// TODO Change all func to method
func NewClientMapper() *ClientMapper {
	return &ClientMapper{}
}

var (
	validAlgorithms = map[string]cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		string(cryptoutilOpenapiModel.A256GCMA256KW): cryptoutilOpenapiModel.A256GCMA256KW,
		string(cryptoutilOpenapiModel.A192GCMA256KW): cryptoutilOpenapiModel.A192GCMA256KW,
		string(cryptoutilOpenapiModel.A128GCMA256KW): cryptoutilOpenapiModel.A128GCMA256KW,
		string(cryptoutilOpenapiModel.A256GCMA192KW): cryptoutilOpenapiModel.A256GCMA192KW,
		string(cryptoutilOpenapiModel.A192GCMA192KW): cryptoutilOpenapiModel.A192GCMA192KW,
		string(cryptoutilOpenapiModel.A128GCMA192KW): cryptoutilOpenapiModel.A128GCMA192KW,
		string(cryptoutilOpenapiModel.A256GCMA128KW): cryptoutilOpenapiModel.A256GCMA128KW,
		string(cryptoutilOpenapiModel.A192GCMA128KW): cryptoutilOpenapiModel.A192GCMA128KW,
		string(cryptoutilOpenapiModel.A128GCMA128KW): cryptoutilOpenapiModel.A128GCMA128KW,

		string(cryptoutilOpenapiModel.A256GCMA256GCMKW): cryptoutilOpenapiModel.A256GCMA256GCMKW,
		string(cryptoutilOpenapiModel.A192GCMA256GCMKW): cryptoutilOpenapiModel.A192GCMA256GCMKW,
		string(cryptoutilOpenapiModel.A128GCMA256GCMKW): cryptoutilOpenapiModel.A128GCMA256GCMKW,
		string(cryptoutilOpenapiModel.A256GCMA192GCMKW): cryptoutilOpenapiModel.A256GCMA192GCMKW,
		string(cryptoutilOpenapiModel.A192GCMA192GCMKW): cryptoutilOpenapiModel.A192GCMA192GCMKW,
		string(cryptoutilOpenapiModel.A128GCMA192GCMKW): cryptoutilOpenapiModel.A128GCMA192GCMKW,
		string(cryptoutilOpenapiModel.A256GCMA128GCMKW): cryptoutilOpenapiModel.A256GCMA128GCMKW,
		string(cryptoutilOpenapiModel.A192GCMA128GCMKW): cryptoutilOpenapiModel.A192GCMA128GCMKW,
		string(cryptoutilOpenapiModel.A128GCMA128GCMKW): cryptoutilOpenapiModel.A128GCMA128GCMKW,

		string(cryptoutilOpenapiModel.A256GCMdir): cryptoutilOpenapiModel.A256GCMdir,
		string(cryptoutilOpenapiModel.A192GCMdir): cryptoutilOpenapiModel.A192GCMdir,
		string(cryptoutilOpenapiModel.A128GCMdir): cryptoutilOpenapiModel.A128GCMdir,

		string(cryptoutilOpenapiModel.A256GCMRSAOAEP512): cryptoutilOpenapiModel.A256GCMRSAOAEP512,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP512): cryptoutilOpenapiModel.A192GCMRSAOAEP512,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP512): cryptoutilOpenapiModel.A128GCMRSAOAEP512,
		string(cryptoutilOpenapiModel.A256GCMRSAOAEP384): cryptoutilOpenapiModel.A256GCMRSAOAEP384,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP384): cryptoutilOpenapiModel.A192GCMRSAOAEP384,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP384): cryptoutilOpenapiModel.A128GCMRSAOAEP384,
		string(cryptoutilOpenapiModel.A256GCMRSAOAEP256): cryptoutilOpenapiModel.A256GCMRSAOAEP256,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP256): cryptoutilOpenapiModel.A192GCMRSAOAEP256,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP256): cryptoutilOpenapiModel.A128GCMRSAOAEP256,
		string(cryptoutilOpenapiModel.A256GCMRSAOAEP):    cryptoutilOpenapiModel.A256GCMRSAOAEP,
		string(cryptoutilOpenapiModel.A192GCMRSAOAEP):    cryptoutilOpenapiModel.A192GCMRSAOAEP,
		string(cryptoutilOpenapiModel.A128GCMRSAOAEP):    cryptoutilOpenapiModel.A128GCMRSAOAEP,
		string(cryptoutilOpenapiModel.A256GCMRSA15):      cryptoutilOpenapiModel.A256GCMRSA15,
		string(cryptoutilOpenapiModel.A192GCMRSA15):      cryptoutilOpenapiModel.A192GCMRSA15,
		string(cryptoutilOpenapiModel.A128GCMRSA15):      cryptoutilOpenapiModel.A128GCMRSA15,

		string(cryptoutilOpenapiModel.A256GCMECDHESA256KW): cryptoutilOpenapiModel.A256GCMECDHESA256KW,
		string(cryptoutilOpenapiModel.A192GCMECDHESA256KW): cryptoutilOpenapiModel.A192GCMECDHESA256KW,
		string(cryptoutilOpenapiModel.A128GCMECDHESA256KW): cryptoutilOpenapiModel.A128GCMECDHESA256KW,
		string(cryptoutilOpenapiModel.A256GCMECDHESA192KW): cryptoutilOpenapiModel.A256GCMECDHESA192KW,
		string(cryptoutilOpenapiModel.A192GCMECDHESA192KW): cryptoutilOpenapiModel.A192GCMECDHESA192KW,
		string(cryptoutilOpenapiModel.A128GCMECDHESA192KW): cryptoutilOpenapiModel.A128GCMECDHESA192KW,
		string(cryptoutilOpenapiModel.A256GCMECDHESA128KW): cryptoutilOpenapiModel.A256GCMECDHESA128KW,
		string(cryptoutilOpenapiModel.A192GCMECDHESA128KW): cryptoutilOpenapiModel.A192GCMECDHESA128KW,
		string(cryptoutilOpenapiModel.A128GCMECDHESA128KW): cryptoutilOpenapiModel.A128GCMECDHESA128KW,
		string(cryptoutilOpenapiModel.A256GCMECDHES):       cryptoutilOpenapiModel.A256GCMECDHES,
		string(cryptoutilOpenapiModel.A192GCMECDHES):       cryptoutilOpenapiModel.A192GCMECDHES,
		string(cryptoutilOpenapiModel.A128GCMECDHES):       cryptoutilOpenapiModel.A128GCMECDHES,

		string(cryptoutilOpenapiModel.A256CBCHS512A256KW): cryptoutilOpenapiModel.A256CBCHS512A256KW,
		string(cryptoutilOpenapiModel.A192CBCHS384A256KW): cryptoutilOpenapiModel.A192CBCHS384A256KW,
		string(cryptoutilOpenapiModel.A128CBCHS256A256KW): cryptoutilOpenapiModel.A128CBCHS256A256KW,
		string(cryptoutilOpenapiModel.A256CBCHS512A192KW): cryptoutilOpenapiModel.A256CBCHS512A192KW,
		string(cryptoutilOpenapiModel.A192CBCHS384A192KW): cryptoutilOpenapiModel.A192CBCHS384A192KW,
		string(cryptoutilOpenapiModel.A128CBCHS256A192KW): cryptoutilOpenapiModel.A128CBCHS256A192KW,
		string(cryptoutilOpenapiModel.A256CBCHS512A128KW): cryptoutilOpenapiModel.A256CBCHS512A128KW,
		string(cryptoutilOpenapiModel.A192CBCHS384A128KW): cryptoutilOpenapiModel.A192CBCHS384A128KW,
		string(cryptoutilOpenapiModel.A128CBCHS256A128KW): cryptoutilOpenapiModel.A128CBCHS256A128KW,

		string(cryptoutilOpenapiModel.A256CBCHS512A256GCMKW): cryptoutilOpenapiModel.A256CBCHS512A256GCMKW,
		string(cryptoutilOpenapiModel.A192CBCHS384A256GCMKW): cryptoutilOpenapiModel.A192CBCHS384A256GCMKW,
		string(cryptoutilOpenapiModel.A128CBCHS256A256GCMKW): cryptoutilOpenapiModel.A128CBCHS256A256GCMKW,
		string(cryptoutilOpenapiModel.A256CBCHS512A192GCMKW): cryptoutilOpenapiModel.A256CBCHS512A192GCMKW,
		string(cryptoutilOpenapiModel.A192CBCHS384A192GCMKW): cryptoutilOpenapiModel.A192CBCHS384A192GCMKW,
		string(cryptoutilOpenapiModel.A128CBCHS256A192GCMKW): cryptoutilOpenapiModel.A128CBCHS256A192GCMKW,
		string(cryptoutilOpenapiModel.A256CBCHS512A128GCMKW): cryptoutilOpenapiModel.A256CBCHS512A128GCMKW,
		string(cryptoutilOpenapiModel.A192CBCHS384A128GCMKW): cryptoutilOpenapiModel.A192CBCHS384A128GCMKW,
		string(cryptoutilOpenapiModel.A128CBCHS256A128GCMKW): cryptoutilOpenapiModel.A128CBCHS256A128GCMKW,

		string(cryptoutilOpenapiModel.A256CBCHS512dir): cryptoutilOpenapiModel.A256CBCHS512dir,
		string(cryptoutilOpenapiModel.A192CBCHS384dir): cryptoutilOpenapiModel.A192CBCHS384dir,
		string(cryptoutilOpenapiModel.A128CBCHS256dir): cryptoutilOpenapiModel.A128CBCHS256dir,

		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512): cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512): cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512): cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512,
		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384): cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384): cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384): cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384,
		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256): cryptoutilOpenapiModel.A256CBCHS512RSAOAEP256,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256): cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256): cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256,
		string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP):    cryptoutilOpenapiModel.A256CBCHS512RSAOAEP,
		string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP):    cryptoutilOpenapiModel.A192CBCHS384RSAOAEP,
		string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP):    cryptoutilOpenapiModel.A128CBCHS256RSAOAEP,
		string(cryptoutilOpenapiModel.A256CBCHS512RSA15):      cryptoutilOpenapiModel.A256CBCHS512RSA15,
		string(cryptoutilOpenapiModel.A192CBCHS384RSA15):      cryptoutilOpenapiModel.A192CBCHS384RSA15,
		string(cryptoutilOpenapiModel.A128CBCHS256RSA15):      cryptoutilOpenapiModel.A128CBCHS256RSA15,

		string(cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW): cryptoutilOpenapiModel.A256CBCHS512ECDHESA256KW,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW): cryptoutilOpenapiModel.A192CBCHS384ECDHESA256KW,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW): cryptoutilOpenapiModel.A128CBCHS256ECDHESA256KW,
		string(cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW): cryptoutilOpenapiModel.A256CBCHS512ECDHESA192KW,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW): cryptoutilOpenapiModel.A192CBCHS384ECDHESA192KW,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW): cryptoutilOpenapiModel.A128CBCHS256ECDHESA192KW,
		string(cryptoutilOpenapiModel.A256CBCHS512ECDHESA128KW): cryptoutilOpenapiModel.A256CBCHS512ECDHESA128KW,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHESA128KW): cryptoutilOpenapiModel.A192CBCHS384ECDHESA128KW,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW): cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW,
		string(cryptoutilOpenapiModel.A256CBCHS512ECDHES):       cryptoutilOpenapiModel.A256CBCHS512ECDHES,
		string(cryptoutilOpenapiModel.A192CBCHS384ECDHES):       cryptoutilOpenapiModel.A192CBCHS384ECDHES,
		string(cryptoutilOpenapiModel.A128CBCHS256ECDHES):       cryptoutilOpenapiModel.A128CBCHS256ECDHES,

		string(cryptoutilOpenapiModel.RS256): cryptoutilOpenapiModel.RS256,
		string(cryptoutilOpenapiModel.RS384): cryptoutilOpenapiModel.RS384,
		string(cryptoutilOpenapiModel.RS512): cryptoutilOpenapiModel.RS512,
		string(cryptoutilOpenapiModel.PS256): cryptoutilOpenapiModel.PS256,
		string(cryptoutilOpenapiModel.PS384): cryptoutilOpenapiModel.PS384,
		string(cryptoutilOpenapiModel.PS512): cryptoutilOpenapiModel.PS512,
		string(cryptoutilOpenapiModel.ES256): cryptoutilOpenapiModel.ES256,
		string(cryptoutilOpenapiModel.ES384): cryptoutilOpenapiModel.ES384,
		string(cryptoutilOpenapiModel.ES512): cryptoutilOpenapiModel.ES512,
		string(cryptoutilOpenapiModel.HS256): cryptoutilOpenapiModel.HS256,
		string(cryptoutilOpenapiModel.HS384): cryptoutilOpenapiModel.HS384,
		string(cryptoutilOpenapiModel.HS512): cryptoutilOpenapiModel.HS512,
		string(cryptoutilOpenapiModel.EdDSA): cryptoutilOpenapiModel.EdDSA,
	}
)

func MapElasticKeyCreate(name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) (*cryptoutilOpenapiModel.ElasticKeyCreate, error) {
	elasticKeyName, err1 := MapElasticKeyName(name)
	elasticKeyDescription, err2 := MapElasticKeyDescription(description)
	elasticKeyAlgorithm, err3 := MapElasticKeyAlgorithm(algorithm)
	elasticKeyProvider, err4 := MapElasticKeyProvider(provider)
	elasticKeyElasticKeyExportAllowed := MapElasticKeyExportAllowed(exportAllowed)
	elasticKeyElasticKeyImportAllowed := MapElasticKeyImportAllowed(importAllowed)
	elasticKeyElasticKeyVersioningAllowed := MapElasticKeyVersioningAllowed(versioningAllowed)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, fmt.Errorf("failed to map elastic Key: %v", errors.Join(err1, err2, err3, err4))
	}
	return &cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:              *elasticKeyName,
		Description:       *elasticKeyDescription,
		Provider:          elasticKeyProvider,
		Algorithm:         elasticKeyAlgorithm,
		ExportAllowed:     elasticKeyElasticKeyExportAllowed,
		ImportAllowed:     elasticKeyElasticKeyImportAllowed,
		VersioningAllowed: elasticKeyElasticKeyVersioningAllowed,
	}, nil
}

func MapElasticKey(openapiCreateElasticKeyResponse *cryptoutilOpenapiClient.PostElastickeyResponse) (*cryptoutilOpenapiModel.ElasticKey, error) {
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
		if elasticKey.ElasticKeyId == nil {
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

func MapMaterialKeyGenerate(openapiMaterialKeyGenerateResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDMaterialkeyResponse) (*cryptoutilOpenapiModel.MaterialKey, error) {
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
		if key.ElasticKeyId == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.Pool is zero")
		} else if key.MaterialKeyId == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.Id is zero")
		} else if key.GenerateDate == nil {
			return nil, fmt.Errorf("failed to generate key, elasticKey.GenerateDate is nil") // TODO nil allowed if import not nil
		}
		return key, nil
	default:
		return nil, fmt.Errorf("failed to generate key, nextElasticKeyName(), Status: %v, Message: %s, Body: %s", openapiMaterialKeyGenerateResponse.HTTPResponse.StatusCode, openapiMaterialKeyGenerateResponse.HTTPResponse.Status, openapiMaterialKeyGenerateResponse.Body)
	}
}

func MapEncryptParams(encryptParams *cryptoutilOpenapiModel.EncryptParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptParams {
	elastickeyElasticKeyIDEncryptParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptParams{}
	if encryptParams != nil {
		elastickeyElasticKeyIDEncryptParams.Context = encryptParams.Context
	}
	return elastickeyElasticKeyIDEncryptParams
}

func MapEncryptRequest(cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	encryptRequest := cryptoutilOpenapiModel.EncryptRequest(*cleartext)
	return &encryptRequest
}

func MapEncryptResponse(openapiEncryptResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDEncryptResponse) (*string, error) {
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

func MapDecryptRequest(ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	decryptRequest := cryptoutilOpenapiModel.DecryptRequest(*ciphertext)
	return &decryptRequest
}

func MapDecryptResponse(openapiDecryptResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDDecryptResponse) (*string, error) {
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

func MapSignParams(signParams *cryptoutilOpenapiModel.SignParams) cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignParams {
	elastickeyElasticKeyIDSignParams := cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignParams{}
	if signParams != nil {
		elastickeyElasticKeyIDSignParams.Context = signParams.Context
	}
	return elastickeyElasticKeyIDSignParams
}

func MapSignRequest(cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	signRequest := cryptoutilOpenapiModel.SignRequest(*cleartext)
	return &signRequest
}

func MapSignResponse(openapiSignResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDSignResponse) (*string, error) {
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

func MapVerifyRequest(signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	verifyRequest := cryptoutilOpenapiModel.VerifyRequest(*signedtext)
	return &verifyRequest
}

func MapVerifyResponse(openapiVerifyResponse *cryptoutilOpenapiClient.PostElastickeyElasticKeyIDVerifyResponse) (*string, error) {
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

func MapMaterialKeyGenerater() (*cryptoutilOpenapiClient.PostElastickeyElasticKeyIDMaterialkeyJSONRequestBody, error) {
	return &cryptoutilOpenapiModel.MaterialKeyGenerate{}, nil
}

func MapElasticKeyName(name string) (*cryptoutilOpenapiModel.ElasticKeyName, error) {
	if err := ValidateString(name); err != nil {
		return nil, fmt.Errorf("invalid elastic Key name: %w", err)
	}
	elasticKeyName := cryptoutilOpenapiModel.ElasticKeyName(name)
	return &elasticKeyName, nil
}

func MapElasticKeyDescription(description string) (*cryptoutilOpenapiModel.ElasticKeyDescription, error) {
	if err := ValidateString(description); err != nil {
		return nil, fmt.Errorf("invalid elastic Key description: %w", err)
	}
	elasticKeyDescription := cryptoutilOpenapiModel.ElasticKeyDescription(description)
	return &elasticKeyDescription, nil
}

func MapElasticKeyAlgorithm(algorithm string) (*cryptoutilOpenapiModel.ElasticKeyAlgorithm, error) {
	if err := ValidateString(algorithm); err != nil {
		return nil, fmt.Errorf("invalid elastic Key algorithm: %w", err)
	}
	if alg, exists := validAlgorithms[algorithm]; exists {
		return &alg, nil
	}
	return nil, fmt.Errorf("invalid elastic Key algorithm: %s", algorithm)
}

func MapElasticKeyProvider(provider string) (*cryptoutilOpenapiModel.ElasticKeyProvider, error) {
	if err := ValidateString(provider); err != nil {
		return nil, fmt.Errorf("invalid elastic Key provider value: %w", err)
	}
	var elasticKeyProvider cryptoutilOpenapiModel.ElasticKeyProvider
	switch provider {
	case string(cryptoutilOpenapiModel.Internal):
		elasticKeyProvider = cryptoutilOpenapiModel.Internal
	default:
		return nil, fmt.Errorf("invalid elastic Key provider option: %s", provider)
	}
	return &elasticKeyProvider, nil
}

func MapElasticKeyImportAllowed(importAllowed bool) *cryptoutilOpenapiModel.ElasticKeyImportAllowed {
	elasticKeyElasticKeyImportAllowed := cryptoutilOpenapiModel.ElasticKeyImportAllowed(importAllowed)
	return &elasticKeyElasticKeyImportAllowed
}

func MapElasticKeyExportAllowed(exportAllowed bool) *cryptoutilOpenapiModel.ElasticKeyExportAllowed {
	elasticKeyElasticKeyExportAllowed := cryptoutilOpenapiModel.ElasticKeyExportAllowed(exportAllowed)
	return &elasticKeyElasticKeyExportAllowed
}

func MapElasticKeyVersioningAllowed(versioningAllowed bool) *cryptoutilOpenapiModel.ElasticKeyVersioningAllowed {
	elasticKeyElasticKeyVersioningAllowed := cryptoutilOpenapiModel.ElasticKeyVersioningAllowed(versioningAllowed)
	return &elasticKeyElasticKeyVersioningAllowed
}

func ValidateString(value string) error {
	length := len(value)
	trimmedLength := len(strings.TrimSpace(value))
	if length == 0 {
		return fmt.Errorf("string can't be empty")
	} else if trimmedLength == 0 {
		return fmt.Errorf("string can't contain all whitespace")
	} else if trimmedLength != length {
		return fmt.Errorf("string can't contain leading or trailing whitespace")
	}
	return nil
}
