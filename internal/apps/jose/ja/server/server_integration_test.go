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
	cryptoutilContract "cryptoutil/internal/apps/template/service/testing/contract"
	cryptoutilTestingTestserver "cryptoutil/internal/apps/template/service/testing/testserver"
)

func TestJoseJAServer_Lifecycle(t *testing.T) {
	t.Parallel()
	// Verify admin endpoints accessible.
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should not be empty")

	// Test /admin/api/v1/livez endpoint.
	livezResp, err := testHealthClient.Livez()
	require.NoError(t, err, "livez request should succeed")
	require.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")
	require.NoError(t, livezResp.Body.Close())

	// Test /admin/api/v1/readyz endpoint.
	readyzResp, err := testHealthClient.Readyz()
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure start goroutine exits when test returns

	// Create test configuration with different ports.
	cfg := cryptoutilAppsJoseJaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create separate server instance.
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err, "server creation should succeed")

	// Start server and wait for both ports using shared helper.
	// t.Cleanup will call server.Shutdown when test completes.
	cryptoutilTestingTestserver.StartAndWait(ctx, t, server)

	// Shutdown the server explicitly (covers Shutdown happy path).
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(),
		cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err, "shutdown should succeed")
}

func TestJoseJAServer_ContractCompliance(t *testing.T) {
	t.Parallel()
	cryptoutilContract.RunContractTests(t, testServer)
}
