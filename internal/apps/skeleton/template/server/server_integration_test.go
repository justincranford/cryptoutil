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
cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/stretchr/testify/require"
)

func TestSkeletonTemplateServer_Lifecycle(t *testing.T) {
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

// Test that calling Shutdown on a running server succeeds, and Start returns cleanly.
// Note: We can't shut down testServer as other tests need it.
// This test creates a separate server instance to test shutdown and start-return coverage.
ctx, cancel := context.WithCancel(context.Background())

// Create test configuration with different ports.
cfg := cryptoutilAppsSkeletonTemplateServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

// Create separate server instance.
server, err := NewFromConfig(ctx, cfg)
require.NoError(t, err, "server creation should succeed")

// Start server in background, capture Start() return value.
startErrCh := make(chan error, 1)

cryptoutilAppsTemplateServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(server, func() error {
startErr := server.Start(ctx)
startErrCh <- startErr

return startErr
})

// Mark server as ready.
server.SetReady(true)

// Shutdown the server explicitly (covers Shutdown happy path).
shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
defer shutdownCancel()

err = server.Shutdown(shutdownCtx)
require.NoError(t, err, "shutdown should succeed")

// Cancel context to trigger app.Start() to return (it selects on ctx.Done()).
// Without this, server.Start() would block forever since Shutdown() only stops
// the Fiber servers but doesn't fire ctx.Done() or errChan.
cancel()

// Wait for Start to return after shutdown (covers Start return-nil path).
select {
case startErr := <-startErrCh:
// Start may return nil or context error after shutdown - both are acceptable.
if startErr != nil {
t.Logf("Start() returned error after shutdown (acceptable): %v", startErr)
}
case <-time.After(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second):
t.Fatal("Start() did not return after shutdown within timeout")
}
}
