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
	validAlgorithms = map[string]cryptoutilOpenapiModel.KeyPoolAlgorithm{
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

func MapKeyPoolCreate(name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) (*cryptoutilOpenapiModel.KeyPoolCreate, error) {
	keyPoolName, err1 := MapKeyPoolName(name)
	keyPoolDescription, err2 := MapKeyPoolDescription(description)
	keyPoolAlgorithm, err3 := MapKeyPoolAlgorithm(algorithm)
	keyPoolProvider, err4 := MapKeyPoolProvider(provider)
	keyPoolKeyPoolExportAllowed := MapKeyPoolExportAllowed(exportAllowed)
	keyPoolKeyPoolImportAllowed := MapKeyPoolImportAllowed(importAllowed)
	keyPoolKeyPoolVersioningAllowed := MapKeyPoolVersioningAllowed(versioningAllowed)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, fmt.Errorf("failed to map key pool: %v", errors.Join(err1, err2, err3, err4))
	}
	return &cryptoutilOpenapiModel.KeyPoolCreate{
		Name:              *keyPoolName,
		Description:       *keyPoolDescription,
		Provider:          keyPoolProvider,
		Algorithm:         keyPoolAlgorithm,
		ExportAllowed:     keyPoolKeyPoolExportAllowed,
		ImportAllowed:     keyPoolKeyPoolImportAllowed,
		VersioningAllowed: keyPoolKeyPoolVersioningAllowed,
	}, nil
}

func MapKeyPool(openapiCreateKeyPoolResponse *cryptoutilOpenapiClient.PostKeypoolResponse) (*cryptoutilOpenapiModel.KeyPool, error) {
	if openapiCreateKeyPoolResponse == nil {
		return nil, fmt.Errorf("failed to create key pool, response is nil")
	} else if openapiCreateKeyPoolResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to create key pool, HTTP response is nil")
	}
	switch openapiCreateKeyPoolResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiCreateKeyPoolResponse.Body == nil {
			return nil, fmt.Errorf("failed to create key pool, body is nil")
		} else if openapiCreateKeyPoolResponse.JSON200 == nil {
			return nil, fmt.Errorf("failed to create key pool, JSON200 is nil")
		}
		keyPool := openapiCreateKeyPoolResponse.JSON200
		if keyPool.Id == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.Id is nil")
		} else if keyPool.Description == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.Description is nil")
		} else if keyPool.Algorithm == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.Algorithm is nil")
		} else if keyPool.Provider == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.Provider is nil")
		} else if keyPool.ExportAllowed == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.ExportAllowed is nil")
		} else if keyPool.ImportAllowed == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.ImportAllowed is nil")
		} else if keyPool.VersioningAllowed == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.VersioningAllowed is nil")
		} else if keyPool.Status == nil {
			return nil, fmt.Errorf("failed to create key pool, keyPool.Status is nil")
		}
		return keyPool, nil
	default:
		return nil, fmt.Errorf("failed to create key pool, nextKeyPoolName(), Status: %v, Message: %s, Body: %s", openapiCreateKeyPoolResponse.HTTPResponse.StatusCode, openapiCreateKeyPoolResponse.HTTPResponse.Status, openapiCreateKeyPoolResponse.Body)
	}
}

func MapKeyGenerate(openapiKeyGenerateResponse *cryptoutilOpenapiClient.PostKeypoolKeyPoolIDKeyResponse) (*cryptoutilOpenapiModel.Key, error) {
	if openapiKeyGenerateResponse == nil {
		return nil, fmt.Errorf("failed to generate key, response is nil")
	} else if openapiKeyGenerateResponse.HTTPResponse == nil {
		return nil, fmt.Errorf("failed to generate key, HTTP response is nil")
	}
	switch openapiKeyGenerateResponse.HTTPResponse.StatusCode {
	case 200:
		if openapiKeyGenerateResponse.Body == nil {
			return nil, fmt.Errorf("failed to generate key, body is nil")
		} else if openapiKeyGenerateResponse.JSON200 == nil {
			return nil, fmt.Errorf("failed to generate key, JSON200 is nil")
		}
		key := openapiKeyGenerateResponse.JSON200
		if key.Pool == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, keyPool.Pool is zero")
		} else if key.Id == googleUuid.Nil {
			return nil, fmt.Errorf("failed to generate key, keyPool.Id is zero")
		} else if key.GenerateDate == nil {
			return nil, fmt.Errorf("failed to generate key, keyPool.GenerateDate is nil") // TODO nil allowed if import not nil
		}
		return key, nil
	default:
		return nil, fmt.Errorf("failed to generate key, nextKeyPoolName(), Status: %v, Message: %s, Body: %s", openapiKeyGenerateResponse.HTTPResponse.StatusCode, openapiKeyGenerateResponse.HTTPResponse.Status, openapiKeyGenerateResponse.Body)
	}
}

func MapEncryptParams(encryptParams *cryptoutilOpenapiModel.EncryptParams) cryptoutilOpenapiClient.PostKeypoolKeyPoolIDEncryptParams {
	keypoolKeyPoolIDEncryptParams := cryptoutilOpenapiClient.PostKeypoolKeyPoolIDEncryptParams{}
	if encryptParams != nil {
		keypoolKeyPoolIDEncryptParams.Context = encryptParams.Context
	}
	return keypoolKeyPoolIDEncryptParams
}

func MapEncryptRequest(cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	encryptRequest := cryptoutilOpenapiModel.EncryptRequest(*cleartext)
	return &encryptRequest
}

