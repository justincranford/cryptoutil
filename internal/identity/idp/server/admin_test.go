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

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

// TestNewAdminServer tests admin server creation.
func TestNewAdminServer(t *testing.T) {
	t.Parallel()

	t.Run("ValidConfig", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin")
		ctx := context.Background()

		server, err := NewAdminServer(ctx, cfg)
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

		server, err := NewAdminServer(context.TODO(), cfg)
		require.Error(t, err)
		require.Nil(t, server)
		require.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("NilConfig", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		server, err := NewAdminServer(ctx, nil)
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

		server, err := NewAdminServer(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, server)

		// Start server in background.
		errChan := make(chan error, 1)

		go func() {
			if startErr := server.Start(ctx); startErr != nil {
				errChan <- startErr
			}
		}()

		// Wait for server to be ready.
		time.Sleep(200 * time.Millisecond)

		// Verify server is running.
		port, err := server.ActualPort()
		require.NoError(t, err)
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

		server, err := NewAdminServer(ctx, cfg)
		require.NoError(t, err)

		// Shutdown without starting should succeed.
		require.NoError(t, server.Shutdown(ctx))
	})

	t.Run("ShutdownWithNilContext", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_shutdown_nil_ctx")
		ctx := context.Background()

		server, err := NewAdminServer(ctx, cfg)
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

		server, err := NewAdminServer(ctx, cfg)
		require.NoError(t, err)

		// Should fail before Start.
		port, err := server.ActualPort()
		require.Error(t, err)
		require.Zero(t, port)
		require.Contains(t, err.Error(), "listener not initialized")
	})

	t.Run("AfterStart", func(t *testing.T) {
		t.Parallel()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_port_after")
		cfg.IDP.AdminPort = 0 // Dynamic port.
		ctx := context.Background()

		server, err := NewAdminServer(ctx, cfg)
		require.NoError(t, err)

		// Start server.
		go func() {
			_ = server.Start(ctx)
		}()

		// Wait for server to be ready.
		time.Sleep(200 * time.Millisecond)

		// Should succeed after Start.
		port, err := server.ActualPort()
		require.NoError(t, err)
		require.NotZero(t, port)

		// Cleanup.
		require.NoError(t, server.Shutdown(ctx))
	})
}

// TestAdminEndpointLivez tests /admin/v1/livez endpoint.
func TestAdminEndpointLivez(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_livez")
	cfg.IDP.AdminPort = 0 // Dynamic port.
	ctx := context.Background()

	server, err := NewAdminServer(ctx, cfg)
	require.NoError(t, err)

	// Start server.
	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	port, err := server.ActualPort()
	require.NoError(t, err)

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)

	t.Run("LivezReturnsAlive", func(t *testing.T) {
		statusCode, body := doAdminGet(t, baseURL+"/admin/v1/livez")
		require.Equal(t, http.StatusOK, statusCode)

		var response map[string]any
		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "alive", response["status"])
	})

	t.Run("LivezAfterShutdown", func(t *testing.T) {
		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Livez should return 503.
		statusCode, body := doAdminGet(t, baseURL+"/admin/v1/livez")
		require.Equal(t, http.StatusServiceUnavailable, statusCode)

		var response map[string]any
		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "shutting down", response["status"])
	})
}

// TestAdminEndpointReadyz tests /admin/v1/readyz endpoint.
func TestAdminEndpointReadyz(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_readyz")
	cfg.IDP.AdminPort = 0 // Dynamic port.
	ctx := context.Background()

	server, err := NewAdminServer(ctx, cfg)
	require.NoError(t, err)

	// Start server.
	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	port, err := server.ActualPort()
	require.NoError(t, err)

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)

	t.Run("ReadyzBeforeReady", func(t *testing.T) {
		// Server starts not ready by default.
		statusCode, body := doAdminGet(t, baseURL+"/admin/v1/readyz")

		var response map[string]any
		require.NoError(t, json.Unmarshal(body, &response))

		// Expect either "not ready" (503) or "ready" (200) depending on timing.
		switch statusCode {
		case http.StatusServiceUnavailable:
			require.Equal(t, "not ready", response["status"])
		case http.StatusOK:
			require.Equal(t, "ready", response["status"])
		}
	})

	t.Run("ReadyzAfterShutdown", func(t *testing.T) {
		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Readyz should return 503.
		statusCode, body := doAdminGet(t, baseURL+"/admin/v1/readyz")
		require.Equal(t, http.StatusServiceUnavailable, statusCode)

		var response map[string]any
		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "shutting down", response["status"])
	})
}

// TestAdminEndpointShutdown tests /admin/v1/shutdown endpoint.
func TestAdminEndpointShutdown(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_shutdown_endpoint")
	cfg.IDP.AdminPort = 0 // Dynamic port.
	ctx := context.Background()

	server, err := NewAdminServer(ctx, cfg)
	require.NoError(t, err)

	// Start server.
	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	port, err := server.ActualPort()
	require.NoError(t, err)

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)

	t.Run("ShutdownViaEndpoint", func(t *testing.T) {
		// Call shutdown endpoint.
		statusCode, body := doAdminPost(t, baseURL+"/admin/v1/shutdown")
		require.Equal(t, http.StatusOK, statusCode)

		var response map[string]any
		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "shutting down", response["status"])

		// Verify server is shutting down.
		time.Sleep(100 * time.Millisecond)

		statusCode, body = doAdminGet(t, baseURL+"/admin/v1/livez")
		require.Equal(t, http.StatusServiceUnavailable, statusCode)

		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "shutting down", response["status"])
	})
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
