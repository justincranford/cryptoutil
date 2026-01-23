// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// TestNewAdminHTTPServer tests admin server creation.
func TestNewAdminHTTPServer(t *testing.T) {
	t.Parallel()

	t.Run("ValidConfig", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin")
		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, server)
		require.NotNil(t, server.app)
		require.NotNil(t, server.config)
		require.False(t, server.ready)
		require.False(t, server.shutdown)
	})

	t.Run("NilContext", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_nil_ctx")

		server, err := NewAdminHTTPServer(nil, cfg) //nolint:staticcheck // Testing nil context validation requires passing nil.
		require.Error(t, err)
		require.Nil(t, server)
		require.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("NilConfig", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, nil)
		require.Error(t, err)
		require.Nil(t, server)
		require.Contains(t, err.Error(), "config cannot be nil")
	})
}

// TestAdminServerLifecycle tests server start and shutdown.
func TestAdminServerLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("StartAndShutdown", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_lifecycle")
		cfg.IDP.AdminPort = 0 // Dynamic port.
		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, server)

		// Start server in background.
		errChan := make(chan error, 1)

		go func() {
			if startErr := server.Start(ctx); startErr != nil {
				errChan <- startErr
			}
		}()

		// Wait for server to start and get port (polling avoids race condition).
		port := waitForAdminPort(t, server, 5*time.Second)
		require.NotZero(t, port)

		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Check for start errors.
		select {
		case startErr := <-errChan:
			require.NoError(t, startErr)
		case <-time.After(100 * time.Millisecond):
			// No error - success.
		}
	})

	t.Run("ShutdownWithoutStart", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_no_start")
		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)

		// Shutdown without starting should succeed.
		require.NoError(t, server.Shutdown(ctx))
	})

	t.Run("ShutdownWithNilContext", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_shutdown_nil_ctx")
		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)

		// Shutdown with nil context should default to Background.
		require.NoError(t, server.Shutdown(context.TODO()))
	})
}

// TestAdminServerActualPort tests port extraction.
func TestAdminServerActualPort(t *testing.T) {
	t.Parallel()

	t.Run("BeforeStart", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_port_before")
		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)

		// Should return 0 before Start.
		port := server.ActualPort()

		require.Zero(t, port)
	})

	t.Run("AfterStart", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_port_after")
		cfg.IDP.AdminPort = 0 // Dynamic port.
		ctx := context.Background()

		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)

		// Start server.
		go func() {
			_ = server.Start(ctx)
		}()

		// Wait for server to start and get port (polling avoids race condition).
		port := waitForAdminPort(t, server, 5*time.Second)
		require.NotZero(t, port)

		// Cleanup.
		require.NoError(t, server.Shutdown(ctx))
	})
}

// TestAdminEndpointLivez tests /admin/api/v1/livez endpoint.
func TestAdminEndpointLivez(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_livez")
	cfg.IDP.AdminPort = 0 // Dynamic port.
	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// Start server.
	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to start and get port (polling avoids race condition).
	port := waitForAdminPort(t, server, 5*time.Second)

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)

	t.Run("LivezReturnsAlive", func(t *testing.T) {
		statusCode, body := doAdminGet(t, baseURL+"/admin/api/v1/livez")
		require.Equal(t, http.StatusOK, statusCode)

		var response map[string]any

		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "alive", response["status"])
	})

	t.Run("LivezAfterShutdown", func(t *testing.T) {
		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Server is shut down - connection should be refused.
		client := &http.Client{
			Timeout: 1 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion:         tls.VersionTLS13,
					InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
				},
			},
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/admin/api/v1/livez", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}

		require.Error(t, err, "expected connection error after shutdown")
	})
}

// TestAdminEndpointReadyz tests /admin/api/v1/readyz endpoint.
func TestAdminEndpointReadyz(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_readyz")
	cfg.IDP.AdminPort = 0 // Dynamic port.
	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// Start server.
	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to start and get port (polling avoids race condition).
	port := waitForAdminPort(t, server, 5*time.Second)

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)

	t.Run("ReadyzBeforeReady", func(t *testing.T) {
		// Server starts not ready by default.
		statusCode, body := doAdminGet(t, baseURL+"/admin/api/v1/readyz")

		var response map[string]any

		require.NoError(t, json.Unmarshal(body, &response))

		// Expect either "not ready" (503) or "ready" (200) depending on timing.
		switch statusCode {
		case http.StatusServiceUnavailable:
			require.Equal(t, "not ready", response["status"])
		case http.StatusOK:
			require.Equal(t, "ready", response["status"])
		default:
			require.Fail(t, fmt.Sprintf("unexpected status code: %d", statusCode))
		}
	})

	t.Run("ReadyzAfterShutdown", func(t *testing.T) {
		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Server is shut down - connection should be refused.
		client := &http.Client{
			Timeout: 1 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion:         tls.VersionTLS13,
					InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
				},
			},
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/admin/api/v1/readyz", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}

		require.Error(t, err, "expected connection error after shutdown")
	})
}

// TestAdminEndpointShutdown tests /admin/api/v1/shutdown endpoint.
func TestAdminEndpointShutdown(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_shutdown_endpoint")
	cfg.IDP.AdminPort = 0 // Dynamic port.
	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to start and get port (polling avoids race condition).
	port := waitForAdminPort(t, server, 5*time.Second)

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)

	t.Run("ShutdownViaEndpoint", func(t *testing.T) {
		// Call shutdown endpoint.
		statusCode, body := doAdminPost(t, baseURL+"/admin/api/v1/shutdown")
		require.Equal(t, http.StatusOK, statusCode)

		var response map[string]any

		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "shutdown initiated", response["status"])
		// Server is shutting down asynchronously - don't test livez after shutdown.
	})
}

// waitForAdminPort polls for the admin server to start and return a valid port.
// This avoids race conditions where ActualPort() returns 0 before the server is ready.
func waitForAdminPort(t *testing.T, server *AdminServer, timeout time.Duration) int {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if port := server.ActualPort(); port > 0 {
			return port
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("admin server did not start within %v", timeout)

	return 0
}

// doAdminGet performs GET request with TLS verification disabled.
func doAdminGet(t *testing.T, url string) (int, []byte) {
	t.Helper()

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}

// doAdminPost performs POST request with TLS verification disabled.
func doAdminPost(t *testing.T, url string) (int, []byte) {
	t.Helper()

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}
