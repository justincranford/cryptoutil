// Copyright (c) 2025 Justin Cranford
//
//

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	http "net/http"
	"strings"
	"testing"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

var oamOacMapperInstance = NewOamOacMapper()

// WaitUntilReady waits for the server at baseURL to become ready.
func WaitUntilReady(baseURL *string, maxTime, retryTime time.Duration, rootCAsPool *x509.CertPool) {
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

// CheckHealthz checks the server health endpoint.
func CheckHealthz(baseURL *string, rootCAsPool *x509.CertPool) error {
	url := *baseURL + cryptoutilSharedMagic.PrivateAdminLivezRequestPath

	return httpGet(&url, cryptoutilSharedMagic.TimeoutHTTPHealthRequest, rootCAsPool)
}

// CheckReadyz checks the server readiness endpoint.
func CheckReadyz(baseURL *string, rootCAsPool *x509.CertPool) error {
	url := *baseURL + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath

	return httpGet(&url, cryptoutilSharedMagic.TimeoutHTTPHealthRequest, rootCAsPool)
}

func httpGet(url *string, timeout time.Duration, rootCAsPool *x509.CertPool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{}

	if strings.HasPrefix(*url, "https://") {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if rootCAsPool != nil {
			// Use provided root CAs for certificate validation
			tlsConfig.RootCAs = rootCAsPool
		} else {
			// No root CAs provided - skip verification for self-signed certificates
			tlsConfig.InsecureSkipVerify = true //nolint:gosec // G402: TLS InsecureSkipVerify set true for testing with self-signed certs
		}

		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
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

// RequireClientWithResponses creates and returns an OpenAPI client for testing.
func RequireClientWithResponses(t *testing.T, baseURL *string, rootCAsPool *x509.CertPool) *cryptoutilOpenapiClient.ClientWithResponses {
	t.Helper()

	var openapiClient *cryptoutilOpenapiClient.ClientWithResponses

	var err error

	if strings.HasPrefix(*baseURL, "https://") {
		// For HTTPS URLs, configure TLS
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if rootCAsPool != nil {
			// Use provided root CAs for certificate validation
			tlsConfig.RootCAs = rootCAsPool
		} else {
			// No root CAs provided - skip verification for self-signed certificates
			tlsConfig.InsecureSkipVerify = true //nolint:gosec // G402: TLS InsecureSkipVerify set true for testing with self-signed certs
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
		openapiClient, err = cryptoutilOpenapiClient.NewClientWithResponses(*baseURL, cryptoutilOpenapiClient.WithHTTPClient(httpClient))
	} else {
		// For HTTP URLs, use default client
		openapiClient, err = cryptoutilOpenapiClient.NewClientWithResponses(*baseURL)
	}

	require.NoError(t, err)
	require.NotNil(t, openapiClient)

	return openapiClient
}

// RequireCreateElasticKeyRequest creates an ElasticKeyCreate request for testing.
func RequireCreateElasticKeyRequest(t *testing.T, name, description, algorithm, provider *string, importAllowed, versioningAllowed *bool) *cryptoutilOpenapiModel.ElasticKeyCreate {
	t.Helper()

	elasticKeyCreate, err := oamOacMapperInstance.toOamElasticKeyCreate(name, description, algorithm, provider, importAllowed, versioningAllowed)
	require.NotNil(t, elasticKeyCreate)
	require.NoError(t, err)

	return elasticKeyCreate
}

// RequireCreateElasticKeyResponse creates an ElasticKey and returns the response.
func RequireCreateElasticKeyResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate) *cryptoutilOpenapiModel.ElasticKey {
	t.Helper()

	openapiCreateElasticKeyResponse, err := openapiClient.PostElastickeyWithResponse(ctx, *elasticKeyCreate)
	require.NoError(t, err)

	elasticKey, err := oamOacMapperInstance.toOamElasticKey(openapiCreateElasticKeyResponse)
	require.NoError(t, err)
	require.NotNil(t, elasticKey)

	err = ValidateCreateElasticKeyVsElasticKey(elasticKeyCreate, elasticKey)
	require.NoError(t, err)

	return elasticKey
}

// RequireMaterialKeyGenerateRequest creates a MaterialKeyGenerate request for testing.
func RequireMaterialKeyGenerateRequest(t *testing.T) *cryptoutilOpenapiModel.MaterialKeyGenerate {
	t.Helper()

	keyGenerate := cryptoutilOpenapiModel.MaterialKeyGenerate{}

	return &keyGenerate
}

// RequireMaterialKeyGenerateResponse generates a MaterialKey and returns the response.
func RequireMaterialKeyGenerateResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyID *cryptoutilOpenapiModel.ElasticKeyID, keyGenerate *cryptoutilOpenapiModel.MaterialKeyGenerate) *cryptoutilOpenapiModel.MaterialKey {
	t.Helper()

	openapiMaterialKeyGenerateResponse, err := openapiClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(ctx, *elasticKeyID, *keyGenerate)
	require.NoError(t, err)

	key, err := oamOacMapperInstance.toOamMaterialKeyGenerate(openapiMaterialKeyGenerateResponse)
	require.NoError(t, err)

	return key
}

// RequireGenerateResponse generates cryptographic material and returns the response.
func RequireGenerateResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyID *cryptoutilOpenapiModel.ElasticKeyID, generateParams *cryptoutilOpenapiModel.GenerateParams) *string {
	t.Helper()

	elastickeyElasticKeyIDGenerateParams := oamOacMapperInstance.toOacGenerateParams(generateParams)
	// failed to encrypt, nextElasticKeyName(), Status: 400, Message: 400 Bad Request, Body: error in openapi3filter.RequestError: request body has an error: value is required but missing

	openapiGenerateResponse, err := openapiClient.PostElastickeyElasticKeyIDGenerateWithBodyWithResponse(ctx, *elasticKeyID, &elastickeyElasticKeyIDGenerateParams, "text/plain", nil)
	require.NoError(t, err)

	encrypted, err := oamOacMapperInstance.toPlainGenerateResponse(openapiGenerateResponse)
	require.NoError(t, err)

	return encrypted
}

// RequireEncryptRequest creates an EncryptRequest for testing.
func RequireEncryptRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.EncryptRequest {
	t.Helper()

	return oamOacMapperInstance.toOamEncryptRequest(cleartext)
}

// RequireEncryptResponse encrypts data and returns the ciphertext.
func RequireEncryptResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyID *cryptoutilOpenapiModel.ElasticKeyID, encryptParams *cryptoutilOpenapiModel.EncryptParams, encryptRequest *cryptoutilOpenapiModel.EncryptRequest) *string {
	t.Helper()

	elastickeyElasticKeyIDEncryptParams := oamOacMapperInstance.toOacEncryptParams(encryptParams)
	openapiEncryptResponse, err := openapiClient.PostElastickeyElasticKeyIDEncryptWithTextBodyWithResponse(ctx, *elasticKeyID, &elastickeyElasticKeyIDEncryptParams, *encryptRequest)
	require.NoError(t, err)

	encrypted, err := oamOacMapperInstance.toPlainEncryptResponse(openapiEncryptResponse)
	require.NoError(t, err)

	return encrypted
}

