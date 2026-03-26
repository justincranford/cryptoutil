// Copyright (c) 2025 Justin Cranford
//
// Integration tests for pki-ca server lifecycle.
package server

import (
	"context"
	"fmt"
	http "net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCAServer_Lifecycle(t *testing.T) {
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

func TestCAServer_CAServices(t *testing.T) {
	t.Parallel()
	// Verify CA-specific services are initialized.
	require.NotNil(t, testServer.Issuer(), "issuer service should be initialized")
	require.NotNil(t, testServer.Storage(), "storage service should be initialized")
	require.NotNil(t, testServer.CRLService(), "CRL service should be initialized")
	require.NotNil(t, testServer.OCSPService(), "OCSP service should be initialized")
}

func TestCAServer_TemplateServices(t *testing.T) {
	t.Parallel()
	// Verify template services are initialized.
	require.NotNil(t, testServer.DB(), "database should be initialized")
	require.NotNil(t, testServer.JWKGen(), "JWK generation service should be initialized")
	require.NotNil(t, testServer.Telemetry(), "telemetry service should be initialized")
	require.NotNil(t, testServer.Barrier(), "barrier service should be initialized")
}

func TestCAServer_PublicHealth(t *testing.T) {
	t.Parallel()
	// Test /service/api/v1/health endpoint on public server (provided by template).
	// Note: /admin/api/v1/livez and /admin/api/v1/readyz are tested in TestCAServer_Lifecycle.
	healthReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/service/api/v1/health", testPublicBaseURL), nil)
	require.NoError(t, err, "health request creation should succeed")

	healthResp, err := testHTTPClient.Do(healthReq)
	require.NoError(t, err, "health request should succeed")
	require.Equal(t, http.StatusOK, healthResp.StatusCode, "health should return 200 OK")
	require.NoError(t, healthResp.Body.Close())
}

func TestCAServer_CRLEndpoint(t *testing.T) {
	t.Parallel()
	// Test CRL distribution endpoint.
	crlReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/service/api/v1/crl", testPublicBaseURL), nil)
	require.NoError(t, err, "CRL request creation should succeed")

	crlResp, err := testHTTPClient.Do(crlReq)
	require.NoError(t, err, "CRL request should succeed")
	require.Equal(t, http.StatusOK, crlResp.StatusCode, "CRL endpoint should return 200 OK")
	require.Equal(t, "application/pkix-crl", crlResp.Header.Get("Content-Type"), "CRL should have correct content type")
	require.NoError(t, crlResp.Body.Close())

	// Test well-known CRL endpoint.
	wellKnownReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/.well-known/pki-ca/crl", testPublicBaseURL), nil)
	require.NoError(t, err, "well-known CRL request creation should succeed")

	wellKnownResp, err := testHTTPClient.Do(wellKnownReq)
	require.NoError(t, err, "well-known CRL request should succeed")
	require.Equal(t, http.StatusOK, wellKnownResp.StatusCode, "well-known CRL endpoint should return 200 OK")
	require.NoError(t, wellKnownResp.Body.Close())
}
