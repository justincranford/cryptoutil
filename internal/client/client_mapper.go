package client

import (
	"errors"
	"fmt"
	"strings"

	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
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
		return nil, fmt.Errorf("failed to create key pool, Status: %v, Message: %s", openapiCreateKeyPoolResponse.HTTPResponse.StatusCode, openapiCreateKeyPoolResponse.HTTPResponse.Status)
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
		if key.Pool == nil {
			return nil, fmt.Errorf("failed to generate key, keyPool.Pool is nil")
		} else if key.Id == nil {
			return nil, fmt.Errorf("failed to generate key, keyPool.Id is nil")
		} else if key.GenerateDate == nil {
			return nil, fmt.Errorf("failed to generate key, keyPool.GenerateDate is nil")
		}
		return key, nil
	default:
		return nil, fmt.Errorf("failed to generate key, Status: %v, Message: %s", openapiKeyGenerateResponse.HTTPResponse.StatusCode, openapiKeyGenerateResponse.HTTPResponse.Status)
	}
}

func MapSymmetricEncryptParams(symmetricEncryptParams *cryptoutilOpenapiModel.SymmetricEncryptParams) cryptoutilOpenapiClient.PostKeypoolKeyPoolIDEncryptParams {
	keypoolKeyPoolIDEncryptParams := cryptoutilOpenapiClient.PostKeypoolKeyPoolIDEncryptParams{}
	if symmetricEncryptParams != nil {
		keypoolKeyPoolIDEncryptParams.Iv = symmetricEncryptParams.Iv
		keypoolKeyPoolIDEncryptParams.Aad = symmetricEncryptParams.Aad
	}
	return keypoolKeyPoolIDEncryptParams
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
		return nil, fmt.Errorf("failed to encrypt, Status: %v, Message: %s", openapiEncryptResponse.HTTPResponse.StatusCode, openapiEncryptResponse.HTTPResponse.Status)
	}
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
		return nil, fmt.Errorf("failed to decrypt, Status: %v, Message: %s", openapiDecryptResponse.HTTPResponse.StatusCode, openapiDecryptResponse.HTTPResponse.Status)
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
	// TODO Add asymmetric algorithms
	var keyPoolAlgorithm cryptoutilOpenapiModel.KeyPoolAlgorithm
	switch algorithm {

	case string(cryptoutilOpenapiModel.A256GCMA256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMA256KW
	case string(cryptoutilOpenapiModel.A192GCMA256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA256KW
	case string(cryptoutilOpenapiModel.A128GCMA256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA256KW
	case string(cryptoutilOpenapiModel.A192GCMA192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA192KW
	case string(cryptoutilOpenapiModel.A128GCMA192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA192KW
	case string(cryptoutilOpenapiModel.A128GCMA128KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA128KW

	case string(cryptoutilOpenapiModel.A256GCMA256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMA256GCMKW
	case string(cryptoutilOpenapiModel.A192GCMA256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA256GCMKW
	case string(cryptoutilOpenapiModel.A128GCMA256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA256GCMKW
	case string(cryptoutilOpenapiModel.A192GCMA192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA192GCMKW
	case string(cryptoutilOpenapiModel.A128GCMA192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA192GCMKW
	case string(cryptoutilOpenapiModel.A128GCMA128GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA128GCMKW

	case string(cryptoutilOpenapiModel.A256GCMdir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMdir
	case string(cryptoutilOpenapiModel.A192GCMdir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMdir
	case string(cryptoutilOpenapiModel.A128GCMdir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMdir

	case string(cryptoutilOpenapiModel.A256CBCHS512A256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512A256KW
	case string(cryptoutilOpenapiModel.A192CBCHS384A256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A256KW
	case string(cryptoutilOpenapiModel.A128CBCHS256A256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A256KW
	case string(cryptoutilOpenapiModel.A192CBCHS384A192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A192KW
	case string(cryptoutilOpenapiModel.A128CBCHS256A192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A192KW
	case string(cryptoutilOpenapiModel.A128CBCHS256A128KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A128KW

	case string(cryptoutilOpenapiModel.A256CBCHS512A256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512A256GCMKW
	case string(cryptoutilOpenapiModel.A192CBCHS384A256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A256GCMKW
	case string(cryptoutilOpenapiModel.A128CBCHS256A256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A256GCMKW
	case string(cryptoutilOpenapiModel.A192CBCHS384A192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A192GCMKW
	case string(cryptoutilOpenapiModel.A128CBCHS256A192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A192GCMKW
	case string(cryptoutilOpenapiModel.A128CBCHS256A128GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A128GCMKW

	case string(cryptoutilOpenapiModel.A256CBCHS512dir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512dir
	case string(cryptoutilOpenapiModel.A192CBCHS384dir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384dir
	case string(cryptoutilOpenapiModel.A128CBCHS256dir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256dir

	case string(cryptoutilOpenapiModel.A256GCMRSAOAEP512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMRSAOAEP512
	case string(cryptoutilOpenapiModel.A192GCMRSAOAEP512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMRSAOAEP512
	case string(cryptoutilOpenapiModel.A128GCMRSAOAEP512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMRSAOAEP512
	case string(cryptoutilOpenapiModel.A256GCMRSAOAEP384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMRSAOAEP384
	case string(cryptoutilOpenapiModel.A192GCMRSAOAEP384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMRSAOAEP384
	case string(cryptoutilOpenapiModel.A128GCMRSAOAEP384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMRSAOAEP256
	case string(cryptoutilOpenapiModel.A256GCMRSAOAEP256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMRSAOAEP256
	case string(cryptoutilOpenapiModel.A192GCMRSAOAEP256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMRSAOAEP256
	case string(cryptoutilOpenapiModel.A128GCMRSAOAEP256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMRSAOAEP256
	case string(cryptoutilOpenapiModel.A256GCMRSAOAEP):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMRSAOAEP
	case string(cryptoutilOpenapiModel.A192GCMRSAOAEP):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMRSAOAEP
	case string(cryptoutilOpenapiModel.A128GCMRSAOAEP):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMRSAOAEP
	case string(cryptoutilOpenapiModel.A256GCMRSA15):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMRSA15
	case string(cryptoutilOpenapiModel.A192GCMRSA15):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMRSA15
	case string(cryptoutilOpenapiModel.A128GCMRSA15):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMRSA15

	case string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMRSAOAEP512
	case string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384RSAOAEP512
	case string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256RSAOAEP512
	case string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512RSAOAEP384
	case string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384RSAOAEP384
	case string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256RSAOAEP384
	case string(cryptoutilOpenapiModel.A256CBCHS256RSAOAEP256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS256RSAOAEP256
	case string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384RSAOAEP256
	case string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256
	case string(cryptoutilOpenapiModel.A256CBCHS512RSAOAEP):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512RSAOAEP
	case string(cryptoutilOpenapiModel.A192CBCHS384RSAOAEP):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384RSAOAEP
	case string(cryptoutilOpenapiModel.A128CBCHS256RSAOAEP):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256RSAOAEP
	case string(cryptoutilOpenapiModel.A256CBCHS512RSA15):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512RSA15
	case string(cryptoutilOpenapiModel.A192CBCHS384RSA15):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384RSA15
	case string(cryptoutilOpenapiModel.A128CBCHS256RSA15):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256RSA15

	case string(cryptoutilOpenapiModel.RS256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.RS256
	case string(cryptoutilOpenapiModel.RS384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.RS384
	case string(cryptoutilOpenapiModel.RS512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.RS512
	case string(cryptoutilOpenapiModel.PS256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.PS256
	case string(cryptoutilOpenapiModel.PS384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.PS384
	case string(cryptoutilOpenapiModel.PS512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.PS512
	case string(cryptoutilOpenapiModel.ES256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.ES256
	case string(cryptoutilOpenapiModel.ES384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.ES384
	case string(cryptoutilOpenapiModel.ES512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.ES512
	case string(cryptoutilOpenapiModel.HS256):
		keyPoolAlgorithm = cryptoutilOpenapiModel.HS256
	case string(cryptoutilOpenapiModel.HS384):
		keyPoolAlgorithm = cryptoutilOpenapiModel.HS384
	case string(cryptoutilOpenapiModel.HS512):
		keyPoolAlgorithm = cryptoutilOpenapiModel.HS512
	case string(cryptoutilOpenapiModel.EdDSA):
		keyPoolAlgorithm = cryptoutilOpenapiModel.EdDSA

	default:
		return nil, fmt.Errorf("invalid key pool algorithm: %s", algorithm)
	}
	return &keyPoolAlgorithm, nil
}

func MapKeyPoolProvider(provider string) (*cryptoutilOpenapiModel.KeyPoolProvider, error) {
	if err := ValidateString(provider); err != nil {
		return nil, fmt.Errorf("invalid key pool provider: %w", err)
	}
	var keyPoolProvider cryptoutilOpenapiModel.KeyPoolProvider
	switch provider {
	case string(cryptoutilOpenapiModel.Internal):
		keyPoolProvider = cryptoutilOpenapiModel.Internal
	default:
		return nil, fmt.Errorf("invalid key pool provider: %s", provider)
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
