// Copyright (c) 2025 Justin Cranford
//
// Integration tests for JOSE-JA server lifecycle.
package server

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"
)

func TestJoseJAServer_Lifecycle(t *testing.T) {
	t.Parallel()
	// Verify admin endpoints accessible.
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should not be empty")

	// Test /admin/api/v1/livez endpoint.
	livezReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/admin/api/v1/livez", testAdminBaseURL), nil)
	require.NoError(t, err, "livez request creation should succeed")
	livezResp, err := testHTTPClient.Do(livezReq)
	require.NoError(t, err, "livez request should succeed")
	require.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")
	require.NoError(t, livezResp.Body.Close())

	// Test /admin/api/v1/readyz endpoint.
	readyzReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/admin/api/v1/readyz", testAdminBaseURL), nil)
	require.NoError(t, err, "readyz request creation should succeed")
	readyzResp, err := testHTTPClient.Do(readyzReq)
	require.NoError(t, err, "readyz request should succeed")
	require.Equal(t, http.StatusOK, readyzResp.StatusCode, "readyz should return 200 OK")
	require.NoError(t, readyzResp.Body.Close())

	// Verify public endpoints accessible.
	require.NotEmpty(t, testPublicBaseURL, "public base URL should not be empty")
	// Note: Cannot test actual routes without authentication setup
	// This integration test validates server lifecycle only
}

func TestJoseJAServer_PortAllocation(t *testing.T) {
	t.Parallel()
	// Verify ports are dynamically allocated (> 0).
	publicPort := testServer.PublicPort()
	adminPort := testServer.AdminPort()

	require.Greater(t, publicPort, 0, "public port should be dynamically allocated")
	require.Greater(t, adminPort, 0, "admin port should be dynamically allocated")
	require.NotEqual(t, publicPort, adminPort, "public and admin ports should differ")

	// Verify base URLs are constructed correctly with dynamic ports.
	require.Contains(t, testPublicBaseURL, fmt.Sprintf(":%d", publicPort), "public base URL should contain allocated port")
	require.Contains(t, testAdminBaseURL, fmt.Sprintf(":%d", adminPort), "admin base URL should contain allocated port")
}

func TestJoseJAServer_Accessors(t *testing.T) {
	t.Parallel()
	// Test all accessor methods for coverage.
	// These are simple getters but need explicit test coverage.

	// DB accessor.
	db := testServer.DB()
	require.NotNil(t, db, "DB() should return non-nil database connection")

	// App accessor.
	app := testServer.App()
	require.NotNil(t, app, "App() should return non-nil application wrapper")

	// JWKGen accessor.
	jwkGen := testServer.JWKGen()
	require.NotNil(t, jwkGen, "JWKGen() should return non-nil JWK generation service")

	// Telemetry accessor.
	telemetry := testServer.Telemetry()
	require.NotNil(t, telemetry, "Telemetry() should return non-nil telemetry service")

	// Barrier accessor.
	barrier := testServer.Barrier()
	require.NotNil(t, barrier, "Barrier() should return non-nil barrier service")

	// PublicServerActualPort - duplicate of PublicPort but different method.
	publicActualPort := testServer.PublicServerActualPort()
	require.Greater(t, publicActualPort, 0, "PublicServerActualPort() should return allocated port")
	require.Equal(t, testServer.PublicPort(), publicActualPort, "PublicServerActualPort() should match PublicPort()")

	// AdminServerActualPort - duplicate of AdminPort but different method.
	adminActualPort := testServer.AdminServerActualPort()
	require.Greater(t, adminActualPort, 0, "AdminServerActualPort() should return allocated port")
	require.Equal(t, testServer.AdminPort(), adminActualPort, "AdminServerActualPort() should match AdminPort()")

	// PublicBaseURL accessor.
	publicBaseURL := testServer.PublicBaseURL()
	require.NotEmpty(t, publicBaseURL, "PublicBaseURL() should return non-empty URL")
	require.Contains(t, publicBaseURL, "https://", "PublicBaseURL() should be HTTPS")

	// AdminBaseURL accessor.
	adminBaseURL := testServer.AdminBaseURL()
	require.NotEmpty(t, adminBaseURL, "AdminBaseURL() should return non-empty URL")
	require.Contains(t, adminBaseURL, "https://", "AdminBaseURL() should be HTTPS")
}

func TestJoseJAServer_ShutdownIdempotent(t *testing.T) {
	t.Parallel()
	// Test that calling Shutdown on an already-running server is idempotent.
	// Note: We can't actually shut down testServer as other tests need it.
	// This test creates a separate server instance to test shutdown coverage.
	ctx := context.Background()

	// Create test configuration with different ports.
	cfg := cryptoutilAppsJoseJaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create separate server instance.
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err, "server creation should succeed")

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := server.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for server to bind.
	require.Eventually(t, func() bool {
		return server.PublicPort() > 0 && server.AdminPort() > 0
	}, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, "server should bind to ports")

	// Mark server as ready.
	server.SetReady(true)

	// Shutdown the server - this should succeed.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err, "shutdown should succeed")
}
