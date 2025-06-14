package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	"github.com/stretchr/testify/require"
)

func WaitUntilReady(baseURL string, maxTime time.Duration, retryTime time.Duration) {
	giveUpTime := time.Now().UTC().Add(maxTime)
	for {
		log.Printf("Checking if server is ready")
		if err := CheckReadyz(baseURL); err == nil {
			log.Printf("Server is ready")
			break
		}
		time.Sleep(retryTime)
		if !time.Now().UTC().Before(giveUpTime) {
			log.Fatalf("server not ready after %v", maxTime)
		}
	}
}

func CheckHealthz(baseURL string) error {
	return httpGet(baseURL+"healthz", 2*time.Second)
}

func CheckReadyz(baseURL string) error {
	return httpGet(baseURL+"readyz", 2*time.Second)
}

func httpGet(url string, timeout time.Duration) error {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("get %s failed: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %d", url, resp.StatusCode)
	}
	return nil
}

func RequireClientWithResponses(t *testing.T, baseUrl string) *cryptoutilOpenapiClient.ClientWithResponses {
	openapiClient, err := cryptoutilOpenapiClient.NewClientWithResponses(baseUrl)
	require.NoError(t, err)
	require.NotNil(t, openapiClient)
	return openapiClient
}

func RequireCreateKeyPoolRequest(t *testing.T, name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) *cryptoutilOpenapiModel.KeyPoolCreate {
	keyPoolCreate, err := MapKeyPoolCreate(name, description, algorithm, provider, exportAllowed, importAllowed, versioningAllowed)
	require.NotNil(t, keyPoolCreate)
	require.NoError(t, err)
	return keyPoolCreate
}

func RequireCreateKeyPoolResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolCreate *cryptoutilOpenapiModel.KeyPoolCreate) *cryptoutilOpenapiModel.KeyPool {
	openapiCreateKeyPoolResponse, err := openapiClient.PostKeypoolWithResponse(context, cryptoutilOpenapiClient.PostKeypoolJSONRequestBody(*keyPoolCreate))
	require.NoError(t, err)

	keyPool, err := MapKeyPool(openapiCreateKeyPoolResponse)
	require.NoError(t, err)
	require.NotNil(t, keyPool)

	err = ValidateCreateKeyPoolVsKeyPool(keyPoolCreate, keyPool)
	require.NoError(t, err)

	return keyPool
}

// TODO Support generate settings (e.g. expiration)
func RequireKeyGenerateRequest(t *testing.T) *cryptoutilOpenapiModel.KeyGenerate {
	keyGenerate := cryptoutilOpenapiModel.KeyGenerate{}
	return &keyGenerate
}

func RequireKeyGenerateResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolId *cryptoutilOpenapiModel.KeyPoolId, keyGenerate *cryptoutilOpenapiModel.KeyGenerate) *cryptoutilOpenapiModel.Key {
	openapiKeyGenerateResponse, err := openapiClient.PostKeypoolKeyPoolIDKeyWithResponse(context, *keyPoolId, *keyGenerate)
	require.NoError(t, err)

	key, err := MapKeyGenerate(openapiKeyGenerateResponse)
	require.NoError(t, err)

	return key
}

func RequireEncryptRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	return MapEncryptRequest(cleartext)
}

func RequireEncryptResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolId *cryptoutilOpenapiModel.KeyPoolId, encryptParams *cryptoutilOpenapiModel.EncryptParams, encryptRequest *cryptoutilOpenapiModel.EncryptRequest) *string {
	keypoolKeyPoolIDEncryptParams := MapEncryptParams(encryptParams)
	openapiEncryptResponse, err := openapiClient.PostKeypoolKeyPoolIDEncryptWithTextBodyWithResponse(context, *keyPoolId, &keypoolKeyPoolIDEncryptParams, *encryptRequest)
	require.NoError(t, err)

	encrypted, err := MapEncryptResponse(openapiEncryptResponse)
	require.NoError(t, err)

	return encrypted
}

func RequireDecryptRequest(t *testing.T, ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	return MapDecryptRequest(ciphertext)
}

func RequireDecryptResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolId *cryptoutilOpenapiModel.KeyPoolId, decryptRequest *cryptoutilOpenapiModel.DecryptRequest) *string {
	openapiDecryptResponse, err := openapiClient.PostKeypoolKeyPoolIDDecryptWithTextBodyWithResponse(context, *keyPoolId, *decryptRequest)
	require.NoError(t, err)

	decrypted, err := MapDecryptResponse(openapiDecryptResponse)
	require.NoError(t, err)

	return decrypted
}

func RequireSignRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	return MapSignRequest(cleartext)
}

func RequireSignResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolId *cryptoutilOpenapiModel.KeyPoolId, signParams *cryptoutilOpenapiModel.SignParams, signRequest *cryptoutilOpenapiModel.SignRequest) *string {
	keypoolKeyPoolIDSignParams := MapSignParams(signParams)
	openapiSignResponse, err := openapiClient.PostKeypoolKeyPoolIDSignWithTextBodyWithResponse(context, *keyPoolId, &keypoolKeyPoolIDSignParams, *signRequest)
	require.NoError(t, err)

	signed, err := MapSignResponse(openapiSignResponse)
	require.NoError(t, err)

	return signed
}

func RequireVerifyRequest(t *testing.T, signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	return MapVerifyRequest(signedtext)
}

func RequireVerifyResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolId *cryptoutilOpenapiModel.KeyPoolId, verifyRequest *cryptoutilOpenapiModel.VerifyRequest) *string {
	openapiVerifyResponse, err := openapiClient.PostKeypoolKeyPoolIDVerifyWithTextBodyWithResponse(context, *keyPoolId, *verifyRequest)
	require.NoError(t, err)

	verified, err := MapVerifyResponse(openapiVerifyResponse)
	require.NoError(t, err)

	return verified
}

func ValidateCreateKeyPoolVsKeyPool(keyPoolCreate *cryptoutilOpenapiModel.KeyPoolCreate, keyPool *cryptoutilOpenapiModel.KeyPool) error {
	if keyPoolCreate == nil {
		return fmt.Errorf("key pool create is nil")
	} else if keyPool == nil {
		return fmt.Errorf("key pool is nil")
	} else if keyPool.Id == nil {
		return fmt.Errorf("key pool ID is nil")
	} else if keyPoolCreate.Name != *keyPool.Name {
		return fmt.Errorf("name mismatch: expected %s, got %s", keyPoolCreate.Name, *keyPool.Name)
	} else if keyPoolCreate.Description != *keyPool.Description {
		return fmt.Errorf("description mismatch: expected %s, got %s", keyPoolCreate.Description, *keyPool.Description)
	} else if *keyPoolCreate.Algorithm != *keyPool.Algorithm {
		return fmt.Errorf("algorithm mismatch: expected %s, got %s", *keyPoolCreate.Algorithm, *keyPool.Algorithm)
	} else if *keyPoolCreate.Provider != *keyPool.Provider {
		return fmt.Errorf("provider mismatch: expected %s, got %s", *keyPoolCreate.Provider, *keyPool.Provider)
	} else if *keyPoolCreate.ExportAllowed != *keyPool.ExportAllowed {
		return fmt.Errorf("exportAllowed mismatch: expected %t, got %t", *keyPoolCreate.ExportAllowed, *keyPool.ExportAllowed)
	} else if *keyPoolCreate.ImportAllowed != *keyPool.ImportAllowed {
		return fmt.Errorf("importAllowed mismatch: expected %t, got %t", *keyPoolCreate.ImportAllowed, *keyPool.ImportAllowed)
	} else if *keyPoolCreate.VersioningAllowed != *keyPool.VersioningAllowed {
		return fmt.Errorf("versioningAllowed mismatch: expected %t, got %t", *keyPoolCreate.VersioningAllowed, *keyPool.VersioningAllowed)
	} else if cryptoutilOpenapiModel.Active != *keyPool.Status {
		return fmt.Errorf("status mismatch: expected %s, got %s", cryptoutilOpenapiModel.Active, *keyPool.Status)
	}
	if *keyPool.ImportAllowed {
		if cryptoutilOpenapiModel.PendingImport != *keyPool.Status {
			return fmt.Errorf("status mismatch: expected %v, got %v", cryptoutilOpenapiModel.PendingImport, *keyPool.Status)
		}
	} else {
		if cryptoutilOpenapiModel.Active != *keyPool.Status {
			return fmt.Errorf("status mismatch: expected %v, got %v", cryptoutilOpenapiModel.Active, *keyPool.Status)
		}
	}
	return nil
}
