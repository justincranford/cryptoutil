// Copyright (c) 2025 Justin Cranford

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	cryptoutilJoseConfig "cryptoutil/internal/jose/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Test setup for JoseServer (new path tests).
var (
	joseTestServer     *JoseServer
	joseTestBaseURL    string
	joseTestHTTPClient *http.Client
	joseSetupOnce      sync.Once
	joseSetupErr       error
)

// setupJoseTestServer initializes the JoseServer once for new path tests.
func setupJoseTestServer() error {
	joseSetupOnce.Do(func() {
		ctx := context.Background()

		// Create test settings.
		cfg := cryptoutilJoseConfig.NewTestSettings()

		// Create JoseServer.
		joseTestServer, joseSetupErr = NewFromConfig(ctx, cfg)
		if joseSetupErr != nil {
			joseSetupErr = fmt.Errorf("failed to create jose server: %w", joseSetupErr)

			return
		}

		// Start server in background.
		go func() {
			if err := joseTestServer.Start(ctx); err != nil {
				// Log but don't fail - server might be stopped by tests.
				fmt.Printf("JoseServer stopped: %v\n", err)
			}
		}()

		// Wait for server to be ready by polling for a valid port.
		var actualPort int
		for i := 0; i < 50; i++ { // Up to 5 seconds (50 * 100ms)
			actualPort = joseTestServer.PublicPort()
			if actualPort > 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if actualPort == 0 {
			joseSetupErr = fmt.Errorf("server failed to start: port is still 0 after timeout")

			return
		}

		// Get the actual port.
		joseTestBaseURL = fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, actualPort)

		// Create HTTP client.
		joseTestHTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Test environment only
				},
			},
		}
	})

	return joseSetupErr
}

// doJoseGet performs a GET request for jose tests.
func doJoseGet(t *testing.T, url string) *http.Response {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := joseTestHTTPClient.Do(req)
	require.NoError(t, err)

	return resp
}

// doJosePost performs a POST request for jose tests.
func doJosePost(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := joseTestHTTPClient.Do(req)
	require.NoError(t, err)

	return resp
}

// closeJoseBody closes the response body.
func closeJoseBody(t *testing.T, resp *http.Response) {
	t.Helper()

	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Warning: failed to close response body: %v", err)
		}
	}
}

