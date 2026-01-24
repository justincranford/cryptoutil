package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/ca/server/config"
)

// TestCAServer_HandleOCSP tests the OCSP endpoint.
func TestCAServer_HandleOCSP(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := config.NewTestConfig("127.0.0.1", 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	time.Sleep(1 * time.Second)

	// Create HTTP client.
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test OCSP endpoint with empty request (should return 400 or 501).
	url := server.PublicBaseURL() + "/service/api/v1/ocsp"
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte{}))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	// Should return 501 NotImplemented or 400 BadRequest.
	require.True(t, resp.StatusCode == http.StatusNotImplemented || resp.StatusCode == http.StatusBadRequest,
		"expected 501 or 400, got %d", resp.StatusCode)

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HandleOCSP_InvalidRequest tests OCSP with invalid request data.
func TestCAServer_HandleOCSP_InvalidRequest(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := config.NewTestConfig("127.0.0.1", 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	time.Sleep(1 * time.Second)

	// Create HTTP client.
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test OCSP endpoint with invalid data.
	url := server.PublicBaseURL() + "/service/api/v1/ocsp"
	invalidData := []byte("invalid OCSP request data")
	req, err := http.NewRequest("POST", url, bytes.NewReader(invalidData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	// Should return error status.
	require.True(t, resp.StatusCode >= 400, "expected error status, got %d", resp.StatusCode)

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HandleCRLDistribution_Error tests CRL endpoint error handling.
func TestCAServer_HandleCRLDistribution_Error(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := config.NewTestConfig("127.0.0.1", 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	time.Sleep(1 * time.Second)

	// Create HTTP client.
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test CRL endpoint (should succeed or return error).
	url := server.PublicBaseURL() + "/service/api/v1/crl"
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	// Should return 200 with CRL or error status.
	if resp.StatusCode == http.StatusOK {
		// Verify Content-Type.
		require.Equal(t, "application/pkix-crl", resp.Header.Get("Content-Type"))

		// Verify response body is not empty.
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
	} else {
		// Error response should have JSON error.
		var errResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		require.Contains(t, errResp, "error")
	}

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HealthEndpoints_EdgeCases tests health endpoint edge cases.
func TestCAServer_HealthEndpoints_EdgeCases(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := config.NewTestConfig("127.0.0.1", 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	time.Sleep(500 * time.Millisecond)

	// Create HTTP client.
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test all health endpoints.
	endpoints := []string{"/health", "/livez", "/readyz"}
	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			url := server.PublicBaseURL() + endpoint
			resp, err := client.Get(url)
			require.NoError(t, err)
			defer func() {
				require.NoError(t, resp.Body.Close())
			}()

			// Should return 200.
			require.Equal(t, http.StatusOK, resp.StatusCode, "endpoint %s failed", endpoint)

			// Read body.
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)
		})
	}

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HealthEndpoints_NotReady tests health endpoints when server not ready.

