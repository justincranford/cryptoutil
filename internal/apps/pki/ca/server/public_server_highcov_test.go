package server

import (
	"bytes"
	"context"
	"crypto/tls"
	json "encoding/json"
	"io"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki/ca/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

// pollServerReady timeout and interval for waiting on server port allocation.

// TestCAServer_HandleOCSP tests the OCSP endpoint.
func TestCAServer_HandleOCSP(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	err = cryptoutilSharedUtilPoll.Until(ctx, cryptoutilSharedMagic.TestPollReadyTimeout, cryptoutilSharedMagic.TestPollReadyInterval, func(_ context.Context) (bool, error) {
		return server.PublicPort() > 0, nil
	})
	require.NoError(t, err, "server did not bind to port")

	// Create HTTP client.
	client := &http.Client{
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test OCSP endpoint with empty request (should return 400 or 501).
	url := server.PublicBaseURL() + "/service/api/v1/ocsp"
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader([]byte{}))
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HandleOCSP_InvalidRequest tests OCSP with invalid request data.
func TestCAServer_HandleOCSP_InvalidRequest(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	err = cryptoutilSharedUtilPoll.Until(ctx, cryptoutilSharedMagic.TestPollReadyTimeout, cryptoutilSharedMagic.TestPollReadyInterval, func(_ context.Context) (bool, error) {
		return server.PublicPort() > 0, nil
	})
	require.NoError(t, err, "server did not bind to port")

	// Create HTTP client.
	client := &http.Client{
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test OCSP endpoint with invalid data.
	url := server.PublicBaseURL() + "/service/api/v1/ocsp"
	invalidData := []byte("invalid OCSP request data")
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(invalidData))
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HandleCRLDistribution_Error tests CRL endpoint error handling.
func TestCAServer_HandleCRLDistribution_Error(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	err = cryptoutilSharedUtilPoll.Until(ctx, cryptoutilSharedMagic.TestPollReadyTimeout, cryptoutilSharedMagic.TestPollReadyInterval, func(_ context.Context) (bool, error) {
		return server.PublicPort() > 0, nil
	})
	require.NoError(t, err, "server did not bind to port")

	// Create HTTP client.
	client := &http.Client{
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test CRL endpoint (should succeed or return error).
	url := server.PublicBaseURL() + "/service/api/v1/crl"
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
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
		var errResp map[string]any

		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		require.Contains(t, errResp, cryptoutilSharedMagic.StringError)
	}

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HealthEndpoints_EdgeCases tests health endpoint edge cases.
// Health endpoints are provided by service-template:
// - Admin: /admin/api/v1/livez, /admin/api/v1/readyz (via AdminServer)
// - Public: /service/api/v1/health, /browser/api/v1/health (via PublicServerBase).
func TestCAServer_HealthEndpoints_EdgeCases(t *testing.T) {
	t.Parallel()

	// Create and start test server.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	ctx := context.Background()
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready and ports to be allocated.
	err = cryptoutilSharedUtilPoll.Until(ctx, cryptoutilSharedMagic.TestPollReadyTimeout, cryptoutilSharedMagic.TestPollReadyInterval, func(_ context.Context) (bool, error) {
		return server.PublicPort() > 0 && server.AdminPort() > 0, nil
	})
	require.NoError(t, err, "server did not bind to ports")

	// Create HTTP client.
	client := &http.Client{
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
	}

	// Test public health endpoint.
	t.Run(cryptoutilSharedMagic.IdentityE2EHealthEndpoint, func(t *testing.T) {
		url := server.PublicBaseURL() + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, resp.Body.Close())
		}()

		require.Equal(t, http.StatusOK, resp.StatusCode, "public health endpoint failed")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
	})

	// Test admin livez endpoint.
	t.Run("/admin/api/v1/livez", func(t *testing.T) {
		url := server.AdminBaseURL() + "/admin/api/v1/livez"
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, resp.Body.Close())
		}()

		require.Equal(t, http.StatusOK, resp.StatusCode, "admin livez endpoint failed")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
	})

	// Test admin readyz endpoint.
	t.Run("/admin/api/v1/readyz", func(t *testing.T) {
		url := server.AdminBaseURL() + "/admin/api/v1/readyz"
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, resp.Body.Close())
		}()

		// Readyz may return 503 if server is not fully ready yet.
		require.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable,
			"admin readyz endpoint returned unexpected status: %d", resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
	})

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
}

// TestCAServer_HealthEndpoints_NotReady tests health endpoints when server not ready.
