// Copyright (c) 2025 Justin Cranford
//
// Integration tests for skeleton-template server lifecycle.
package server

import (
	"context"
	"fmt"
	http "net/http"
	"testing"
	"time"

	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton/template/server/config"
	cryptoutilTestingTestserver "cryptoutil/internal/apps/template/service/testing/testserver"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestSkeletonTemplateServer_Lifecycle(t *testing.T) {
	t.Parallel()

	// Verify admin endpoints accessible.
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should not be empty")

	// Test /admin/api/v1/livez endpoint using shared health client.
	livezResp, err := testHealthClient.Livez()
	require.NoError(t, err, "livez request should succeed")
	require.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")
	require.NoError(t, livezResp.Body.Close())

	// Test /admin/api/v1/readyz endpoint using shared health client.
	readyzResp, err := testHealthClient.Readyz()
	require.NoError(t, err, "readyz request should succeed")
	require.Equal(t, http.StatusOK, readyzResp.StatusCode, "readyz should return 200 OK")
	require.NoError(t, readyzResp.Body.Close())

	// Verify public endpoints accessible.
	require.NotEmpty(t, testPublicBaseURL, "public base URL should not be empty")
}

func TestSkeletonTemplateServer_PortAllocation(t *testing.T) {
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

func TestSkeletonTemplateServer_Accessors(t *testing.T) {
	t.Parallel()

	// Verify all accessor methods return non-nil values.
	require.NotNil(t, testServer.DB(), "DB accessor should return non-nil")
	require.NotNil(t, testServer.App(), "App accessor should return non-nil")
	require.NotNil(t, testServer.JWKGen(), "JWKGen accessor should return non-nil")
	require.NotNil(t, testServer.Telemetry(), "Telemetry accessor should return non-nil")
	require.NotNil(t, testServer.Barrier(), "Barrier accessor should return non-nil")

	// Verify port accessors return valid values.
	require.Greater(t, testServer.PublicPort(), 0, "PublicPort should be positive")
	require.Greater(t, testServer.AdminPort(), 0, "AdminPort should be positive")
	require.Greater(t, testServer.PublicServerActualPort(), 0, "PublicServerActualPort should be positive")
	require.Greater(t, testServer.AdminServerActualPort(), 0, "AdminServerActualPort should be positive")

	// Verify URL accessors.
	require.NotEmpty(t, testServer.PublicBaseURL(), "PublicBaseURL should not be empty")
	require.NotEmpty(t, testServer.AdminBaseURL(), "AdminBaseURL should not be empty")
}

func TestSkeletonTemplateServer_HealthEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{name: "public browser health", url: fmt.Sprintf("%s/browser/api/v1/health", testPublicBaseURL), wantStatus: http.StatusOK},
		{name: "public service health", url: fmt.Sprintf("%s/service/api/v1/health", testPublicBaseURL), wantStatus: http.StatusOK},
		{name: "admin livez", url: fmt.Sprintf("%s/admin/api/v1/livez", testAdminBaseURL), wantStatus: http.StatusOK},
		{name: "admin readyz", url: fmt.Sprintf("%s/admin/api/v1/readyz", testAdminBaseURL), wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.url, nil)
			require.NoError(t, reqErr)

			resp, respErr := testHTTPClient.Do(req)
			require.NoError(t, respErr)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tc.wantStatus, resp.StatusCode)
		})
	}
}

func TestSkeletonTemplateServer_ShutdownIdempotent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure start goroutine exits when test returns

	// Create test configuration with different ports.
	cfg := cryptoutilAppsSkeletonTemplateServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

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