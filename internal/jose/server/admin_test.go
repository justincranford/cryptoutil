// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"net/http/httptest"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestNewAdminHTTPServer_NilContext tests that nil context returns error.
func TestNewAdminHTTPServer_NilContext(t *testing.T) {
	t.Parallel()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := createTestTLSConfig()

	_, err := NewAdminHTTPServer(nil, settings, tlsCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewAdminHTTPServer_NilSettings tests that nil settings returns error.
func TestNewAdminHTTPServer_NilSettings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tlsCfg := createTestTLSConfig()

	_, err := NewAdminHTTPServer(ctx, nil, tlsCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestNewAdminHTTPServer_NilTLSConfig tests that nil TLS config returns error.
func TestNewAdminHTTPServer_NilTLSConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	_, err := NewAdminHTTPServer(ctx, settings, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TLS configuration cannot be nil")
}

// TestNewAdminHTTPServer_Success tests successful creation.
func TestNewAdminHTTPServer_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := createTestTLSConfig()

	server, err := NewAdminHTTPServer(ctx, settings, tlsCfg)
	require.NoError(t, err)
	require.NotNil(t, server)
	require.NotNil(t, server.app)
	require.False(t, server.ready)
	require.False(t, server.shutdown)
}

// TestAdminServer_HandleLivez_Alive tests livez handler returns alive when server is running.
func TestAdminServer_HandleLivez_Alive(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adminServer := &AdminServer{
		ready:    false,
		shutdown: false,
	}
	app.Get("/admin/api/v1/livez", adminServer.handleLivez)

	req := httptest.NewRequest("GET", "/admin/api/v1/livez", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
}

// TestAdminServer_HandleLivez_ShuttingDown tests livez handler returns 503 during shutdown.
func TestAdminServer_HandleLivez_ShuttingDown(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adminServer := &AdminServer{
		ready:    true,
		shutdown: true,
	}
	app.Get("/admin/api/v1/livez", adminServer.handleLivez)

	req := httptest.NewRequest("GET", "/admin/api/v1/livez", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 503, resp.StatusCode)
}

// TestAdminServer_HandleReadyz_Ready tests readyz handler returns ready status.
func TestAdminServer_HandleReadyz_Ready(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adminServer := &AdminServer{
		ready:    true,
		shutdown: false,
	}
	app.Get("/admin/api/v1/readyz", adminServer.handleReadyz)

	req := httptest.NewRequest("GET", "/admin/api/v1/readyz", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
}

// TestAdminServer_HandleReadyz_NotReady tests readyz handler returns 503 when not ready.
func TestAdminServer_HandleReadyz_NotReady(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adminServer := &AdminServer{
		ready:    false,
		shutdown: false,
	}
	app.Get("/admin/api/v1/readyz", adminServer.handleReadyz)

	req := httptest.NewRequest("GET", "/admin/api/v1/readyz", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 503, resp.StatusCode)
}

// TestAdminServer_HandleReadyz_ShuttingDown tests readyz handler returns 503 during shutdown.
func TestAdminServer_HandleReadyz_ShuttingDown(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adminServer := &AdminServer{
		ready:    true,
		shutdown: true,
	}
	app.Get("/admin/api/v1/readyz", adminServer.handleReadyz)

	req := httptest.NewRequest("GET", "/admin/api/v1/readyz", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 503, resp.StatusCode)
}

// TestAdminServer_HandleShutdown tests shutdown handler initiates shutdown.
func TestAdminServer_HandleShutdown(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adminServer := &AdminServer{
		ready:    true,
		shutdown: false,
	}
	app.Post("/admin/api/v1/shutdown", adminServer.handleShutdown)

	req := httptest.NewRequest("POST", "/admin/api/v1/shutdown", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)

	adminServer.mu.RLock()
	isShutdown := adminServer.shutdown
	adminServer.mu.RUnlock()
	require.True(t, isShutdown)
}

// TestAdminServer_Start_NilContext tests Start with nil context returns error.
func TestAdminServer_Start_NilContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := createTestTLSConfig()

	server, err := NewAdminHTTPServer(ctx, settings, tlsCfg)
	require.NoError(t, err)

	err = server.Start(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestAdminServer_Shutdown_NilContext tests Shutdown with nil context handles gracefully.
func TestAdminServer_Shutdown_NilContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := createTestTLSConfig()

	server, err := NewAdminHTTPServer(ctx, settings, tlsCfg)
	require.NoError(t, err)

	// Shutdown with nil context should not return error (uses context.Background()).
	err = server.Shutdown(nil)
	require.NoError(t, err)
}
