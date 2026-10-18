// Copyright (c) 2025-2026 Justin Cranford.
//
//

package listener_test

import (
	"context"
	"crypto/tls"
	json "encoding/json"
	"io"
	"net"
	http "net/http"
	"strconv"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceServerListener "cryptoutil/internal/apps-framework/service/server/listener"
	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps-framework/service/server/testutil"
	cryptoutilAppsFrameworkServiceTestingHttpservertests "cryptoutil/internal/apps-framework/service/testing/httpservertests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForAdminHTTPSResponse(t *testing.T, client *http.Client, endpointURL string) (int, []byte) {
	t.Helper()

	var (
		statusCode int
		body       []byte
	)

	require.Eventually(t, func() bool {
		requestCtx, requestCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
		defer requestCancel()

		request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpointURL, nil)
		if err != nil {
			return false
		}

		response, err := client.Do(request)
		if err != nil {
			return false
		}

		responseBody, readErr := io.ReadAll(response.Body)

		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			return false
		}

		statusCode = response.StatusCode
		body = responseBody

		return statusCode == http.StatusOK
	}, cryptoutilSharedMagic.DefaultHealthCheckTimeout, cryptoutilSharedMagic.DefaultHealthCheckInterval, "admin HTTPS endpoint should return HTTP 200")

	return statusCode, body
}

// TestNewAdminHTTPServer_HappyPath tests successful admin server creation.
func TestNewAdminHTTPServer_HappyPath(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)

	require.NoError(t, err)
	require.NotNil(t, server)
}

// TestNewAdminHTTPServer_NilContext tests that NewAdminHTTPServer rejects nil context.
func TestNewAdminHTTPServer_NilContext(t *testing.T) {
	t.Parallel()

	// Use shared TestMain-generated fixture for private TLS settings.
	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(nil, cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
	assert.Nil(t, server)
}

// TestNewAdminHTTPServer_NilSettings tests that NewAdminHTTPServer rejects nil settings.
func TestNewAdminHTTPServer_NilSettings(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), nil, tlsCfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "settings cannot be nil")
	assert.Nil(t, server)
}

// TestNewAdminHTTPServer_NilTLSCfg tests that NewAdminHTTPServer rejects nil TLS configuration.
func TestNewAdminHTTPServer_NilTLSCfg(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TLS configuration cannot be nil")
	assert.Nil(t, server)
}

// TestAdminServer_Start_Success tests admin server starts and listens on dynamic port.
func TestAdminServer_Start_Success(t *testing.T) {
	t.Parallel()

	// NOT parallel - all admin server tests compete for port 9090.
	// Use shared TestMain-generated fixture for private TLS settings.
	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	startResult := make(chan error, 1)

	go func() {
		startResult <- server.Start(ctx)
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

	require.NoError(t, waitForServerStartResult(t, startResult))

	// Wait for port to be fully released before next test.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestAdminServer_Readyz_NotReady tests that readyz returns 503 when server not marked ready.
func TestAdminServer_Readyz_NotReady(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
	require.NoError(t, err)

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/api/v1/readyz", nil)
	require.NoError(t, err)

	request.Host = cryptoutilSharedMagic.HostnameLocalhost

	response, err := server.App().Test(request, -1)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }()

	require.Equal(t, http.StatusServiceUnavailable, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	require.Equal(t, "not ready", result[cryptoutilSharedMagic.StringStatus])
}

// TestAdminServer_HealthChecks_DuringShutdown tests livez and readyz before and during shutdown.
func TestAdminServer_HealthChecks_DuringShutdown(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
	require.NoError(t, err)

	server.SetReady(true)

	readyzRequest, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/api/v1/readyz", nil)
	require.NoError(t, err)

	readyzRequest.Host = cryptoutilSharedMagic.HostnameLocalhost

	readyzResponse, err := server.App().Test(readyzRequest, -1)
	require.NoError(t, err)

	defer func() { _ = readyzResponse.Body.Close() }()

	require.Equal(t, http.StatusOK, readyzResponse.StatusCode)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultServerShutdownTimeout)
	defer shutdownCancel()

	require.NoError(t, server.Shutdown(shutdownCtx))

	for _, path := range []string{"/admin/api/v1/livez", "/admin/api/v1/readyz"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			request, requestErr := http.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
			require.NoError(t, requestErr)

			request.Host = cryptoutilSharedMagic.HostnameLocalhost

			response, requestErr := server.App().Test(request, -1)
			require.NoError(t, requestErr)

			defer func() { _ = response.Body.Close() }()

			require.Equal(t, http.StatusServiceUnavailable, response.StatusCode)

			body, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)

			var result map[string]any
			require.NoError(t, json.Unmarshal(body, &result))
			require.Equal(t, cryptoutilSharedMagic.StringStatusShuttingDown, result[cryptoutilSharedMagic.StringStatus])
		})
	}
}

// TestAdminServer_Start_NilContext tests that Start rejects nil context.
func TestAdminServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsFrameworkServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
		server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsFrameworkServiceTestingHttpservertests.TestStartNilContext(t, createServer)
}

// TestAdminServer_Livez_Alive tests /admin/api/v1/livez endpoint when server is alive.
func TestAdminServer_Livez_Alive(t *testing.T) {
	t.Parallel()

	// NOT parallel - all admin server tests compete for port 9090.
	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	startResult := make(chan error, 1)

	go func() {
		startResult <- server.Start(ctx)
	}()

	// Wait for server to be ready.
	require.Eventually(t, func() bool { return server.ActualPort() > 0 },
		cryptoutilSharedMagic.DefaultHealthCheckTimeout, cryptoutilSharedMagic.DefaultHealthCheckInterval,
		"server should allocate a dynamic port")

	port := server.ActualPort()

	// Query livez endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    cryptoutilAppsFrameworkServiceServerTestutil.PrivateRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	url := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + "/admin/api/v1/livez"

	statusCode, body := waitForAdminHTTPSResponse(t, client, url)
	assert.Equal(t, http.StatusOK, statusCode)

	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Equal(t, "alive", response[cryptoutilSharedMagic.StringStatus])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	require.NoError(t, waitForServerStartResult(t, startResult))

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Readyz_Ready tests /admin/api/v1/readyz endpoint when server is ready.
func TestAdminServer_Readyz_Ready(t *testing.T) {
	t.Parallel()

	// NOT parallel - all admin server tests compete for port 9090.
	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	startResult := make(chan error, 1)

	go func() {
		startResult <- server.Start(ctx)
	}()

	// Wait for server to be ready.
	require.Eventually(t, func() bool { return server.ActualPort() > 0 },
		cryptoutilSharedMagic.DefaultHealthCheckTimeout, cryptoutilSharedMagic.DefaultHealthCheckInterval,
		"server should allocate a dynamic port")

	port := server.ActualPort()

	// Mark server as ready (application's responsibility after dependencies initialized).
	server.SetReady(true)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    cryptoutilAppsFrameworkServiceServerTestutil.PrivateRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	url := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) +
		cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath

	statusCode, body := waitForAdminHTTPSResponse(t, client, url)
	assert.Equal(t, http.StatusOK, statusCode)

	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Equal(t, "ready", response[cryptoutilSharedMagic.StringStatus])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	require.NoError(t, waitForServerStartResult(t, startResult))

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Shutdown_Endpoint tests POST /admin/api/v1/shutdown triggers graceful shutdown.