// TestNewPaths_ServiceAPIv1 tests the new /service/api/v1/jose/** paths.
func TestNewPaths_ServiceAPIv1(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupJoseTestServer())

	basePath := "/service/api/v1/jose"

	t.Run("JWK_Generate", func(t *testing.T) {
		t.Parallel()

		reqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "EC/P256",
			Use:       "sig",
		})
		require.NoError(t, err)

		resp := doJosePost(t, joseTestBaseURL+basePath+"/jwk/generate", reqBody)
		defer closeJoseBody(t, resp)

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var genResp JWKGenerateResponse

		err = json.NewDecoder(resp.Body).Decode(&genResp)
		require.NoError(t, err)
		require.NotEmpty(t, genResp.KID)
		require.Equal(t, "EC/P256", genResp.Algorithm)
	})

	t.Run("JWK_List", func(t *testing.T) {
		t.Parallel()

		resp := doJoseGet(t, joseTestBaseURL+basePath+"/jwk")
		defer closeJoseBody(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any

		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "keys")
		require.Contains(t, result, "count")
	})

	t.Run("JWKS", func(t *testing.T) {
		t.Parallel()

		resp := doJoseGet(t, joseTestBaseURL+basePath+"/jwks")
		defer closeJoseBody(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any

		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "keys")
	})

	t.Run("JWS_Sign_And_Verify", func(t *testing.T) {
		t.Parallel()

		// Generate key.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "EC/P256",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doJosePost(t, joseTestBaseURL+basePath+"/jwk/generate", genReqBody)
		defer closeJoseBody(t, genResp)

		require.Equal(t, http.StatusCreated, genResp.StatusCode)

		var key JWKGenerateResponse

		err = json.NewDecoder(genResp.Body).Decode(&key)
		require.NoError(t, err)

		// Sign.
		signReqBody, err := json.Marshal(JWSSignRequest{
			KID:     key.KID,
			Payload: "Test message for service API",
		})
		require.NoError(t, err)

		signResp := doJosePost(t, joseTestBaseURL+basePath+"/jws/sign", signReqBody)
		defer closeJoseBody(t, signResp)

		require.Equal(t, http.StatusOK, signResp.StatusCode)

		var signResult JWSSignResponse

		err = json.NewDecoder(signResp.Body).Decode(&signResult)
		require.NoError(t, err)
		require.NotEmpty(t, signResult.JWS)

		// Verify.
		verifyReqBody, err := json.Marshal(JWSVerifyRequest{
			JWS: signResult.JWS,
			KID: key.KID,
		})
		require.NoError(t, err)

		verifyResp := doJosePost(t, joseTestBaseURL+basePath+"/jws/verify", verifyReqBody)
		defer closeJoseBody(t, verifyResp)

		require.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var verifyResult JWSVerifyResponse

		err = json.NewDecoder(verifyResp.Body).Decode(&verifyResult)
		require.NoError(t, err)
		require.True(t, verifyResult.Valid)
		require.Equal(t, "Test message for service API", verifyResult.Payload)
	})

	t.Run("JWE_Encrypt_And_Decrypt", func(t *testing.T) {
		t.Parallel()

		// Generate encryption key.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "oct/256",
			Use:       "enc",
		})
		require.NoError(t, err)

		genResp := doJosePost(t, joseTestBaseURL+basePath+"/jwk/generate", genReqBody)
		defer closeJoseBody(t, genResp)

		require.Equal(t, http.StatusCreated, genResp.StatusCode)

		var key JWKGenerateResponse

		err = json.NewDecoder(genResp.Body).Decode(&key)
		require.NoError(t, err)

		// Encrypt.
		encryptReqBody, err := json.Marshal(JWEEncryptRequest{
			KID:       key.KID,
			Plaintext: "Secret message for service API",
		})
		require.NoError(t, err)

		encryptResp := doJosePost(t, joseTestBaseURL+basePath+"/jwe/encrypt", encryptReqBody)
		defer closeJoseBody(t, encryptResp)

		require.Equal(t, http.StatusOK, encryptResp.StatusCode)

		var encryptResult JWEEncryptResponse

		err = json.NewDecoder(encryptResp.Body).Decode(&encryptResult)
		require.NoError(t, err)
		require.NotEmpty(t, encryptResult.JWE)

		// Decrypt.
		decryptReqBody, err := json.Marshal(JWEDecryptRequest{
			KID: key.KID,
			JWE: encryptResult.JWE,
		})
		require.NoError(t, err)

		decryptResp := doJosePost(t, joseTestBaseURL+basePath+"/jwe/decrypt", decryptReqBody)
		defer closeJoseBody(t, decryptResp)

		require.Equal(t, http.StatusOK, decryptResp.StatusCode)

		var decryptResult JWEDecryptResponse

		err = json.NewDecoder(decryptResp.Body).Decode(&decryptResult)
		require.NoError(t, err)
		require.Equal(t, "Secret message for service API", decryptResult.Plaintext)
	})

	t.Run("JWT_Sign_And_Verify", func(t *testing.T) {
		t.Parallel()

		// Generate key.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "EC/P256",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doJosePost(t, joseTestBaseURL+basePath+"/jwk/generate", genReqBody)
		defer closeJoseBody(t, genResp)

		require.Equal(t, http.StatusCreated, genResp.StatusCode)

		var key JWKGenerateResponse

		err = json.NewDecoder(genResp.Body).Decode(&key)
		require.NoError(t, err)

		// Sign JWT.
		signReqBody, err := json.Marshal(map[string]any{
			"kid": key.KID,
			"claims": map[string]any{
				"sub":  "test-subject",
				"name": "Service API Test",
			},
		})
		require.NoError(t, err)

		signResp := doJosePost(t, joseTestBaseURL+basePath+"/jwt/sign", signReqBody)
		defer closeJoseBody(t, signResp)

		require.Equal(t, http.StatusOK, signResp.StatusCode)

		var signResult map[string]any

		err = json.NewDecoder(signResp.Body).Decode(&signResult)
		require.NoError(t, err)
		require.Contains(t, signResult, "jwt")
		require.NotEmpty(t, signResult["jwt"])

		// Verify JWT.
		verifyReqBody, err := json.Marshal(map[string]any{
			"jwt": signResult["jwt"],
			"kid": key.KID,
		})
		require.NoError(t, err)

		verifyResp := doJosePost(t, joseTestBaseURL+basePath+"/jwt/verify", verifyReqBody)
		defer closeJoseBody(t, verifyResp)

		require.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var verifyResult map[string]any

		err = json.NewDecoder(verifyResp.Body).Decode(&verifyResult)
		require.NoError(t, err)

		valid, ok := verifyResult["valid"].(bool)
		require.True(t, ok, "expected valid field to be bool")
		require.True(t, valid)
	})
}

