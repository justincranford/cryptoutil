// Copyright (c) 2025 Justin Cranford
//

//go:build !integration

package server

import (
	"context"
	"crypto/tls"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Helper function to create test TLS material.
func createTestTLSMaterial(t *testing.T) *cryptoutilAppsTemplateServiceConfig.TLSMaterial {
	t.Helper()

	// Generate TLS settings with auto-generated CA hierarchy.
	tlsSettings, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost"},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)
	require.NotNil(t, tlsSettings)

	// Convert TLSGeneratedSettings (PEM bytes) to TLSMaterial (parsed tls.Config).
	tlsMaterial, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateTLSMaterial(tlsSettings)
	require.NoError(t, err)
	require.NotNil(t, tlsMaterial)

	return tlsMaterial
}

// TestNewPublicServerBase_NilConfig tests constructor with nil config.
func TestNewPublicServerBase_NilConfig(t *testing.T) {
	t.Parallel()

	server, err := NewPublicServerBase(nil)

	require.Error(t, err)
	require.Nil(t, server)
	require.Contains(t, err.Error(), "configuration cannot be nil")
}

// TestNewPublicServerBase_EmptyBindAddress tests constructor with empty bind address.
func TestNewPublicServerBase_EmptyBindAddress(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: "",
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)

	require.Error(t, err)
	require.Nil(t, server)
	require.Contains(t, err.Error(), "bind address cannot be empty")
}

// TestNewPublicServerBase_NilTLSMaterial tests constructor with nil TLS material.
func TestNewPublicServerBase_NilTLSMaterial(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: nil,
	}

	server, err := NewPublicServerBase(cfg)

	require.Error(t, err)
	require.Nil(t, server)
	require.Contains(t, err.Error(), "TLS material cannot be nil")
}

// TestNewPublicServerBase_Success tests successful construction.
func TestNewPublicServerBase_Success(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)

	require.NoError(t, err)
	require.NotNil(t, server)
	require.NotNil(t, server.App())
}

// TestPublicServerBase_HandleServiceHealth_Healthy tests healthy status.
func TestPublicServerBase_HandleServiceHealth_Healthy(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/service/api/v1/health", nil)

	resp, err := server.App().Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]string

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "healthy", result["status"])
}

// TestPublicServerBase_HandleServiceHealth_ShuttingDown tests shutting down status.
func TestPublicServerBase_HandleServiceHealth_ShuttingDown(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	// Mark as shutting down.
	server.mu.Lock()
	server.shutdown = true
	server.mu.Unlock()

	req := httptest.NewRequest("GET", "/service/api/v1/health", nil)

	resp, err := server.App().Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	var result map[string]string

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "shutting down", result["status"])
}

// TestPublicServerBase_HandleBrowserHealth_Healthy tests browser health healthy status.
func TestPublicServerBase_HandleBrowserHealth_Healthy(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/browser/api/v1/health", nil)

	resp, err := server.App().Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]string

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "healthy", result["status"])
}

// TestPublicServerBase_HandleBrowserHealth_ShuttingDown tests browser health shutting down status.
func TestPublicServerBase_HandleBrowserHealth_ShuttingDown(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	// Mark as shutting down.
	server.mu.Lock()
	server.shutdown = true
	server.mu.Unlock()

	req := httptest.NewRequest("GET", "/browser/api/v1/health", nil)

	resp, err := server.App().Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	var result map[string]string

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "shutting down", result["status"])
}

// TestPublicServerBase_Shutdown_NotStarted tests shutdown when not started.
func TestPublicServerBase_Shutdown_NotStarted(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	err = server.Shutdown(context.Background())

	require.NoError(t, err)
}

// TestPublicServerBase_Shutdown_AlreadyShutdown tests double shutdown.
func TestPublicServerBase_Shutdown_AlreadyShutdown(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	// First shutdown succeeds.
	err = server.Shutdown(context.Background())
	require.NoError(t, err)

	// Second shutdown fails.
	err = server.Shutdown(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "already shutdown")
}

// TestPublicServerBase_ActualPort_BeforeStart tests ActualPort before start.
func TestPublicServerBase_ActualPort_BeforeStart(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	// Before start, actual port is 0.
	require.Equal(t, 0, server.ActualPort())
}

// TestPublicServerBase_PublicBaseURL_BeforeStart tests PublicBaseURL before start.
func TestPublicServerBase_PublicBaseURL_BeforeStart(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	// Before start, URL uses port 0.
	require.Equal(t, "https://127.0.0.1:0", server.PublicBaseURL())
}

// TestPublicServerBase_App tests App accessor.
func TestPublicServerBase_App(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	app := server.App()
	require.NotNil(t, app)
}

// TestPublicServerBase_Start_NilContext tests Start with nil context.
func TestPublicServerBase_Start_NilContext(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	err = server.Start(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestPublicServerBase_StartAndShutdown tests full lifecycle.
func TestPublicServerBase_StartAndShutdown(t *testing.T) {
	t.Parallel()

	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		errChan <- server.Start(ctx)
	}()

	// Give server time to start.
	time.Sleep(100 * time.Millisecond)

	// Verify actual port is assigned.
	actualPort := server.ActualPort()
	require.Greater(t, actualPort, 0)

	// Verify base URL is correct.
	baseURL := server.PublicBaseURL()
	require.Contains(t, baseURL, "https://127.0.0.1:")

	// Shutdown via context cancellation.
	cancel()

	// Wait for server to stop.
	select {
	case err := <-errChan:
		// Expected: context cancelled error.
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop within timeout")
	}
}

// TestPublicServerBase_StartAndMakeRequest tests making a request to running server.
func TestPublicServerBase_StartAndMakeRequest(t *testing.T) {
	t.Parallel()

	tlsMaterial := createTestTLSMaterial(t)
	cfg := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: tlsMaterial,
	}

	server, err := NewPublicServerBase(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		errChan <- server.Start(ctx)
	}()

	// Wait for server to be ready using polling pattern.
	waitCtx, waitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer waitCancel()

	require.Eventually(t, func() bool {
		return server.ActualPort() > 0
	}, 10*time.Second, 100*time.Millisecond, "server should allocate port")

	// Make actual HTTPS request to running server.
	actualPort := server.ActualPort()
	require.Greater(t, actualPort, 0)

	// Create client that trusts our CA.
	client := createTestHTTPClient(t, tlsMaterial)

	healthURL := server.PublicBaseURL() + "/service/api/v1/health"

	// Use http.NewRequestWithContext for proper context handling.
	req, err := http.NewRequestWithContext(waitCtx, http.MethodGet, healthURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	// Cleanup.
	cancel()

	select {
	case <-errChan:
		// Server stopped.
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop within timeout")
	}
}

// createTestHTTPClient creates an HTTP client that trusts the test CA.
func createTestHTTPClient(t *testing.T, tlsMaterial *cryptoutilAppsTemplateServiceConfig.TLSMaterial) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    tlsMaterial.RootCAPool,
				MinVersion: tls.VersionTLS12,
			},
		},
		Timeout: 5 * time.Second,
	}
}
