// Copyright (c) 2025 Justin Cranford
//
// Integration tests for JOSE-JA server lifecycle.
package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJoseJAServer_Lifecycle(t *testing.T) {
	// Verify admin endpoints accessible.
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should not be empty")

	// Test /admin/api/v1/livez endpoint.
	livezResp, err := testHTTPClient.Get(fmt.Sprintf("%s/admin/api/v1/livez", testAdminBaseURL))
	require.NoError(t, err, "livez request should succeed")
	require.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")
	require.NoError(t, livezResp.Body.Close())

	// Test /admin/api/v1/readyz endpoint.
	readyzResp, err := testHTTPClient.Get(fmt.Sprintf("%s/admin/api/v1/readyz", testAdminBaseURL))
	require.NoError(t, err, "readyz request should succeed")
	require.Equal(t, http.StatusOK, readyzResp.StatusCode, "readyz should return 200 OK")
	require.NoError(t, readyzResp.Body.Close())

	// Verify public endpoints accessible.
	require.NotEmpty(t, testPublicBaseURL, "public base URL should not be empty")

	// Note: Cannot test actual routes without authentication setup
	// This integration test validates server lifecycle only
}

func TestJoseJAServer_PortAllocation(t *testing.T) {
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
