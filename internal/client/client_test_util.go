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

func RequireCreateElasticKeyRequest(t *testing.T, name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) *cryptoutilOpenapiModel.ElasticKeyCreate {
	elasticKeyCreate, err := MapElasticKeyCreate(name, description, algorithm, provider, exportAllowed, importAllowed, versioningAllowed)
	require.NotNil(t, elasticKeyCreate)
	require.NoError(t, err)
	return elasticKeyCreate
}

func RequireCreateElasticKeyResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate) *cryptoutilOpenapiModel.ElasticKey {
	openapiCreateElasticKeyResponse, err := openapiClient.PostElastickeyWithResponse(context, cryptoutilOpenapiClient.PostElastickeyJSONRequestBody(*elasticKeyCreate))
	require.NoError(t, err)

	elasticKey, err := MapElasticKey(openapiCreateElasticKeyResponse)
	require.NoError(t, err)
	require.NotNil(t, elasticKey)

	err = ValidateCreateElasticKeyVsElasticKey(elasticKeyCreate, elasticKey)
	require.NoError(t, err)

	return elasticKey
}

// TODO Support generate settings (e.g. expiration)
func RequireMaterialKeyGenerateRequest(t *testing.T) *cryptoutilOpenapiModel.MaterialKeyGenerate {
	keyGenerate := cryptoutilOpenapiModel.MaterialKeyGenerate{}
	return &keyGenerate
}

func RequireMaterialKeyGenerateResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyId, keyGenerate *cryptoutilOpenapiModel.MaterialKeyGenerate) *cryptoutilOpenapiModel.MaterialKey {
	openapiMaterialKeyGenerateResponse, err := openapiClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(context, *elasticKeyId, *keyGenerate)
	require.NoError(t, err)

	key, err := MapMaterialKeyGenerate(openapiMaterialKeyGenerateResponse)
	require.NoError(t, err)

	return key
}

func RequireEncryptRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	return MapEncryptRequest(cleartext)
}

func RequireEncryptResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyId, encryptParams *cryptoutilOpenapiModel.EncryptParams, encryptRequest *cryptoutilOpenapiModel.EncryptRequest) *string {
	elastickeyElasticKeyIDEncryptParams := MapEncryptParams(encryptParams)
	openapiEncryptResponse, err := openapiClient.PostElastickeyElasticKeyIDEncryptWithTextBodyWithResponse(context, *elasticKeyId, &elastickeyElasticKeyIDEncryptParams, *encryptRequest)
	require.NoError(t, err)

	encrypted, err := MapEncryptResponse(openapiEncryptResponse)
	require.NoError(t, err)

	return encrypted
}

func RequireDecryptRequest(t *testing.T, ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	return MapDecryptRequest(ciphertext)
}

func RequireDecryptResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyId, decryptRequest *cryptoutilOpenapiModel.DecryptRequest) *string {
	openapiDecryptResponse, err := openapiClient.PostElastickeyElasticKeyIDDecryptWithTextBodyWithResponse(context, *elasticKeyId, *decryptRequest)
	require.NoError(t, err)

	decrypted, err := MapDecryptResponse(openapiDecryptResponse)
	require.NoError(t, err)

	return decrypted
}

func RequireSignRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	return MapSignRequest(cleartext)
}

func RequireSignResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyId, signParams *cryptoutilOpenapiModel.SignParams, signRequest *cryptoutilOpenapiModel.SignRequest) *string {
	elastickeyElasticKeyIDSignParams := MapSignParams(signParams)
	openapiSignResponse, err := openapiClient.PostElastickeyElasticKeyIDSignWithTextBodyWithResponse(context, *elasticKeyId, &elastickeyElasticKeyIDSignParams, *signRequest)
	require.NoError(t, err)

	signed, err := MapSignResponse(openapiSignResponse)
	require.NoError(t, err)

	return signed
}

func RequireVerifyRequest(t *testing.T, signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	return MapVerifyRequest(signedtext)
}

func RequireVerifyResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyId, verifyRequest *cryptoutilOpenapiModel.VerifyRequest) *string {
	openapiVerifyResponse, err := openapiClient.PostElastickeyElasticKeyIDVerifyWithTextBodyWithResponse(context, *elasticKeyId, *verifyRequest)
	require.NoError(t, err)

	verified, err := MapVerifyResponse(openapiVerifyResponse)
	require.NoError(t, err)

	return verified
}

func ValidateCreateElasticKeyVsElasticKey(elasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate, elasticKey *cryptoutilOpenapiModel.ElasticKey) error {
	if elasticKeyCreate == nil {
		return fmt.Errorf("elastic Key create is nil")
	} else if elasticKey == nil {
		return fmt.Errorf("elastic Key is nil")
	} else if elasticKey.ElasticKeyId == nil {
		return fmt.Errorf("elastic Key ID is nil")
	} else if elasticKeyCreate.Name != *elasticKey.Name {
		return fmt.Errorf("name mismatch: expected %s, got %s", elasticKeyCreate.Name, *elasticKey.Name)
	} else if elasticKeyCreate.Description != *elasticKey.Description {
		return fmt.Errorf("description mismatch: expected %s, got %s", elasticKeyCreate.Description, *elasticKey.Description)
	} else if *elasticKeyCreate.Algorithm != *elasticKey.Algorithm {
		return fmt.Errorf("algorithm mismatch: expected %s, got %s", *elasticKeyCreate.Algorithm, *elasticKey.Algorithm)
	} else if *elasticKeyCreate.Provider != *elasticKey.Provider {
		return fmt.Errorf("provider mismatch: expected %s, got %s", *elasticKeyCreate.Provider, *elasticKey.Provider)
	} else if *elasticKeyCreate.ExportAllowed != *elasticKey.ExportAllowed {
		return fmt.Errorf("exportAllowed mismatch: expected %t, got %t", *elasticKeyCreate.ExportAllowed, *elasticKey.ExportAllowed)
	} else if *elasticKeyCreate.ImportAllowed != *elasticKey.ImportAllowed {
		return fmt.Errorf("importAllowed mismatch: expected %t, got %t", *elasticKeyCreate.ImportAllowed, *elasticKey.ImportAllowed)
	} else if *elasticKeyCreate.VersioningAllowed != *elasticKey.VersioningAllowed {
		return fmt.Errorf("versioningAllowed mismatch: expected %t, got %t", *elasticKeyCreate.VersioningAllowed, *elasticKey.VersioningAllowed)
	} else if cryptoutilOpenapiModel.Active != *elasticKey.Status {
		return fmt.Errorf("status mismatch: expected %s, got %s", cryptoutilOpenapiModel.Active, *elasticKey.Status)
	}
	if *elasticKey.ImportAllowed {
		if cryptoutilOpenapiModel.PendingImport != *elasticKey.Status {
			return fmt.Errorf("status mismatch: expected %v, got %v", cryptoutilOpenapiModel.PendingImport, *elasticKey.Status)
		}
	} else {
		if cryptoutilOpenapiModel.Active != *elasticKey.Status {
			return fmt.Errorf("status mismatch: expected %v, got %v", cryptoutilOpenapiModel.Active, *elasticKey.Status)
		}
	}
	return nil
}
