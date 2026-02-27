// Copyright (c) 2025 Justin Cranford

package server_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRPServer_Lifecycle tests admin endpoints (livez, readyz).
func TestRPServer_Lifecycle(t *testing.T) {
	t.Parallel()

	requireTestSetup(t)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)
	defer cancel()

	// Test /admin/api/v1/livez endpoint.
	livezReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/admin/api/v1/livez", testAdminBaseURL), nil)
	require.NoError(t, err, "failed to create livez request")

	livezResp, err := testHTTPClient.Do(livezReq)
	require.NoError(t, err, "failed to send livez request")
	require.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")

	_ = livezResp.Body.Close()

	// Test /admin/api/v1/readyz endpoint.
	readyzReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/admin/api/v1/readyz", testAdminBaseURL), nil)
	require.NoError(t, err, "failed to create readyz request")

	readyzResp, err := testHTTPClient.Do(readyzReq)
	require.NoError(t, err, "failed to send readyz request")
	require.Equal(t, http.StatusOK, readyzResp.StatusCode, "readyz should return 200 OK")

	_ = readyzResp.Body.Close()
}

// TestRPServer_PortAllocation verifies dynamic port allocation.
func TestRPServer_PortAllocation(t *testing.T) {
	t.Parallel()

	requireTestSetup(t)

	// Verify ports were dynamically allocated (not 0).
	publicPort := testServer.PublicPort()
	adminPort := testServer.AdminPort()

	require.Greater(t, publicPort, 0, "public port should be allocated")
	require.Greater(t, adminPort, 0, "admin port should be allocated")
	require.NotEqual(t, publicPort, adminPort, "public and admin ports should be different")
}

// TestRPServer_TemplateServices verifies template services are initialized.
func TestRPServer_TemplateServices(t *testing.T) {
	t.Parallel()

	requireTestSetup(t)

	// Verify template services are available.
	require.NotNil(t, testServer.DB(), "database should be initialized")
	require.NotNil(t, testServer.JWKGen(), "JWK generator should be initialized")
	require.NotNil(t, testServer.Telemetry(), "telemetry should be initialized")
	require.NotNil(t, testServer.Barrier(), "barrier should be initialized")
	require.NotNil(t, testServer.App(), "application should be initialized")
}

// TestRPServer_PublicHealth tests public health endpoints.
func TestRPServer_PublicHealth(t *testing.T) {
	t.Parallel()

	requireTestSetup(t)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)
	defer cancel()

	// Test /health endpoint on public server.
	healthReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/health", testPublicBaseURL), nil)
	require.NoError(t, err, "failed to create health request")

	healthResp, err := testHTTPClient.Do(healthReq)
	require.NoError(t, err, "failed to send health request")
	require.Equal(t, http.StatusOK, healthResp.StatusCode, "health should return 200 OK")

	_ = healthResp.Body.Close()

	// Test /livez endpoint on public server.
	livezReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/livez", testPublicBaseURL), nil)
	require.NoError(t, err, "failed to create livez request")

	livezResp, err := testHTTPClient.Do(livezReq)
	require.NoError(t, err, "failed to send livez request")
	require.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")

	_ = livezResp.Body.Close()

	// Test /readyz endpoint on public server.
	// Note: In test config, AuthzServerURL is set but AuthZ server isn't running.
	// The readyz check will fail since AuthZ is unavailable.
	readyzReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/readyz", testPublicBaseURL), nil)
	require.NoError(t, err, "failed to create readyz request")

	readyzResp, err := testHTTPClient.Do(readyzReq)
	require.NoError(t, err, "failed to send readyz request")
	// Accept either 200 (AuthZ available) or 503 (AuthZ unavailable) - both are valid behaviors.
	require.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, readyzResp.StatusCode,
		"readyz should return 200 OK or 503 Service Unavailable")

	_ = readyzResp.Body.Close()
}

// TestRPServer_AccessorMethods verifies all accessor methods work correctly.
func TestRPServer_AccessorMethods(t *testing.T) {
	t.Parallel()

	requireTestSetup(t)

	// Test PublicBaseURL.
	publicURL := testServer.PublicBaseURL()
	require.NotEmpty(t, publicURL, "public base URL should not be empty")
	require.Contains(t, publicURL, "https://", "public URL should use HTTPS")

	// Test AdminBaseURL.
	adminURL := testServer.AdminBaseURL()
	require.NotEmpty(t, adminURL, "admin base URL should not be empty")
	require.Contains(t, adminURL, "https://", "admin URL should use HTTPS")

	// Test PublicServerActualPort.
	publicActualPort := testServer.PublicServerActualPort()
	require.Equal(t, testServer.PublicPort(), publicActualPort,
		"PublicServerActualPort should match PublicPort")

	// Test AdminServerActualPort.
	adminActualPort := testServer.AdminServerActualPort()
	require.Equal(t, testServer.AdminPort(), adminActualPort,
		"AdminServerActualPort should match AdminPort")
}