// RequireDecryptRequest creates a DecryptRequest for testing.
func RequireDecryptRequest(t *testing.T, ciphertext *string) *cryptoutilOpenapiModel.DecryptRequest {
	t.Helper()

	return oamOacMapperInstance.toOamDecryptRequest(ciphertext)
}

// RequireDecryptResponse decrypts data and returns the plaintext.
func RequireDecryptResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyID *cryptoutilOpenapiModel.ElasticKeyID, decryptRequest *cryptoutilOpenapiModel.DecryptRequest) *string {
	t.Helper()

	openapiDecryptResponse, err := openapiClient.PostElastickeyElasticKeyIDDecryptWithTextBodyWithResponse(ctx, *elasticKeyID, *decryptRequest)
	require.NoError(t, err)

	decrypted, err := oamOacMapperInstance.toPlainDecryptResponse(openapiDecryptResponse)
	require.NoError(t, err)

	return decrypted
}

// RequireSignRequest creates a SignRequest for testing.
func RequireSignRequest(t *testing.T, cleartext *string) *cryptoutilOpenapiModel.SignRequest {
	t.Helper()

	return oamOacMapperInstance.toOamSignRequest(cleartext)
}

// RequireSignResponse signs data and returns the signature.
func RequireSignResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyID *cryptoutilOpenapiModel.ElasticKeyID, signParams *cryptoutilOpenapiModel.SignParams, signRequest *cryptoutilOpenapiModel.SignRequest) *string {
	t.Helper()

	elastickeyElasticKeyIDSignParams := oamOacMapperInstance.toOacSignParams(signParams)
	openapiSignResponse, err := openapiClient.PostElastickeyElasticKeyIDSignWithTextBodyWithResponse(ctx, *elasticKeyID, &elastickeyElasticKeyIDSignParams, *signRequest)
	require.NoError(t, err)

	signed, err := oamOacMapperInstance.toPlainSignResponse(openapiSignResponse)
	require.NoError(t, err)

	return signed
}

// RequireVerifyRequest creates a VerifyRequest for testing.
func RequireVerifyRequest(t *testing.T, signedtext *string) *cryptoutilOpenapiModel.VerifyRequest {
	t.Helper()

	return oamOacMapperInstance.toOamVerifyRequest(signedtext)
}

// RequireVerifyResponse verifies a signature and returns the result.
func RequireVerifyResponse(ctx context.Context, t *testing.T, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, elasticKeyID *cryptoutilOpenapiModel.ElasticKeyID, verifyRequest *cryptoutilOpenapiModel.VerifyRequest) *string {
	t.Helper()

	openapiVerifyResponse, err := openapiClient.PostElastickeyElasticKeyIDVerifyWithTextBodyWithResponse(ctx, *elasticKeyID, *verifyRequest)
	require.NoError(t, err)

	verified, err := oamOacMapperInstance.toPlainVerifyResponse(openapiVerifyResponse)
	require.NoError(t, err)

	return verified
}

// ValidateCreateElasticKeyVsElasticKey validates an ElasticKey against its create request.
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
