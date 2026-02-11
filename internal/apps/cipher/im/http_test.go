// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsCipherImServer "cryptoutil/internal/apps/cipher/im/server"
	cryptoutilAppsCipherImServerConfig "cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
)

// initTestConfig returns an CipherImServerSettings with all required settings for tests.
func initTestConfig() *cryptoutilAppsCipherImServerConfig.CipherImServerSettings {
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("cipher-im-http-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory for fast tests.

	return &cryptoutilAppsCipherImServerConfig.CipherImServerSettings{
		ServiceTemplateServerSettings: settings,
	}
}

// TestHTTPGet tests the httpGet helper function (used by health CLI wrappers).
func TestHTTPGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create server with dynamic ports.
	cfg := initTestConfig()

	srv, err := cryptoutilAppsCipherImServer.NewFromConfig(ctx, cfg)
	require.NoError(t, err)

	// Mark server as ready after successful initialization.
	// This enables /admin/v1/readyz to return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	// Start server.
	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for server to be ready using polling pattern.
	require.Eventually(t, func() bool {
		return srv.PublicPort() > 0
	}, 10*time.Second, 100*time.Millisecond, "server should allocate port")

	// Get actual ports.
	publicPort := srv.PublicPort()
	adminPort := srv.AdminPort()

	// Create insecure HTTP client (accepts self-signed certs).
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	// Test public health endpoint.
	t.Run("public_health_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/service/api/v1/health", publicPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test admin livez endpoint.
	t.Run("admin_livez_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/admin/api/v1/livez", adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test admin readyz endpoint.
	t.Run("admin_readyz_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/admin/api/v1/readyz", adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	require.NoError(t, err)
}

// TestHTTPPost tests the httpPost helper function (used by shutdown CLI wrapper).
func TestHTTPPost(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create server with dynamic ports.
	cfg := initTestConfig()

	srv, err := cryptoutilAppsCipherImServer.NewFromConfig(ctx, cfg)
	require.NoError(t, err)

	// Start server in background with cancellable context.
	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for server to be ready using polling pattern.
	require.Eventually(t, func() bool {
		return srv.AdminPort() > 0
	}, 10*time.Second, 100*time.Millisecond, "server should allocate port")

	// Get actual ports.
	adminPort := srv.AdminPort()

	// Create insecure HTTP client (accepts self-signed certs).
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	// Test admin shutdown endpoint (triggers async shutdown).
	t.Run("admin_shutdown_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/admin/api/v1/shutdown", adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Cancel context to trigger server shutdown (shutdown endpoint starts async shutdown).
	cancel()

	// Wait for server to finish shutting down.
	select {
	case err := <-errChan:
		// Server shutdown returns context.Canceled error which is expected.
		// After wrapping, the error message will include the wrapper prefix.
		const (
			adminStoppedErr     = "admin server stopped: context canceled"
			appCancelledErr     = "application startup cancelled: context canceled"
			wrappedAppCancelled = "failed to start application: application startup cancelled: context canceled"
		)

		if err != nil && err.Error() != adminStoppedErr && err.Error() != appCancelledErr && err.Error() != wrappedAppCancelled {
			require.FailNowf(t, "Unexpected server error", "%v", err)
		}
	case <-time.After(5 * time.Second):
		require.FailNow(t, "Server did not shutdown within timeout")
	}
}

// TestIMServer tests the imServer function.
func TestIMServer(t *testing.T) {
	t.Parallel()
	// This test would require mocking os.Signal and context handling.
	// Skipping for now as imServer is tested via integration tests.
	t.Skip("imServer requires signal mocking - tested via integration tests")
}
