// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilJoseServerMiddleware "cryptoutil/internal/jose/server/middleware"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// createTestTLSConfig creates a TLSGeneratedSettings for testing.
func createTestTLSConfig() *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings {
	tlsCfg, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost", "jose-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test TLS config: %v", err))
	}

	return tlsCfg
}

var (
	testSettings   *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	testServer     *Server
	testBaseURL    string
	testHTTPClient *http.Client
	setupOnce      sync.Once
	setupErr       error
)

// setupTestServer initializes the test server once for all tests.
// Uses sync.Once to ensure safe concurrent access from parallel tests.
// CRITICAL: This pattern replaces TestMain to avoid os.Exit() deadlock with t.Parallel().
func setupTestServer() error {
	setupOnce.Do(func() {
		// Create test settings with dynamic port allocation using test helper.
		// NewTestConfig bypasses pflag global FlagSet to allow multiple test instances.
		testSettings = cryptoutilAppsTemplateServiceConfig.NewTestConfig(
			cryptoutilSharedMagic.IPv4Loopback,
			0, // Dynamic port allocation.
			true,
		)

		// Create server.
		testServer, setupErr = New(testSettings)
		if setupErr != nil {
			setupErr = fmt.Errorf("failed to create server: %w", setupErr)

			return
		}

		// Start server without blocking.
		if setupErr = testServer.StartNonBlocking(); setupErr != nil {
			setupErr = fmt.Errorf("failed to start server: %w", setupErr)

			return
		}

		// Wait for server to be ready.
		time.Sleep(cryptoutilSharedMagic.ServerStartupWait)

		// Get the actual port from the listener.
		testBaseURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, testServer.ActualPort())

		// Create HTTP client with TLS config for self-signed certificates.
		testHTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Test environment only
				},
			},
		}
	})

	return setupErr
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

	require.NoError(t, setupTestServer())

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

	require.NoError(t, setupTestServer())

	resp := doGet(t, testBaseURL+"/health")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "healthy", result["status"])
	require.Contains(t, result, "time")
}

