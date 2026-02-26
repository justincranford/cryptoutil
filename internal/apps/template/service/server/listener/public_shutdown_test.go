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

func TestPublicHTTPServer_ServiceHealth_DuringShutdown(t *testing.T) {
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

	// Create client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Trigger shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// Make request during/after shutdown - should return 503.
	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.ActualPort())
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

			require.Equal(t, "shutting down", result[cryptoutilSharedMagic.StringStatus])
		}
	}

	// Cleanup.
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestPublicHTTPServer_BrowserHealth_DuringShutdown tests browser health during shutdown.
func TestPublicHTTPServer_BrowserHealth_DuringShutdown(t *testing.T) {
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

	// Create client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Trigger shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// Make request during/after shutdown - should return 503.
	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.ActualPort())
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

			require.Equal(t, "shutting down", result[cryptoutilSharedMagic.StringStatus])
		}
	}

	// Cleanup.
	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestPublicHTTPServer_Shutdown_DoubleCall tests calling Shutdown twice returns error.
func TestPublicHTTPServer_Shutdown_DoubleCall(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()
		server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestShutdownDoubleCall(t, createServer)
}

// TestPublicHTTPServer_PublicBaseURL tests PublicBaseURL returns correct URL format.
func TestPublicHTTPServer_PublicBaseURL(t *testing.T) {
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

	// Test PublicBaseURL returns correct format.
	baseURL := server.PublicBaseURL()
	expectedURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, port)
	require.Equal(t, expectedURL, baseURL)

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP * time.Millisecond)
}

// TestPublicHTTPServer_ServiceHealth_DuringShutdown_InMemory tests that /service/api/v1/health returns 503 during shutdown.
// Uses app.Test() for deterministic in-memory testing without HTTPS listener.
func TestPublicHTTPServer_ServiceHealth_DuringShutdown_InMemory(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// Call Shutdown to set server state to shutdown.
	err = server.Shutdown(context.Background())
	require.NoError(t, err)

	// Use app.Test() to make in-memory request after shutdown.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, cryptoutilSharedMagic.IdentityE2EHealthEndpoint, nil)
	require.NoError(t, err)

	resp, err := server.App().Test(req, -1)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	// Verify service unavailable during shutdown.
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.Equal(t, "shutting down", result[cryptoutilSharedMagic.StringStatus])
}

// TestPublicHTTPServer_BrowserHealth_DuringShutdown_InMemory tests that /browser/api/v1/health returns 503 during shutdown.
// Uses app.Test() for deterministic in-memory testing without HTTPS listener.
func TestPublicHTTPServer_BrowserHealth_DuringShutdown_InMemory(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// Call Shutdown to set server state to shutdown.
	err = server.Shutdown(context.Background())
	require.NoError(t, err)

	// Use app.Test() to make in-memory request after shutdown.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/browser/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := server.App().Test(req, -1)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	// Verify service unavailable during shutdown.
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.Equal(t, "shutting down", result[cryptoutilSharedMagic.StringStatus])
}

// TestPublicHTTPServer_Shutdown_Idempotent tests that shutdown behavior is consistent.
// Note: Public server returns error on subsequent shutdown calls (different from admin server).
func TestPublicHTTPServer_Shutdown_Idempotent(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

	server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// First shutdown call should succeed.
	err = server.Shutdown(context.Background())
	require.NoError(t, err)

	// Subsequent shutdown calls should return error (server already shutdown).
	err = server.Shutdown(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "already shutdown")
}
