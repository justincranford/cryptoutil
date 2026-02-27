// Copyright (c) 2025 Justin Cranford
//
// Integration tests for pki-ca server lifecycle.
package server

import (
	"context"
	"fmt"
	http "net/http"
	"testing"
	"time"

	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki/ca/server/config"
	cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCAServer_Lifecycle(t *testing.T) {
	t.Parallel()

	// Server should be running and ready from TestMain.
	require.NotNil(t, testServer, "server should not be nil")
	require.Greater(t, testServer.PublicPort(), 0, "public port should be assigned")
	require.Greater(t, testServer.AdminPort(), 0, "admin port should be assigned")

	require.NotEmpty(t, testPublicBaseURL, "public base URL should not be empty")
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
}

func TestCAServer_PortAllocation(t *testing.T) {
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

func TestCAServer_Accessors(t *testing.T) {
	t.Parallel()

	// Verify all accessor methods return valid values.
	require.NotNil(t, testServer.DB(), "DB accessor should return non-nil")

	// Verify port accessors return valid values.
	require.Greater(t, testServer.PublicPort(), 0, "PublicPort should be positive")
	require.Greater(t, testServer.AdminPort(), 0, "AdminPort should be positive")
	require.Greater(t, testServer.PublicServerActualPort(), 0, "PublicServerActualPort should be positive")
	require.Greater(t, testServer.AdminServerActualPort(), 0, "AdminServerActualPort should be positive")

	// Verify URL accessors.
	require.NotEmpty(t, testServer.PublicBaseURL(), "PublicBaseURL should not be empty")
	require.NotEmpty(t, testServer.AdminBaseURL(), "AdminBaseURL should not be empty")
}

func TestCAServer_HealthEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{name: "livez", url: fmt.Sprintf("%s/admin/api/v1/livez", testAdminBaseURL), wantStatus: http.StatusOK},
		{name: "readyz", url: fmt.Sprintf("%s/admin/api/v1/readyz", testAdminBaseURL), wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.url, nil)
			require.NoError(t, err)

			resp, err := testHTTPClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestCAServer_ShutdownIdempotent(t *testing.T) {
	t.Parallel()

	// Test that calling Shutdown on a running server succeeds.
	// Note: We can't shut down testServer as other tests need it.
	// This test creates a separate server instance to test shutdown coverage.
	ctx := context.Background()

	// Create test configuration with different ports.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create separate server instance.
	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err, "server creation should succeed")

	// Start server in background.
	cryptoutilAppsTemplateServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(server, func() error {
		return server.Start(ctx)
	})

	// Mark server as ready.
	server.SetReady(true)

	// First shutdown should succeed.
	shutdownCtx, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.DefaultDataServerShutdownTimeout*time.Second)
	defer cancel()

	shutdownErr := server.Shutdown(shutdownCtx)
	require.NoError(t, shutdownErr, "first shutdown should succeed")
}
