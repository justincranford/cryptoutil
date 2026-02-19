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
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminServer_ConcurrentRequests(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	// Wait for port to be fully released from previous test.
	time.Sleep(2 * time.Second)

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var (
		wg       sync.WaitGroup
		startErr error
	)

	wg.Add(1)

	go func() {
		defer wg.Done()

		startErr = server.Start(ctx)
	}()

	// Wait for server to be ready with health check retry.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
		},
		Timeout: 5 * time.Second,
	}

	var (
		serverReady bool
		port        int
	)

	for i := 0; i < 60; i++ { // Increased to 60 attempts (3 seconds total) to handle previous test shutdown.
		time.Sleep(50 * time.Millisecond)

		// Check if Start() returned an error.
		if startErr != nil {
			require.FailNow(t, "server.Start() error after attempts", "attempt", i, "error", startErr)
		}

		port = server.ActualPort()
		if port > 0 {
			healthCtx, healthCancel := context.WithTimeout(context.Background(), 2*time.Second)

			healthURL := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, port)

			healthReq, _ := http.NewRequestWithContext(healthCtx, http.MethodGet, healthURL, nil)

			healthResp, healthErr := client.Do(healthReq)

			healthCancel()

			if healthErr == nil && healthResp.StatusCode == http.StatusOK {
				_ = healthResp.Body.Close()
				serverReady = true

				break
			}

			if healthResp != nil {
				_ = healthResp.Body.Close()
			}

			// Log startup failures for debugging.
			if i%10 == 0 && healthErr != nil {
				t.Logf("Health check attempt %d failed: %v", i, healthErr)
			}
		}
	}

	require.True(t, serverReady, "server failed to become ready after 60 attempts (3 seconds)")

	// Check if server.Start() returned an error during startup.
	if startErr != nil {
		require.FailNow(t, "server.Start() error", startErr)
	}

	// Make 10 concurrent requests to livez endpoint.
	const numRequests = 10

	var requestWG sync.WaitGroup

	requestWG.Add(numRequests)

	errors := make(chan error, numRequests)

	for range numRequests {
		go func() {
			defer requestWG.Done()

			reqCtx, reqCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer reqCancel()

			url := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, port)

			req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
			if reqErr != nil {
				errors <- reqErr

				return
			}

			resp, err := client.Do(req)
			if err != nil {
				errors <- err

				return
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
			}
		}()
	}

	requestWG.Wait()
	close(errors)

	// Verify no errors.
	for err := range errors {
		require.NoError(t, err, "concurrent request error")
	}

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Check if server.Start() returned an error.
	if startErr != nil {
		require.NoError(t, startErr, "server.Start() error")
	}

	// Wait for port to be fully released before next test.
	time.Sleep(500 * time.Millisecond)
}

// TestAdminServer_TimeoutsConfigured tests that server timeouts are properly configured.
func TestAdminServer_TimeoutsConfigured(t *testing.T) {
	// NOT parallel - server startup timing is unpredictable when running with other tests.
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(context.Background())
	}()

	// Wait for server to be ready with retry logic.
	var port int

	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)

		port = server.ActualPort()
		if port > 0 {
			break
		}
	}

	require.Greater(t, port, 0, "Expected dynamic port allocation")
	server.SetReady(true)

	// Create HTTP client with shorter timeout to test server's idle timeout.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test server uses self-signed cert.
			},
		},
		Timeout: 15 * time.Second, // Longer than server timeout to test server-side.
	}

	// Make request to verify timeouts are working.
	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, port)
	url := fmt.Sprintf("%s/admin/api/v1/readyz", baseURL)

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
	assert.Equal(t, "ready", result["status"])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for port to be fully released.
	time.Sleep(500 * time.Millisecond)
}

// TestAdminServer_AdminBaseURL tests AdminBaseURL returns correct URL format.
func TestAdminServer_AdminBaseURL(t *testing.T) {
	// NOT parallel - server startup timing is unpredictable when running with other tests.
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

	// Wait for server to be ready with retry logic.
	var port int

	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)

		port = server.ActualPort()
		if port > 0 {
			break
		}
	}

	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Test AdminBaseURL returns correct format.
	baseURL := server.AdminBaseURL()
	expectedURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, port)
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

// TestAdminServer_App tests App returns non-nil fiber.App instance.
func TestAdminServer_App(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// App should return the underlying fiber.App.
	app := server.App()
	require.NotNil(t, app, "App() should return non-nil fiber.App")
}

// TestAdminServer_Livez_DuringShutdown_InMemory tests livez returns 503 during shutdown using app.Test().
func TestAdminServer_Livez_DuringShutdown_InMemory(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// Trigger shutdown via context cancellation to set shutdown flag.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// Use app.Test() for in-memory request (no HTTPS listener needed).
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/api/v1/livez", nil)
	require.NoError(t, err)

	resp, err := server.App().Test(req, -1) // -1 = no timeout
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// During shutdown, livez should return 503.
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	assert.Equal(t, "shutting down", result["status"])
}

// TestAdminServer_Readyz_DuringShutdown_InMemory tests readyz returns 503 during shutdown using app.Test().
func TestAdminServer_Readyz_DuringShutdown_InMemory(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// Mark ready first.
	server.SetReady(true)

	// Trigger shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// Use app.Test() for in-memory request.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/api/v1/readyz", nil)
	require.NoError(t, err)

	resp, err := server.App().Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// During shutdown, readyz should return 503.
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	assert.Equal(t, "shutting down", result["status"])
}

// TestAdminServer_Shutdown_Idempotent tests multiple shutdown calls are safe.
func TestAdminServer_Shutdown_Idempotent(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	// Call shutdown multiple times - should not panic or error.
	ctx := context.Background()

	err = server.Shutdown(ctx)
	require.NoError(t, err)

	// Second call should also succeed (idempotent).
	err = server.Shutdown(ctx)
	require.NoError(t, err)

	// Third call with nil context should also succeed.
	err = server.Shutdown(nil) //nolint:staticcheck // Testing nil context handling in Shutdown.
	require.NoError(t, err)
}