func TestJWKGenerateAndRetrieve(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	tests := []struct {
		name      string
		algorithm string
		use       string
		wantKty   string
	}{
		{"RSA4096", "RSA/4096", "sig", "RSA"},
		{"RSA3072", "RSA/3072", "sig", "RSA"},
		{"RSA2048", "RSA/2048", "sig", "RSA"},
		{"ECP256", "EC/P256", "sig", "EC"},
		{"ECP384", "EC/P384", "sig", "EC"},
		{"ECP521", "EC/P521", "sig", "EC"},
		{"OKPEd25519", "OKP/Ed25519", "sig", "OKP"},
		{"Oct512", "oct/512", "sig", "oct"},
		{"Oct384", "oct/384", "sig", "oct"},
		{"Oct256", "oct/256", "enc", "oct"},
		{"Oct192", "oct/192", "enc", "oct"},
		{"Oct128", "oct/128", "enc", "oct"},
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

	require.NoError(t, setupTestServer())

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

	require.NoError(t, setupTestServer())

	resp := doGet(t, testBaseURL+"/jose/v1/jwk/nonexistent-kid")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWKDeleteNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	resp := doDelete(t, testBaseURL+"/jose/v1/jwk/nonexistent-kid/delete")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWKList(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	resp := doGet(t, testBaseURL+"/jose/v1/jwk")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "keys")
	require.Contains(t, result, "count")
}

func TestJWSSignAndVerify(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

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
	t.Parallel()

	require.NoError(t, setupTestServer())

	tests := []struct {
		name       string
		algorithm  string
		use        string
		plaintext  string
		shouldPass bool
	}{
		{"Oct512", "oct/512", "enc", "AES-256-GCM test", true},
		{"Oct384", "oct/384", "enc", "AES-192-GCM test", true},
		{"Oct256", "oct/256", "enc", "Hello, World!", true},
		{"Oct192", "oct/192", "enc", "Test message", true},
		{"Oct128", "oct/128", "enc", "Short", true},
		{"RSA4096", "RSA/4096", "enc", "RSA-4096 encryption", true},
		{"RSA3072", "RSA/3072", "enc", "RSA-3072 encryption", true},
		{"RSA2048", "RSA/2048", "enc", "RSA-2048 encryption", true},
		{"ECP521", "EC/P521", "enc", "ECDH-P521 test", true},
		{"ECP384", "EC/P384", "enc", "ECDH-P384 test", true},
		{"ECP256", "EC/P256", "enc", "ECDH-P256 test", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate encryption key.
			genReqBody, err := json.Marshal(JWKGenerateRequest{
				Algorithm: tc.algorithm,
				Use:       tc.use,
			})
			require.NoError(t, err)

			genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
			defer closeBody(t, genResp)

			require.Equal(t, http.StatusCreated, genResp.StatusCode)

			var key JWKGenerateResponse

			err = json.NewDecoder(genResp.Body).Decode(&key)
			require.NoError(t, err)
			require.NotEmpty(t, key.KID)
			require.Equal(t, tc.algorithm, key.Algorithm)
			require.Equal(t, tc.use, key.Use)

			// Encrypt plaintext.
			encryptReqBody, err := json.Marshal(JWEEncryptRequest{
				KID:       key.KID,
				Plaintext: tc.plaintext,
			})
			require.NoError(t, err)

			encryptResp := doPost(t, testBaseURL+"/jose/v1/jwe/encrypt", encryptReqBody)
			defer closeBody(t, encryptResp)

			require.Equal(t, http.StatusOK, encryptResp.StatusCode)

			var encryptResult JWEEncryptResponse

			err = json.NewDecoder(encryptResp.Body).Decode(&encryptResult)
			require.NoError(t, err)
			require.NotEmpty(t, encryptResult.JWE)

			// Decrypt the JWE.
			decryptReqBody, err := json.Marshal(JWEDecryptRequest{
				JWE: encryptResult.JWE,
				KID: key.KID,
			})
			require.NoError(t, err)

			decryptResp := doPost(t, testBaseURL+"/jose/v1/jwe/decrypt", decryptReqBody)
			defer closeBody(t, decryptResp)

			require.Equal(t, http.StatusOK, decryptResp.StatusCode)

			var decryptResult JWEDecryptResponse

			err = json.NewDecoder(decryptResp.Body).Decode(&decryptResult)
			require.NoError(t, err)
			require.Equal(t, tc.plaintext, decryptResult.Plaintext)
			require.Equal(t, key.KID, decryptResult.KID)
		})
	}
}

func TestJWTCreateAndVerify(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

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
		Claims: map[string]any{
			"sub":  "user123",
			"name": "Test User",
			"iat":  time.Now().UTC().Unix(),
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

	require.NoError(t, setupTestServer())

	resp := doGet(t, testBaseURL+"/.well-known/jwks.json")
	defer closeBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var result map[string]any

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "keys")
}

// Additional edge case tests for better coverage.

