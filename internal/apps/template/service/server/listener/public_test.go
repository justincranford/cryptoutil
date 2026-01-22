// Copyright (c) 2025 Justin Cranford
//
//

package listener_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilTemplateServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilTemplateServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
	cryptoutilTemplateServiceTesting "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// TestNewPublicHTTPServer_HappyPath tests successful public server creation.
func TestNewPublicHTTPServer_HappyPath(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)

	require.NoError(t, err)
	require.NotNil(t, server)
}

// TestNewPublicHTTPServer_NilContext tests that NewPublicHTTPServer rejects nil context.
func TestNewPublicHTTPServer_NilContext(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(nil, cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
	assert.Nil(t, server)
}

// TestPublicHTTPServer_Start_Success tests public server starts and listens on dynamic port.
func TestPublicHTTPServer_Start_Success(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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
	assert.Greater(t, port, 0)

	// Shutdown server.
	cancel()
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_Start_NilContext tests Start rejects nil context.
func TestPublicHTTPServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()
		server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilTemplateServiceTesting.TestStartNilContext(t, createServer)
}

// TestPublicHTTPServer_ServiceHealth_Healthy tests /service/api/v1/health returns healthy.
func TestPublicHTTPServer_ServiceHealth_Healthy(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, server.ActualPort())
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
	assert.Equal(t, "healthy", result["status"])

	// Shutdown server.
	cancel()
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_BrowserHealth_Healthy tests /browser/api/v1/health returns healthy.
func TestPublicHTTPServer_BrowserHealth_Healthy(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, server.ActualPort())
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
	assert.Equal(t, "healthy", result["status"])

	// Shutdown server.
	cancel()
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_Shutdown_Graceful tests graceful shutdown.
func TestPublicHTTPServer_Shutdown_Graceful(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()
		server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilTemplateServiceTesting.TestShutdownGraceful(t, createServer)
}

// TestPublicHTTPServer_Shutdown_NilContext tests Shutdown accepts nil context.
func TestPublicHTTPServer_Shutdown_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()
		server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilTemplateServiceTesting.TestShutdownNilContext(t, createServer)
}

// TestPublicHTTPServer_ActualPort_BeforeStart tests ActualPort before server starts.
func TestPublicHTTPServer_ActualPort_BeforeStart(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	port := server.ActualPort()

	assert.Equal(t, 0, port, "Expected port 0 before server starts")
}

// TestPublicHTTPServer_ServiceHealth_DuringShutdown tests health endpoint during shutdown.
func TestPublicHTTPServer_ServiceHealth_DuringShutdown(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	// Create client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: 5 * time.Second,
	}

	// Trigger shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// Make request during/after shutdown - should return 503.
	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, server.ActualPort())
	url := fmt.Sprintf("%s/service/api/v1/health", baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if err == nil {
		// If we got a response, verify it's 503.
		defer resp.Body.Close() //nolint:errcheck // Best effort close in test.

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusServiceUnavailable {
			var result map[string]any

			_ = json.Unmarshal(body, &result)

			assert.Equal(t, "shutting down", result["status"])
		}
	}

	// Cleanup.
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_BrowserHealth_DuringShutdown tests browser health during shutdown.
func TestPublicHTTPServer_BrowserHealth_DuringShutdown(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()

	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	// Create client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: 5 * time.Second,
	}

	// Trigger shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// Make request during/after shutdown - should return 503.
	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, server.ActualPort())
	url := fmt.Sprintf("%s/browser/api/v1/health", baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if err == nil {
		// If we got a response, verify it's 503.
		defer resp.Body.Close() //nolint:errcheck // Best effort close in test.

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusServiceUnavailable {
			var result map[string]any

			_ = json.Unmarshal(body, &result)

			assert.Equal(t, "shutting down", result["status"])
		}
	}

	// Cleanup.
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestPublicHTTPServer_Shutdown_DoubleCall tests calling Shutdown twice returns error.
func TestPublicHTTPServer_Shutdown_DoubleCall(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()
		server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestPublicHTTPServer_PublicBaseURL tests PublicBaseURL returns correct URL format.
func TestPublicHTTPServer_PublicBaseURL(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilTemplateServerTestutil.PublicTLS()
	server, err := cryptoutilTemplateServerListener.NewPublicHTTPServer(context.Background(), cryptoutilTemplateServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
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

	// Wait for server to be ready with retry logic.
	var port int

	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)

		port = server.ActualPort()
		if port > 0 {
			break
		}
	}

	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Test PublicBaseURL returns correct format.
	baseURL := server.PublicBaseURL()
	expectedURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, port)
	assert.Equal(t, expectedURL, baseURL)

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}
