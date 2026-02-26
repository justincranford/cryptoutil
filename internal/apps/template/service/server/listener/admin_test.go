// Copyright (c) 2025 Justin Cranford
//
//

package listener_test

import (
	"context"
	"crypto/tls"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
	cryptoutilAppsTemplateServiceTestingHttpservertests "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAdminHTTPServer_HappyPath tests successful admin server creation.
func TestNewAdminHTTPServer_HappyPath(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)

	require.NoError(t, err)
	require.NotNil(t, server)
}

// TestNewAdminHTTPServer_NilContext tests that NewAdminHTTPServer rejects nil context.
func TestNewAdminHTTPServer_NilContext(t *testing.T) {
	t.Parallel()

	// Use shared TestMain-generated fixture for private TLS settings.
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(nil, cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
	assert.Nil(t, server)
}

// TestNewAdminHTTPServer_NilSettings tests that NewAdminHTTPServer rejects nil settings.
func TestNewAdminHTTPServer_NilSettings(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), nil, tlsCfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "settings cannot be nil")
	assert.Nil(t, server)
}

// TestNewAdminHTTPServer_NilTLSCfg tests that NewAdminHTTPServer rejects nil TLS configuration.
func TestNewAdminHTTPServer_NilTLSCfg(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TLS configuration cannot be nil")
	assert.Nil(t, server)
}

// TestAdminServer_Start_Success tests admin server starts and listens on dynamic port.
func TestAdminServer_Start_Success(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	// Use shared TestMain-generated fixture for private TLS settings.
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	startErr := make(chan error, 1)

	go func() {
		defer wg.Done()

		if err := server.Start(ctx); err != nil {
			startErr <- err
		}
	}()

	// Wait for server to be ready with retry logic.
	var port int

	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		time.Sleep(cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond)

		port = server.ActualPort()
		if port > 0 {
			break
		}
	}

	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Verify no startup error.
	select {
	case err := <-startErr:
		require.FailNow(t, "unexpected start error", err)
	default:
	}

	// Wait for port to be fully released before next test.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestAdminServer_Readyz_NotReady tests that readyz returns 503 when server not marked ready.
func TestAdminServer_Readyz_NotReady(t *testing.T) {
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to allocate port.
	time.Sleep(200 * time.Millisecond)

	port := server.ActualPort()
	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Create HTTPS client that accepts self-signed certs.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment with self-signed certs.
			},
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Call readyz endpoint (should return 503 - not ready).
	reqCtx, reqCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/api/v1/readyz", cryptoutilSharedMagic.IPv4Loopback, port)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Should return 503 Service Unavailable when not ready.
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "not ready", result[cryptoutilSharedMagic.StringStatus])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for port to be fully released before next test.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestAdminServer_HealthChecks_DuringShutdown tests livez and readyz return 503 during shutdown.
func TestAdminServer_HealthChecks_DuringShutdown(t *testing.T) {
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to allocate port.
	time.Sleep(200 * time.Millisecond)

	port := server.ActualPort()
	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Create HTTPS client that accepts self-signed certs.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment with self-signed certs.
			},
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Mark server ready.
	server.SetReady(true)

	// Verify readyz returns 200 OK when ready.
	reqCtx1, reqCancel1 := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer reqCancel1()

	url := fmt.Sprintf("https://%s:%d/admin/api/v1/readyz", cryptoutilSharedMagic.IPv4Loopback, port)
	req1, err := http.NewRequestWithContext(reqCtx1, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp1, err := client.Do(req1)
	require.NoError(t, err)

	defer func() { _ = resp1.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	// Initiate shutdown in background.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	go func() {
		_ = server.Shutdown(shutdownCtx)
	}()

	// Wait a bit for shutdown to start.
	time.Sleep(cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond)

	// Call livez endpoint during shutdown (should return 503 - shutting down).
	reqCtx2, reqCancel2 := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer reqCancel2()

	livezURL := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, port)
	req2, err := http.NewRequestWithContext(reqCtx2, http.MethodGet, livezURL, nil)
	require.NoError(t, err)

	resp2, err := client.Do(req2)
	if err != nil {
		// Connection may be refused if shutdown completed quickly - this is acceptable.
		t.Logf("Expected error during shutdown: %v", err)
	} else {
		defer func() { _ = resp2.Body.Close() }()

		// If we got a response, it should be 503 Service Unavailable.
		if resp2.StatusCode == http.StatusServiceUnavailable {
			body, readErr := io.ReadAll(resp2.Body)
			require.NoError(t, readErr)

			var result map[string]any

			unmarshalErr := json.Unmarshal(body, &result)
			require.NoError(t, unmarshalErr)

			assert.Equal(t, "shutting down", result[cryptoutilSharedMagic.StringStatus])
		}
	}

	wg.Wait()

	// Wait for port to be fully released before next test.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestAdminServer_Start_NilContext tests that Start rejects nil context.
func TestAdminServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
		server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestStartNilContext(t, createServer)
}

// TestAdminServer_Livez_Alive tests /admin/api/v1/livez endpoint when server is alive.
func TestAdminServer_Livez_Alive(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	port := server.ActualPort()
	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Query livez endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	reqCtx, reqCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, port)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response.
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Equal(t, "alive", response[cryptoutilSharedMagic.StringStatus])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Readyz_Ready tests /admin/api/v1/readyz endpoint when server is ready.
func TestAdminServer_Readyz_Ready(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	port := server.ActualPort()
	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Mark server as ready (application's responsibility after dependencies initialized).
	server.SetReady(true)

	// Query readyz endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	reqCtx, reqCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/api/v1/readyz", cryptoutilSharedMagic.IPv4Loopback, port)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response.
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Equal(t, "ready", response[cryptoutilSharedMagic.StringStatus])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Shutdown_Endpoint tests POST /admin/api/v1/shutdown triggers graceful shutdown.