func TestJWSSignMissingKID(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWSSignRequest{
		KID:     "",
		Payload: "test",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jws/sign", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWSSignMissingPayload(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWSSignRequest{
		KID:     "some-kid",
		Payload: "",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jws/sign", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWSSignKeyNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWSSignRequest{
		KID:     "nonexistent-kid",
		Payload: "test",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jws/sign", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWSVerifyMissingJWS(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWSVerifyRequest{
		JWS: "",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jws/verify", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWSVerifyKeyNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWSVerifyRequest{
		JWS: "eyJhbGciOiJFUzI1NiJ9.dGVzdA.signature",
		KID: "nonexistent-kid",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jws/verify", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWEEncryptMissingKID(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWEEncryptRequest{
		KID:       "",
		Plaintext: "test",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwe/encrypt", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWEEncryptMissingPlaintext(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWEEncryptRequest{
		KID:       "some-kid",
		Plaintext: "",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwe/encrypt", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWEEncryptKeyNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWEEncryptRequest{
		KID:       "nonexistent-kid",
		Plaintext: "test",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwe/encrypt", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWEDecryptMissingJWE(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWEDecryptRequest{
		JWE: "",
		KID: "some-kid",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwe/decrypt", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWEDecryptMissingKID(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWEDecryptRequest{
		JWE: "some-jwe",
		KID: "",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwe/decrypt", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWEDecryptKeyNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWEDecryptRequest{
		JWE: "eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0...",
		KID: "nonexistent-kid",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwe/decrypt", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWTCreateMissingKID(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWTCreateRequest{
		KID: "",
		Claims: map[string]any{
			"sub": "user",
		},
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwt/sign", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWTCreateMissingClaims(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWTCreateRequest{
		KID:    "some-kid",
		Claims: nil,
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwt/sign", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWTCreateKeyNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWTCreateRequest{
		KID: "nonexistent-kid",
		Claims: map[string]any{
			"sub": "user",
		},
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwt/sign", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWTVerifyMissingJWT(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWTVerifyRequest{
		JWT: "",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJWTVerifyKeyNotFound(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	reqBody, err := json.Marshal(JWTVerifyRequest{
		JWT: "eyJhbGciOiJFUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.sig",
		KID: "nonexistent-kid",
	})
	require.NoError(t, err)

	resp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", reqBody)
	defer closeBody(t, resp)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestJWKGetMissingKID(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	// The path /jose/v1/jwk without a KID returns the list endpoint.
	resp := doGet(t, testBaseURL+"/jose/v1/jwk")
	defer closeBody(t, resp)

	// Should return 200 and list keys (as this matches list endpoint).
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestJWKDeleteSuccess(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	// Generate a key first.
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

	// Delete the key.
	deleteResp := doDelete(t, testBaseURL+"/jose/v1/jwk/"+key.KID+"/delete")
	defer closeBody(t, deleteResp)

	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// Verify key is deleted.
	getResp := doGet(t, testBaseURL+"/jose/v1/jwk/"+key.KID)
	defer closeBody(t, getResp)

	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestInvalidJSONBody(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	tests := []struct {
		name     string
		endpoint string
	}{
		{"generate", "/jose/v1/jwk/generate"},
		{"sign", "/jose/v1/jws/sign"},
		{"verify", "/jose/v1/jws/verify"},
		{"encrypt", "/jose/v1/jwe/encrypt"},
		{"decrypt", "/jose/v1/jwe/decrypt"},
		{"jwtCreate", "/jose/v1/jwt/sign"},
		{"jwtVerify", "/jose/v1/jwt/verify"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := doPost(t, testBaseURL+tc.endpoint, []byte("invalid json"))
			defer closeBody(t, resp)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

// TestJWSVerifyErrorPaths tests JWS verification error scenarios.
func TestJWSVerifyErrorPaths(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	t.Run("MissingJWS", func(t *testing.T) {
		t.Parallel()

		reqBody, err := json.Marshal(JWSVerifyRequest{})
		require.NoError(t, err)

		resp := doPost(t, testBaseURL+"/jose/v1/jws/verify", reqBody)
		defer closeBody(t, resp)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		t.Parallel()

		reqBody, err := json.Marshal(JWSVerifyRequest{
			JWS: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			KID: "nonexistent-kid",
		})
		require.NoError(t, err)

		resp := doPost(t, testBaseURL+"/jose/v1/jws/verify", reqBody)
		defer closeBody(t, resp)

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		t.Parallel()

		// Generate a key.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "RSA/2048",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
		defer closeBody(t, genResp)

		var genResult JWKGenerateResponse

		require.NoError(t, json.NewDecoder(genResp.Body).Decode(&genResult))

		// Try to verify invalid JWS with specific key.
		reqBody, err := json.Marshal(JWSVerifyRequest{
			JWS: "invalid.jws.signature",
			KID: genResult.KID,
		})
		require.NoError(t, err)

		resp := doPost(t, testBaseURL+"/jose/v1/jws/verify", reqBody)
		defer closeBody(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var verifyResp JWSVerifyResponse

		require.NoError(t, json.NewDecoder(resp.Body).Decode(&verifyResp))
		require.False(t, verifyResp.Valid)
	})

	t.Run("VerifyWithoutKID", func(t *testing.T) {
		t.Parallel()

		// Generate a key and sign.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "EC/P256",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
		defer closeBody(t, genResp)

		var genResult JWKGenerateResponse

		require.NoError(t, json.NewDecoder(genResp.Body).Decode(&genResult))

		// Sign a payload.
		signReqBody, err := json.Marshal(JWSSignRequest{
			KID:     genResult.KID,
			Payload: "Test payload for KID-less verify",
		})
		require.NoError(t, err)

		signResp := doPost(t, testBaseURL+"/jose/v1/jws/sign", signReqBody)
		defer closeBody(t, signResp)

		var signResult JWSSignResponse

		require.NoError(t, json.NewDecoder(signResp.Body).Decode(&signResult))

		// Verify WITHOUT providing KID (should try all keys).
		verifyReqBody, err := json.Marshal(JWSVerifyRequest{
			JWS: signResult.JWS,
		})
		require.NoError(t, err)

		verifyResp := doPost(t, testBaseURL+"/jose/v1/jws/verify", verifyReqBody)
		defer closeBody(t, verifyResp)

		require.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var verifyResult JWSVerifyResponse

		require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&verifyResult))
		require.True(t, verifyResult.Valid)
		require.Equal(t, "Test payload for KID-less verify", verifyResult.Payload)
	})
}

// TestJWTVerifyErrorPaths tests JWT verification error scenarios.
func TestJWTVerifyErrorPaths(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	t.Run("MissingJWT", func(t *testing.T) {
		t.Parallel()

		reqBody, err := json.Marshal(JWTVerifyRequest{})
		require.NoError(t, err)

		resp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", reqBody)
		defer closeBody(t, resp)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		t.Parallel()

		reqBody, err := json.Marshal(JWTVerifyRequest{
			JWT: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			KID: "nonexistent-kid",
		})
		require.NoError(t, err)

		resp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", reqBody)
		defer closeBody(t, resp)

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidJWTFormat", func(t *testing.T) {
		t.Parallel()

		// First create a valid key.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "oct/256",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
		defer closeBody(t, genResp)

		var genRespData JWKGenerateResponse

		err = json.NewDecoder(genResp.Body).Decode(&genRespData)
		require.NoError(t, err)

		// Try to verify an invalid JWT.
		reqBody, err := json.Marshal(JWTVerifyRequest{
			JWT: "invalid.jwt.format",
			KID: genRespData.KID,
		})
		require.NoError(t, err)

		resp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", reqBody)
		defer closeBody(t, resp)

		var verifyResp JWTVerifyResponse

		err = json.NewDecoder(resp.Body).Decode(&verifyResp)
		require.NoError(t, err)
		require.False(t, verifyResp.Valid)
		require.NotEmpty(t, verifyResp.Error)
	})

	t.Run("VerifyWithoutKID", func(t *testing.T) {
		t.Parallel()

		// Generate a key and create JWT.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "EC/P384",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doPost(t, testBaseURL+"/jose/v1/jwk/generate", genReqBody)
		defer closeBody(t, genResp)

		var genResult JWKGenerateResponse

		require.NoError(t, json.NewDecoder(genResp.Body).Decode(&genResult))

		// Create JWT.
		createReqBody, err := json.Marshal(JWTCreateRequest{
			KID:    genResult.KID,
			Claims: map[string]any{"sub": "test-subject", "test": "claim"},
		})
		require.NoError(t, err)

		createResp := doPost(t, testBaseURL+"/jose/v1/jwt/sign", createReqBody)
		defer closeBody(t, createResp)

		require.Equal(t, http.StatusOK, createResp.StatusCode)

		var createResult JWTCreateResponse

		require.NoError(t, json.NewDecoder(createResp.Body).Decode(&createResult))

		// Verify WITHOUT providing KID (should try all keys).
		verifyReqBody, err := json.Marshal(JWTVerifyRequest{
			JWT: createResult.JWT,
		})
		require.NoError(t, err)

		verifyResp := doPost(t, testBaseURL+"/jose/v1/jwt/verify", verifyReqBody)
		defer closeBody(t, verifyResp)

		require.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var verifyResult JWTVerifyResponse

		require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&verifyResult))
		require.True(t, verifyResult.Valid)
		require.NotEmpty(t, verifyResult.Claims)
	})
}

// TestServerLifecycle tests server Start and Shutdown.
func TestServerLifecycle(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(
		cryptoutilSharedMagic.IPv4Loopback,
		0, // Dynamic port.
		true,
	)

	server, err := NewServer(context.Background(), settings, createTestTLSConfig())
	require.NoError(t, err)

	// Test StartNonBlocking.
	require.NoError(t, server.StartNonBlocking())
	require.NotZero(t, server.ActualPort())

	// Test Shutdown.
	require.NoError(t, server.Shutdown())
}

// TestAPIKeyMiddleware tests API key configuration.
func TestAPIKeyMiddleware(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(
		cryptoutilSharedMagic.IPv4Loopback,
		0,
		true,
	)

	server, err := NewServer(context.Background(), settings, createTestTLSConfig())
	require.NoError(t, err)

	// Initially nil middleware.
	require.Nil(t, server.GetAPIKeyMiddleware())

	// Configure API key auth.
	server.ConfigureAPIKeyAuth(&cryptoutilJoseServerMiddleware.APIKeyConfig{
		HeaderName: "X-API-Key",
		ValidKeys: map[string]string{
			"test-key-123": "test-client",
		},
	})

	middleware := server.GetAPIKeyMiddleware()
	require.NotNil(t, middleware)

	// Configure with nil config (should use defaults).
	server.ConfigureAPIKeyAuth(nil)
	require.NotNil(t, server.GetAPIKeyMiddleware())
}

// TestNewServerErrorPaths tests NewServer error scenarios.
func TestNewServerErrorPaths(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	validSettings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(
		cryptoutilSharedMagic.IPv4Loopback,
		0,
		true,
	)

	t.Run("NilContext", func(t *testing.T) {
		t.Parallel()

		_, err := NewServer(nil, validSettings, nil) //nolint:staticcheck // Testing nil context error path.
		require.Error(t, err)
		require.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("NilSettings", func(t *testing.T) {
		t.Parallel()

		_, err := NewServer(context.Background(), nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "settings cannot be nil")
	})
}

// TestStartBlocking tests the blocking Start method.
func TestStartBlocking(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(
		cryptoutilSharedMagic.IPv4Loopback,
		0,
		true,
	)

	server, err := NewServer(context.Background(), settings, createTestTLSConfig())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start should return nil when context is cancelled.
	err = server.Start(ctx)
	require.NoError(t, err) // Context cancellation is normal shutdown.
}

// TestShutdownCoverage tests explicit Shutdown calls for coverage.
func TestShutdownCoverage(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupTestServer())

	t.Run("NormalShutdown", func(t *testing.T) {
		t.Parallel()

		settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(
			cryptoutilSharedMagic.IPv4Loopback,
			0,
			true,
		)

		server, err := NewServer(context.Background(), settings, createTestTLSConfig())
		require.NoError(t, err)

		require.NoError(t, server.StartNonBlocking())

		// Wait for server to be ready.
		time.Sleep(100 * time.Millisecond)

		// Shutdown should succeed.
		require.NoError(t, server.Shutdown())
	})

	t.Run("ShutdownWithoutStart", func(t *testing.T) {
		t.Parallel()

		settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(
			cryptoutilSharedMagic.IPv4Loopback,
			0,
			true,
		)

		server, err := NewServer(context.Background(), settings, createTestTLSConfig())
		require.NoError(t, err)

		// Shutdown without starting should work.
		require.NoError(t, server.Shutdown())
	})
}
