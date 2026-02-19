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

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
	cryptoutilAppsTemplateServiceTestingHttpservertests "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestNewPublicHTTPServer_HappyPath tests successful public server creation.
func TestNewPublicHTTPServer_HappyPath(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)

	require.NoError(t, err)
	require.NotNil(t, server)
}

// TestNewPublicHTTPServer_NilContext tests that NewPublicHTTPServer rejects nil context.
func TestNewPublicHTTPServer_NilContext(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(nil, cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
	require.Nil(t, server)
}

// TestNewPublicHTTPServer_NilSettings tests that NewPublicHTTPServer rejects nil settings.
func TestNewPublicHTTPServer_NilSettings(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), nil, tlsCfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "settings cannot be nil")
	require.Nil(t, server)
}

// TestNewPublicHTTPServer_NilTLSCfg tests that NewPublicHTTPServer rejects nil TLS configuration.
func TestNewPublicHTTPServer_NilTLSCfg(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "TLS config cannot be nil")
	require.Nil(t, server)
}

// TestPublicHTTPServer_Start_Success tests public server starts and listens on dynamic port.
func TestPublicHTTPServer_Start_Success(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	// Wait for server to start.
	time.Sleep(1 * time.Second)

	// Verify server is listening on a port.
	port := server.ActualPort()
	require.Greater(t, port, 0)

	// Shutdown server.
	cancel()
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_Start_NilContext tests Start rejects nil context.
func TestPublicHTTPServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()
		server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestStartNilContext(t, createServer)
}

// TestPublicHTTPServer_ServiceHealth_Healthy tests /service/api/v1/health returns healthy.
func TestPublicHTTPServer_ServiceHealth_Healthy(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	// Wait for server to start.
	time.Sleep(1 * time.Second)

	// Make request to service health endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: 5 * time.Second,
	}

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.ActualPort())
	url := fmt.Sprintf("%s/service/api/v1/health", baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.Equal(t, "healthy", result["status"])

	// Shutdown server.
	cancel()
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_BrowserHealth_Healthy tests /browser/api/v1/health returns healthy.
func TestPublicHTTPServer_BrowserHealth_Healthy(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	// Wait for server to start.
	time.Sleep(1 * time.Second)

	// Make request to browser health endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: 5 * time.Second,
	}

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.ActualPort())
	url := fmt.Sprintf("%s/browser/api/v1/health", baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.Equal(t, "healthy", result["status"])

	// Shutdown server.
	cancel()
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_Shutdown_Graceful tests graceful shutdown.
func TestPublicHTTPServer_Shutdown_Graceful(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()
		server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestShutdownGraceful(t, createServer)
}

// TestPublicHTTPServer_Shutdown_NilContext tests Shutdown accepts nil context.
func TestPublicHTTPServer_Shutdown_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()
		server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestShutdownNilContext(t, createServer)
}

// TestPublicHTTPServer_ActualPort_BeforeStart tests ActualPort before server starts.
func TestPublicHTTPServer_ActualPort_BeforeStart(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	port := server.ActualPort()

	require.Equal(t, 0, port, "Expected port 0 before server starts")
}

// TestPublicHTTPServer_ServiceHealth_DuringShutdown tests health endpoint during shutdown.
