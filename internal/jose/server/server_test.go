// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

var (
	testSettings   *cryptoutilConfig.Settings
	testServer     *Server
	testBaseURL    string
	testHTTPClient *http.Client
)

func TestMain(m *testing.M) {
	// Create test settings with dynamic port allocation.
	testSettings = cryptoutilConfig.NewForJOSEServer(
		cryptoutilMagic.IPv4Loopback,
		0, // Dynamic port allocation.
		true,
	)

	// Create server.
	var err error

	testServer, err = New(testSettings)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Start server without blocking.
	if err := testServer.StartNonBlocking(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for server to be ready.
	time.Sleep(cryptoutilMagic.ServerStartupWait)

	// Get the actual port from the listener.
	testBaseURL = fmt.Sprintf("http://%s:%d", cryptoutilMagic.IPv4Loopback, testServer.ActualPort())

	// Create HTTP client for tests.
	testHTTPClient = &http.Client{}

	// Run tests.
	exitCode := m.Run()

	// Shutdown server.
	if shutdownErr := testServer.Shutdown(); shutdownErr != nil {
		fmt.Printf("Server shutdown error: %v\n", shutdownErr)
	}

	os.Exit(exitCode)
}

// doGet performs a GET request with context.
func doGet(t *testing.T, url string) *http.Response {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)

	return resp
}

// doPost performs a POST request with context.
func doPost(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)

	return resp
}

// doDelete performs a DELETE request with context.
func doDelete(t *testing.T, url string) *http.Response {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	require.NoError(t, err)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)

	return resp
}

// closeBody closes the response body and checks for errors.
func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()

	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Warning: failed to close response body: %v", err)
		}
	}
}

func TestHealthEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{"livez", "/livez", "OK"},
		{"readyz", "/readyz", "OK"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := doGet(t, testBaseURL+tc.endpoint)
			defer closeBody(t, resp)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, tc.want, string(body))
		})
	}
}

func TestHealthJSON(t *testing.T) {
	t.Parallel()

	resp := doGet(t, testBaseURL+"/health")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "healthy", result["status"])
	require.Contains(t, result, "time")
}

func TestJWKGenerateAndRetrieve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		use       string
		wantKty   string
	}{
		{"RSA4096", "RSA/4096", "sig", "RSA"},
		{"RSA2048", "RSA/2048", "sig", "RSA"},
		{"ECP256", "EC/P256", "sig", "EC"},
		{"ECP384", "EC/P384", "sig", "EC"},
		{"ECP521", "EC/P521", "sig", "EC"},
		{"OKPEd25519", "OKP/Ed25519", "sig", "OKP"},
		{"Oct256", "oct/256", "enc", "oct"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate key.
			reqBody, err := json.Marshal(JWKGenerateRequest{
				Algorithm: tc.algorithm,
				Use:       tc.use,
			})
			require.NoError(t, err)

			resp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", reqBody)
			defer closeBody(t, resp)

			require.Equal(t, http.StatusCreated, resp.StatusCode)

			var genResp JWKGenerateResponse

			err = json.NewDecoder(resp.Body).Decode(&genResp)
			require.NoError(t, err)
			require.NotEmpty(t, genResp.KID)
			require.Equal(t, tc.algorithm, genResp.Algorithm)
			require.Equal(t, tc.use, genResp.Use)
			require.Equal(t, tc.wantKty, genResp.KeyType)

			// Get key.
			getResp := doGet(t, testBaseURL+"/jose/v1/jwk/"+genResp.KID)
			defer closeBody(t, getResp)

			require.Equal(t, http.StatusOK, getResp.StatusCode)

			var gotKey JWKGenerateResponse

			err = json.NewDecoder(getResp.Body).Decode(&gotKey)
			require.NoError(t, err)
			require.Equal(t, genResp.KID, gotKey.KID)
		})
	}
}

