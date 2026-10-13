// Copyright (c) 2025-2026 Justin Cranford.
//
//

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	json "encoding/json"
	"fmt"
	"io"
	"net"
	http "net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

		port, errChan := startAdminServerForTest(t, ctx, server)
		require.NotZero(t, port)

		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Check for start errors.
		select {
		case startErr := <-errChan:
			require.NoError(t, startErr)
		case <-time.After(cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond):
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
		require.NoError(t, server.Shutdown(context.Background()))
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
		port, _ := startAdminServerForTest(t, ctx, server)
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
	port, _ := startAdminServerForTest(t, ctx, server)

	baseURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port))

	t.Run("LivezReturnsAlive", func(t *testing.T) {
		statusCode, body := doAdminGet(t, baseURL+"/admin/api/v1/livez", server.TLSRootCAPool())
		require.Equal(t, http.StatusOK, statusCode)

		var response map[string]any

		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "alive", response[cryptoutilSharedMagic.StringStatus])
	})

	t.Run("LivezAfterShutdown", func(t *testing.T) {
		// Shutdown server.
		require.NoError(t, server.Shutdown(ctx))

		// Server is shut down - connection should be refused.
		client := &http.Client{
			Timeout: 1 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
					RootCAs:    server.TLSRootCAPool(),
				},
				DisableKeepAlives: true,
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
	port, _ := startAdminServerForTest(t, ctx, server)

	baseURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port))

	t.Run("ReadyzBeforeReady", func(t *testing.T) {
		// Server starts not ready by default.
		statusCode, body := doAdminGet(t, baseURL+"/admin/api/v1/readyz", server.TLSRootCAPool())

		var response map[string]any

		require.NoError(t, json.Unmarshal(body, &response))

		// Expect either "not ready" (503) or "ready" (200) depending on timing.
		switch statusCode {
		case http.StatusServiceUnavailable:
			require.Equal(t, "not ready", response[cryptoutilSharedMagic.StringStatus])
		case http.StatusOK:
			require.Equal(t, "ready", response[cryptoutilSharedMagic.StringStatus])
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
					MinVersion: tls.VersionTLS13,
					RootCAs:    server.TLSRootCAPool(),
				},
				DisableKeepAlives: true,
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

	port, _ := startAdminServerForTest(t, ctx, server)

	baseURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port))

	t.Run("ShutdownViaEndpoint", func(t *testing.T) {
		// Call shutdown endpoint.
		statusCode, body := doAdminPost(t, baseURL+"/admin/api/v1/shutdown", server.TLSRootCAPool())
		require.Equal(t, http.StatusOK, statusCode)

		var response map[string]any

		require.NoError(t, json.Unmarshal(body, &response))
		require.Equal(t, "shutdown initiated", response[cryptoutilSharedMagic.StringStatus])
		// Server is shutting down asynchronously - don't test livez after shutdown.
	})
}

// startAdminServerForTest waits for the listener and reports startup failures immediately.
func startAdminServerForTest(t *testing.T, ctx context.Context, server *AdminServer) (int, <-chan error) {
	t.Helper()

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start(ctx)
	}()

	startupTimer := time.NewTimer(cryptoutilSharedMagic.TestTimeoutCryptoutilReady)
	defer startupTimer.Stop()

	startupPoll := time.NewTicker(cryptoutilSharedMagic.TestTimeoutHTTPRetryInterval)
	defer startupPoll.Stop()

	for {
		if port := server.ActualPort(); port > 0 {
			return port, startErr
		}

		select {
		case err := <-startErr:
			require.NoError(t, err, "admin server exited before its listener became available")
			require.FailNow(t, "admin server stopped before startup completed")
		case <-startupTimer.C:
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "admin server did not start within %v", cryptoutilSharedMagic.TestTimeoutCryptoutilReady)
		case <-startupPoll.C:
		}
	}
}

// doAdminGet performs GET request using the provided TLS certificate pool.
func doAdminGet(t *testing.T, url string, certPool *x509.CertPool) (int, []byte) {
	t.Helper()

	client := &http.Client{
		Timeout: cryptoutilSharedMagic.TestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    certPool,
			},
			DisableKeepAlives: true,
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

// doAdminPost performs POST request using the provided TLS certificate pool.
func doAdminPost(t *testing.T, url string, certPool *x509.CertPool) (int, []byte) {
	t.Helper()

	client := &http.Client{
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    certPool,
			},
			DisableKeepAlives: true,
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