func MapEncryptResponse(openapiEncryptResponse *cryptoutilOpenapiClient.PostKeypoolKeyPoolIDEncryptResponse) (*string, error) {
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
		return nil, fmt.Errorf("failed to encrypt, nextKeyPoolName(), Status: %v, Message: %s, Body: %s", openapiEncryptResponse.HTTPResponse.StatusCode, openapiEncryptResponse.HTTPResponse.Status, openapiEncryptResponse.Body)
	}
}

func MapDecryptRequest(ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	decryptRequest := cryptoutilOpenapiModel.DecryptRequest(*ciphertext)
	return &decryptRequest
}

func MapDecryptResponse(openapiDecryptResponse *cryptoutilOpenapiClient.PostKeypoolKeyPoolIDDecryptResponse) (*string, error) {
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
		return nil, fmt.Errorf("failed to decrypt, nextKeyPoolName(), Status: %v, Message: %s, Body: %s", openapiDecryptResponse.HTTPResponse.StatusCode, openapiDecryptResponse.HTTPResponse.Status, openapiDecryptResponse.Body)
	}
}

func MapSignParams(signParams *cryptoutilOpenapiModel.SignParams) cryptoutilOpenapiClient.PostKeypoolKeyPoolIDSignParams {
	keypoolKeyPoolIDSignParams := cryptoutilOpenapiClient.PostKeypoolKeyPoolIDSignParams{}
	if signParams != nil {
		keypoolKeyPoolIDSignParams.Context = signParams.Context
	}
	return keypoolKeyPoolIDSignParams
}

func MapSignRequest(cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	signRequest := cryptoutilOpenapiModel.SignRequest(*cleartext)
	return &signRequest
}

func MapSignResponse(openapiSignResponse *cryptoutilOpenapiClient.PostKeypoolKeyPoolIDSignResponse) (*string, error) {
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
		return nil, fmt.Errorf("failed to sign, nextKeyPoolName(), Status: %v, Message: %s, Body: %s", openapiSignResponse.HTTPResponse.StatusCode, openapiSignResponse.HTTPResponse.Status, openapiSignResponse.Body)
	}
}

func MapVerifyRequest(signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	verifyRequest := cryptoutilOpenapiModel.VerifyRequest(*signedtext)
	return &verifyRequest
}

func MapVerifyResponse(openapiVerifyResponse *cryptoutilOpenapiClient.PostKeypoolKeyPoolIDVerifyResponse) (*string, error) {
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
		return nil, fmt.Errorf("failed to verify, nextKeyPoolName(), Status: %v, Message: %s, Body: %s", openapiVerifyResponse.HTTPResponse.StatusCode, openapiVerifyResponse.HTTPResponse.Status, openapiVerifyResponse.Body)
	}
}

func MapKeyGenerater() (*cryptoutilOpenapiClient.PostKeypoolKeyPoolIDKeyJSONRequestBody, error) {
	return &cryptoutilOpenapiModel.KeyGenerate{}, nil
}

func MapKeyPoolName(name string) (*cryptoutilOpenapiModel.KeyPoolName, error) {
	if err := ValidateString(name); err != nil {
		return nil, fmt.Errorf("invalid key pool name: %w", err)
	}
	keyPoolName := cryptoutilOpenapiModel.KeyPoolName(name)
	return &keyPoolName, nil
}

func MapKeyPoolDescription(description string) (*cryptoutilOpenapiModel.KeyPoolDescription, error) {
	if err := ValidateString(description); err != nil {
		return nil, fmt.Errorf("invalid key pool description: %w", err)
	}
	keyPoolDescription := cryptoutilOpenapiModel.KeyPoolDescription(description)
	return &keyPoolDescription, nil
}

func MapKeyPoolAlgorithm(algorithm string) (*cryptoutilOpenapiModel.KeyPoolAlgorithm, error) {
	if err := ValidateString(algorithm); err != nil {
		return nil, fmt.Errorf("invalid key pool algorithm: %w", err)
	}
	if alg, exists := validAlgorithms[algorithm]; exists {
		return &alg, nil
	}
	return nil, fmt.Errorf("invalid key pool algorithm: %s", algorithm)
}

func MapKeyPoolProvider(provider string) (*cryptoutilOpenapiModel.KeyPoolProvider, error) {
	if err := ValidateString(provider); err != nil {
		return nil, fmt.Errorf("invalid key pool provider value: %w", err)
	}
	var keyPoolProvider cryptoutilOpenapiModel.KeyPoolProvider
	switch provider {
	case string(cryptoutilOpenapiModel.Internal):
		keyPoolProvider = cryptoutilOpenapiModel.Internal
	default:
		return nil, fmt.Errorf("invalid key pool provider option: %s", provider)
	}
	return &keyPoolProvider, nil
}

func MapKeyPoolImportAllowed(importAllowed bool) *cryptoutilOpenapiModel.KeyPoolImportAllowed {
	keyPoolKeyPoolImportAllowed := cryptoutilOpenapiModel.KeyPoolImportAllowed(importAllowed)
	return &keyPoolKeyPoolImportAllowed
}

func MapKeyPoolExportAllowed(exportAllowed bool) *cryptoutilOpenapiModel.KeyPoolExportAllowed {
	keyPoolKeyPoolExportAllowed := cryptoutilOpenapiModel.KeyPoolExportAllowed(exportAllowed)
	return &keyPoolKeyPoolExportAllowed
}

func MapKeyPoolVersioningAllowed(versioningAllowed bool) *cryptoutilOpenapiModel.KeyPoolVersioningAllowed {
	keyPoolKeyPoolVersioningAllowed := cryptoutilOpenapiModel.KeyPoolVersioningAllowed(versioningAllowed)
	return &keyPoolKeyPoolVersioningAllowed
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
