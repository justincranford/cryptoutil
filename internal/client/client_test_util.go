package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	"github.com/stretchr/testify/require"
)

var oamOacMapperInstance = NewOamOacMapper()

// TODO Add error checking for https with rootCAsPool=nil

func WaitUntilReady(baseURL *string, maxTime time.Duration, retryTime time.Duration, rootCAsPool *x509.CertPool) {
	giveUpTime := time.Now().UTC().Add(maxTime)
	for {
		log.Printf("Checking if server is ready %s", *baseURL)
		if err := CheckReadyz(baseURL, rootCAsPool); err == nil {
			log.Printf("Server is ready")
			break
		}
		time.Sleep(retryTime)
		if !time.Now().UTC().Before(giveUpTime) {
			log.Fatalf("server not ready after %v", maxTime)
		}
	}
}

func CheckHealthz(baseURL *string, rootCAsPool *x509.CertPool) error {
	url := *baseURL + "/healthz"
	return httpGet(&url, 2*time.Second, rootCAsPool)
}

func CheckReadyz(baseURL *string, rootCAsPool *x509.CertPool) error {
	url := *baseURL + "/readyz"
	return httpGet(&url, 2*time.Second, rootCAsPool)
}

func httpGet(url *string, timeout time.Duration, rootCAsPool *x509.CertPool) error {
	client := &http.Client{Timeout: timeout}

	if strings.HasPrefix(*url, "https://") {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{
			RootCAs:    rootCAsPool,
			MinVersion: tls.VersionTLS12,
		}}
	}

	resp, err := client.Get(*url)
	if err != nil {
		return fmt.Errorf("get %v failed: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %d", *url, resp.StatusCode)
	}
	return nil
}

func RequireClientWithResponses(t *testing.T, baseUrl *string) *cryptoutilOpenapiClient.ClientWithResponses {
	var openapiClient *cryptoutilOpenapiClient.ClientWithResponses
	var err error

	if strings.HasPrefix(*baseUrl, "https://") {
		// For HTTPS URLs, use proper TLS configuration with full cert chain validation
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
		openapiClient, err = cryptoutilOpenapiClient.NewClientWithResponses(*baseUrl, cryptoutilOpenapiClient.WithHTTPClient(httpClient))
	} else {
		// For HTTP URLs, use default client
		openapiClient, err = cryptoutilOpenapiClient.NewClientWithResponses(*baseUrl)
	}

	require.NoError(t, err)
	require.NotNil(t, openapiClient)
	return openapiClient
}

func RequireCreateElasticKeyRequest(t *testing.T, name *string, description *string, algorithm *string, provider *string, importAllowed *bool, versioningAllowed *bool) *cryptoutilOpenapiModel.ElasticKeyCreate {
	elasticKeyCreate, err := oamOacMapperInstance.toOamElasticKeyCreate(name, description, algorithm, provider, importAllowed, versioningAllowed)
	require.NotNil(t, elasticKeyCreate)
	require.NoError(t, err)
	return elasticKeyCreate
}

func RequireCreateElasticKeyResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate) *cryptoutilOpenapiModel.ElasticKey {
	openapiCreateElasticKeyResponse, err := openapiClient.PostElastickeyWithResponse(context, *elasticKeyCreate)
	require.NoError(t, err)

	elasticKey, err := oamOacMapperInstance.toOamElasticKey(openapiCreateElasticKeyResponse)
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

func RequireMaterialKeyGenerateResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyID, keyGenerate *cryptoutilOpenapiModel.MaterialKeyGenerate) *cryptoutilOpenapiModel.MaterialKey {
	openapiMaterialKeyGenerateResponse, err := openapiClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(context, *elasticKeyId, *keyGenerate)
	require.NoError(t, err)

	key, err := oamOacMapperInstance.toOamMaterialKeyGenerate(openapiMaterialKeyGenerateResponse)
	require.NoError(t, err)

	return key
}

func RequireGenerateResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyID, generateParams *cryptoutilOpenapiModel.GenerateParams) *string {
	elastickeyElasticKeyIDGenerateParams := oamOacMapperInstance.toOacGenerateParams(generateParams)
	// failed to encrypt, nextElasticKeyName(), Status: 400, Message: 400 Bad Request, Body: error in openapi3filter.RequestError: request body has an error: value is required but missing

	openapiGenerateResponse, err := openapiClient.PostElastickeyElasticKeyIDGenerateWithBodyWithResponse(context, *elasticKeyId, &elastickeyElasticKeyIDGenerateParams, "text/plain", nil)
	require.NoError(t, err)

	encrypted, err := oamOacMapperInstance.toPlainGenerateResponse(openapiGenerateResponse)
	require.NoError(t, err)

	return encrypted
}

func RequireEncryptRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	return oamOacMapperInstance.toOamEncryptRequest(cleartext)
}

func RequireEncryptResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyID, encryptParams *cryptoutilOpenapiModel.EncryptParams, encryptRequest *cryptoutilOpenapiModel.EncryptRequest) *string {
	elastickeyElasticKeyIDEncryptParams := oamOacMapperInstance.toOacEncryptParams(encryptParams)
	openapiEncryptResponse, err := openapiClient.PostElastickeyElasticKeyIDEncryptWithTextBodyWithResponse(context, *elasticKeyId, &elastickeyElasticKeyIDEncryptParams, *encryptRequest)
	require.NoError(t, err)

	encrypted, err := oamOacMapperInstance.toPlainEncryptResponse(openapiEncryptResponse)
	require.NoError(t, err)

	return encrypted
}

func RequireDecryptRequest(t *testing.T, ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	return oamOacMapperInstance.toOamDecryptRequest(ciphertext)
}

func RequireDecryptResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyID, decryptRequest *cryptoutilOpenapiModel.DecryptRequest) *string {
	openapiDecryptResponse, err := openapiClient.PostElastickeyElasticKeyIDDecryptWithTextBodyWithResponse(context, *elasticKeyId, *decryptRequest)
	require.NoError(t, err)

	decrypted, err := oamOacMapperInstance.toPlainDecryptResponse(openapiDecryptResponse)
	require.NoError(t, err)

	return decrypted
}

func RequireSignRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	return oamOacMapperInstance.toOamSignRequest(cleartext)
}

func RequireSignResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyID, signParams *cryptoutilOpenapiModel.SignParams, signRequest *cryptoutilOpenapiModel.SignRequest) *string {
	elastickeyElasticKeyIDSignParams := oamOacMapperInstance.toOacSignParams(signParams)
	openapiSignResponse, err := openapiClient.PostElastickeyElasticKeyIDSignWithTextBodyWithResponse(context, *elasticKeyId, &elastickeyElasticKeyIDSignParams, *signRequest)
	require.NoError(t, err)

	signed, err := oamOacMapperInstance.toPlainSignResponse(openapiSignResponse)
	require.NoError(t, err)

	return signed
}

func RequireVerifyRequest(t *testing.T, signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	return oamOacMapperInstance.toOamVerifyRequest(signedtext)
}

func RequireVerifyResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyId *cryptoutilOpenapiModel.ElasticKeyID, verifyRequest *cryptoutilOpenapiModel.VerifyRequest) *string {
	openapiVerifyResponse, err := openapiClient.PostElastickeyElasticKeyIDVerifyWithTextBodyWithResponse(context, *elasticKeyId, *verifyRequest)
	require.NoError(t, err)

	verified, err := oamOacMapperInstance.toPlainVerifyResponse(openapiVerifyResponse)
	require.NoError(t, err)

	return verified
}

func ValidateCreateElasticKeyVsElasticKey(elasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate, elasticKey *cryptoutilOpenapiModel.ElasticKey) error {
	if elasticKeyCreate == nil {
		return fmt.Errorf("elastic Key create is nil")
	} else if elasticKey == nil {
		return fmt.Errorf("elastic Key is nil")
	} else if elasticKey.ElasticKeyID == nil {
		return fmt.Errorf("elastic Key ID is nil")
	} else if elasticKeyCreate.Name != *elasticKey.Name {
		return fmt.Errorf("name mismatch: expected %s, got %s", elasticKeyCreate.Name, *elasticKey.Name)
	} else if elasticKeyCreate.Description != *elasticKey.Description {
		return fmt.Errorf("description mismatch: expected %s, got %s", elasticKeyCreate.Description, *elasticKey.Description)
	} else if *elasticKeyCreate.Algorithm != *elasticKey.Algorithm {
		return fmt.Errorf("algorithm mismatch: expected %s, got %s", *elasticKeyCreate.Algorithm, *elasticKey.Algorithm)
	} else if *elasticKeyCreate.Provider != *elasticKey.Provider {
		return fmt.Errorf("provider mismatch: expected %s, got %s", *elasticKeyCreate.Provider, *elasticKey.Provider)
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