func TestJWKGenerateInvalidAlgorithm(t *testing.T) {
	t.Parallel()

	reqBody, err := json.Marshal(JWKGenerateRequest{
		Algorithm: "InvalidAlg",
		Use:       "sig",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWKGetNotFound(t *testing.T) {
	t.Parallel()

	resp := doGet(t, testBaseURL+"/jose/v1/jwk/nonexistent-kid")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWKDeleteNotFound(t *testing.T) {
	t.Parallel()

	resp := doDelete(t, testBaseURL+"/jose/v1/jwk/nonexistent-kid/delete")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWKList(t *testing.T) {
	t.Parallel()

	resp := doGet(t, testBaseURL+"/jose/v1/jwk")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "keys")
	require.Contains(t, result, "count")
}

func TestJWSSignAndVerify(t *testing.T) {
	t.Parallel()

	// First generate a key.
	genReqBody, err := json.Marshal(JWKGenerateRequest{
		Algorithm: "EC/P256",
		Use:       "sig",
	})
	require.NoError(t, err)

	genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
	defer closeBody(t, genResp)

	require.Equal(t, http.StatusCreated, genResp.StatusCode)

	var key JWKGenerateResponse

	err = json.NewDecoder(genResp.Body).Decode(&key)
	require.NoError(t, err)

	// Sign a payload.
	signReqBody, err := json.Marshal(JWSSignRequest{
		KID:     key.KID,
		Payload: "Hello, World!",
	})
	require.NoError(t, err)

	signResp := doPost(t, testBaseURL+"/jose/v1/jws/sign", signReqBody)
	defer closeBody(t, signResp)

	require.Equal(t, http.StatusOK, signResp.StatusCode)

	var signResult JWSSignResponse

	err = json.NewDecoder(signResp.Body).Decode(&signResult)
	require.NoError(t, err)
	require.NotEmpty(t, signResult.JWS)

	// Verify the signature.
	verifyReqBody, err := json.Marshal(JWSVerifyRequest{
		JWS: signResult.JWS,
		KID: key.KID,
	})
	require.NoError(t, err)

	verifyResp := doPost(t, testBaseURL+"/jose/v1/jws/verify", verifyReqBody)
	defer closeBody(t, verifyResp)

	require.Equal(t, http.StatusOK, verifyResp.StatusCode)

	var verifyResult JWSVerifyResponse

	err = json.NewDecoder(verifyResp.Body).Decode(&verifyResult)
	require.NoError(t, err)
	require.True(t, verifyResult.Valid)
	require.Equal(t, "Hello, World!", verifyResult.Payload)
}

func TestJWEEncryptAndDecrypt(t *testing.T) {
	// Skip for now - JWE requires keys with enc header set.
	// The GenerateJWK function creates signing keys, not encryption keys.
	// TODO: Add GenerateJWEJWK endpoint or modify GenerateJWK to support use parameter.
	t.Skip("JWE requires encryption-specific key generation")
}

func TestJWTCreateAndVerify(t *testing.T) {
	t.Parallel()

	// First generate a key.
	genReqBody, err := json.Marshal(JWKGenerateRequest{
		Algorithm: "EC/P384",
		Use:       "sig",
	})
	require.NoError(t, err)

	genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
	defer closeBody(t, genResp)

	require.Equal(t, http.StatusCreated, genResp.StatusCode)

	var key JWKGenerateResponse

	err = json.NewDecoder(genResp.Body).Decode(&key)
	require.NoError(t, err)

	// Create a JWT.
	createReqBody, err := json.Marshal(JWTCreateRequest{
		KID: key.KID,
		Claims: map[string]interface{}{
			"sub":  "user123",
			"name": "Test User",
			"iat":  time.Now().Unix(),
		},
	})
	require.NoError(t, err)

	createResp := doPost(t, testBaseURL+"/jose/v1/jwt/sign", createReqBody)
	defer closeBody(t, createResp)

	require.Equal(t, http.StatusOK, createResp.StatusCode)

	var createResult JWTCreateResponse

	err = json.NewDecoder(createResp.Body).Decode(&createResult)
	require.NoError(t, err)
	require.NotEmpty(t, createResult.JWT)

	// Verify the JWT.
	verifyReqBody, err := json.Marshal(JWTVerifyRequest{
		JWT: createResult.JWT,
		KID: key.KID,
	})
	require.NoError(t, err)

	verifyResp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", verifyReqBody)
	defer closeBody(t, verifyResp)

	require.Equal(t, http.StatusOK, verifyResp.StatusCode)

	var verifyResult JWTVerifyResponse

	err = json.NewDecoder(verifyResp.Body).Decode(&verifyResult)
	require.NoError(t, err)
	require.True(t, verifyResult.Valid)
	require.Equal(t, "user123", verifyResult.Claims["sub"])
	require.Equal(t, "Test User", verifyResult.Claims["name"])
}

func TestWellKnownJWKS(t *testing.T) {
	t.Parallel()

	resp := doGet(t, testBaseURL+"/.well-known/jwks.json")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var result map[string]interface{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "keys")
}