// TestNewPaths_BrowserAPIv1 tests the new /browser/api/v1/jose/** paths.
func TestNewPaths_BrowserAPIv1(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupJoseTestServer())

	basePath := "/browser/api/v1/jose"

	t.Run("JWK_Generate", func(t *testing.T) {
		t.Parallel()

		reqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "RSA/2048",
			Use:       "sig",
		})
		require.NoError(t, err)

		resp := doJosePost(t, joseTestBaseURL+basePath+"/jwk/generate", reqBody)
		defer closeJoseBody(t, resp)

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var genResp JWKGenerateResponse

		err = json.NewDecoder(resp.Body).Decode(&genResp)
		require.NoError(t, err)
		require.NotEmpty(t, genResp.KID)
		require.Equal(t, "RSA/2048", genResp.Algorithm)
	})

	t.Run("JWK_List", func(t *testing.T) {
		t.Parallel()

		resp := doJoseGet(t, joseTestBaseURL+basePath+"/jwk")
		defer closeJoseBody(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any

		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "keys")
		require.Contains(t, result, "count")
	})

	t.Run("JWS_SignAndVerify", func(t *testing.T) {
		t.Parallel()

		// Generate key.
		genReqBody, err := json.Marshal(JWKGenerateRequest{
			Algorithm: "RSA/2048",
			Use:       "sig",
		})
		require.NoError(t, err)

		genResp := doJosePost(t, joseTestBaseURL+basePath+"/jwk/generate", genReqBody)
		defer closeJoseBody(t, genResp)

		require.Equal(t, http.StatusCreated, genResp.StatusCode)

		var key JWKGenerateResponse

		err = json.NewDecoder(genResp.Body).Decode(&key)
		require.NoError(t, err)

		// Sign.
		signReqBody, err := json.Marshal(JWSSignRequest{
			KID:     key.KID,
			Payload: "Test message for browser API",
		})
		require.NoError(t, err)

		signResp := doJosePost(t, joseTestBaseURL+basePath+"/jws/sign", signReqBody)
		defer closeJoseBody(t, signResp)

		require.Equal(t, http.StatusOK, signResp.StatusCode)

		var signResult JWSSignResponse

		err = json.NewDecoder(signResp.Body).Decode(&signResult)
		require.NoError(t, err)
		require.NotEmpty(t, signResult.JWS)

		// Verify.
		verifyReqBody, err := json.Marshal(JWSVerifyRequest{
			JWS: signResult.JWS,
			KID: key.KID,
		})
		require.NoError(t, err)

		verifyResp := doJosePost(t, joseTestBaseURL+basePath+"/jws/verify", verifyReqBody)
		defer closeJoseBody(t, verifyResp)

		require.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var verifyResult JWSVerifyResponse

		err = json.NewDecoder(verifyResp.Body).Decode(&verifyResult)
		require.NoError(t, err)
		require.True(t, verifyResult.Valid)
		require.Equal(t, "Test message for browser API", verifyResult.Payload)
	})
}

// TestNewPaths_WellKnownJWKS tests the /.well-known/jwks.json endpoint.
func TestNewPaths_WellKnownJWKS(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupJoseTestServer())

	resp := doJoseGet(t, joseTestBaseURL+"/.well-known/jwks.json")
	defer closeJoseBody(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Contains(t, result, "keys")
}

// TestNewPaths_RateLimitingApplied tests that rate limiting is applied to new paths.
func TestNewPaths_RateLimitingApplied(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupJoseTestServer())

	// Make requests rapidly to trigger rate limit.
	// Default is 100 requests per second per IP.
	// We'll make 150+ requests rapidly to exceed limit.
	var rateLimited bool

	for range 150 {
		ctx := context.Background()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, joseTestBaseURL+"/service/api/v1/jose/jwk", nil)
		require.NoError(t, err)

		resp, err := joseTestHTTPClient.Do(req)
		require.NoError(t, err)

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true

			// Verify response body.
			body, readErr := io.ReadAll(resp.Body)
			require.NoError(t, readErr)
			require.Contains(t, string(body), "Rate limit exceeded")
		}

		require.NoError(t, resp.Body.Close())

		if rateLimited {
			break
		}
	}

	require.True(t, rateLimited, "Rate limiting should be triggered after 100+ rapid requests")
}

// TestNewPaths_AdminEndpoints tests the /browser/api/v1/admin/** endpoints.
func TestNewPaths_AdminEndpoints(t *testing.T) {
	t.Parallel()

	require.NoError(t, setupJoseTestServer())

	t.Run("GetAuditConfig", func(t *testing.T) {
		t.Parallel()

		resp := doJoseGet(t, joseTestBaseURL+"/browser/api/v1/admin/audit-config")
		defer closeJoseBody(t, resp)

		// Should return 200 OK with audit config.
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any

		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "configs")
	})

	t.Run("GetAuditConfigByOperation", func(t *testing.T) {
		t.Parallel()

		resp := doJoseGet(t, joseTestBaseURL+"/browser/api/v1/admin/audit-config/sign")
		defer closeJoseBody(t, resp)

		// Should return 200 OK or 404 if operation not configured.
		require.True(t,
			resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
			"Expected 200 OK or 404 Not Found, got %d", resp.StatusCode)
	})

	t.Run("SetAuditConfig", func(t *testing.T) {
		t.Parallel()

		// Test setting audit config.
		ctx := context.Background()

		reqBody, err := json.Marshal(map[string]any{
			"operation": "sign",
			"enabled":   true,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, joseTestBaseURL+"/browser/api/v1/admin/audit-config", bytes.NewBuffer(reqBody))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := joseTestHTTPClient.Do(req)
		require.NoError(t, err)

		defer func() { require.NoError(t, resp.Body.Close()) }()

		// Should return 200 OK on successful update.
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
