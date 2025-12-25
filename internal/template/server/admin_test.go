// Copyright (c) 2025 Justin Cranford
//
//

package server_test

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

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTemplateServer "cryptoutil/internal/template/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAdminServer_HappyPath tests successful admin server creation.
func TestNewAdminServer_HappyPath(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)

	require.NoError(t, err)
	require.NotNil(t, server)
}

// TestNewAdminServer_NilContext tests that NewAdminServer rejects nil context.
func TestNewAdminServer_NilContext(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilTemplateServer.NewAdminServer(nil, 0) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
	assert.Nil(t, server)
}

// TestAdminServer_Start_Success tests admin server starts and listens on dynamic port.
func TestAdminServer_Start_Success(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
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

	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)

		port = server.ActualPort()
		if port > 0 {
			break
		}
	}

	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Verify no startup error.
	select {
	case err := <-startErr:
		t.Fatalf("unexpected start error: %v", err)
	default:
	}

	// Wait for port to be fully released before next test.
	time.Sleep(500 * time.Millisecond)
}

// TestAdminServer_Start_NilContext tests that Start rejects nil context.
func TestAdminServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
	require.NoError(t, err)

	err = server.Start(nil) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
}

// TestAdminServer_Livez_Alive tests /admin/v1/livez endpoint when server is alive.
func TestAdminServer_Livez_Alive(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
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
		Timeout: 5 * time.Second,
	}

	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/v1/livez", cryptoutilMagic.IPv4Loopback, port)

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

	assert.Equal(t, "alive", response["status"])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Readyz_Ready tests /admin/v1/readyz endpoint when server is ready.
func TestAdminServer_Readyz_Ready(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
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

	// Query readyz endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
		},
		Timeout: 5 * time.Second,
	}

	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/v1/readyz", cryptoutilMagic.IPv4Loopback, port)

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

	assert.Equal(t, "ready", response["status"])

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Shutdown_Endpoint tests POST /admin/v1/shutdown triggers graceful shutdown.
func TestAdminServer_Shutdown_Endpoint(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background and track goroutine.
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

	// Trigger shutdown via HTTP endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
		},
		Timeout: 5 * time.Second,
	}

	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/v1/shutdown", cryptoutilMagic.IPv4Loopback, port)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

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

	assert.Equal(t, "shutdown initiated", response["status"])

	// The endpoint triggers shutdown in a goroutine with 100ms delay.
	// Cancel context to let Start() exit cleanly, then wait for goroutine.
	cancel()
	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Shutdown_NilContext tests Shutdown accepts nil context and uses Background().
func TestAdminServer_Shutdown_NilContext(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
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

	// Shutdown with nil context (should use Background()).
	err = server.Shutdown(nil) //nolint:staticcheck // Testing nil context handling.
	require.NoError(t, err)

	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release - ConcurrentRequests runs next.
	time.Sleep(2 * time.Second)
}

// TestAdminServer_ActualPort_BeforeStart tests ActualPort before server starts.
func TestAdminServer_ActualPort_BeforeStart(t *testing.T) {
	t.Parallel()

	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
	require.NoError(t, err)

	port := server.ActualPort()

	assert.Equal(t, 0, port, "Expected port 0 before server starts")
}

// TestAdminServer_ConcurrentRequests tests multiple concurrent requests to admin endpoints.
func TestAdminServer_ConcurrentRequests(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.

	// Wait for port to be fully released from previous test.
	time.Sleep(2 * time.Second)

	server, err := cryptoutilTemplateServer.NewAdminServer(context.Background(), 0)
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
			t.Fatalf("server.Start() error after %d attempts: %v", i, startErr)
		}

		port = server.ActualPort()
		if port > 0 {
			healthCtx, healthCancel := context.WithTimeout(context.Background(), 2*time.Second)

			healthURL := fmt.Sprintf("https://%s:%d/admin/v1/livez", cryptoutilMagic.IPv4Loopback, port)

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
		t.Fatalf("server.Start() error: %v", startErr)
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

			url := fmt.Sprintf("https://%s:%d/admin/v1/livez", cryptoutilMagic.IPv4Loopback, port)

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
		t.Errorf("concurrent request error: %v", err)
	}

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)

	wg.Wait()

	// Check if server.Start() returned an error.
	if startErr != nil {
		t.Errorf("server.Start() error: %v", startErr)
	}

	// Wait for port to be fully released before next test.
	time.Sleep(500 * time.Millisecond)
}
