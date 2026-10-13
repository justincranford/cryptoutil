// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestAuthzServer_Lifecycle(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Server should be running and ready from TestMain.
	require.NotNil(t, testServer)
	require.Greater(t, testServer.PublicPort(), 0)
	require.Greater(t, testServer.AdminPort(), 0)
}

func TestAuthzServer_PortAllocation(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Dynamic ports should be allocated (not hardcoded).
	publicPort := testServer.PublicPort()
	adminPort := testServer.AdminPort()

	require.Greater(t, publicPort, 0, "Public port should be allocated")
	require.Greater(t, adminPort, 0, "Admin port should be allocated")
	require.NotEqual(t, publicPort, adminPort, "Ports should be different")
}

func TestAuthzServer_TemplateServices(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Verify template services are initialized.
	require.NotNil(t, testServer.DB(), "Database should be initialized")
	require.NotNil(t, testServer.JWKGen(), "JWK generation service should be initialized")
	require.NotNil(t, testServer.Telemetry(), "Telemetry service should be initialized")
	require.NotNil(t, testServer.Barrier(), "Barrier service should be initialized")
	require.NotNil(t, testServer.App(), "Application should be initialized")
}

func TestAuthzServer_PublicHealth(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	tests := []struct {
		name     string
		baseURL  string
		path     string
		admin    bool
		wantCode int
	}{
		{"health", testServer.PublicBaseURL(), "/health", false, http.StatusOK},
		{"livez", testServer.AdminBaseURL(), cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath, true, http.StatusOK},
		{"readyz", testServer.AdminBaseURL(), cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rootCAs := testServer.TLSRootCAPool()
			if tt.admin {
				rootCAs = testServer.AdminTLSRootCAPool()
			}

			client := &http.Client{
				Timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						MinVersion: tls.VersionTLS13,
						RootCAs:    rootCAs,
					},
					DisableKeepAlives: true,
				},
			}

			url := tt.baseURL + tt.path

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultHealthCheckTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)

			require.Equal(t, tt.wantCode, resp.StatusCode)

			// Read body to verify response.
			body, readErr := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()

			require.NoError(t, readErr)
			require.NoError(t, closeErr)
			require.NotEmpty(t, body)
		})
	}
}

func TestAuthzServer_WellKnownEndpoints(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantBody     string
		untrustedTLS bool
	}{
		{name: "OIDC discovery", path: cryptoutilSharedMagic.PathDiscovery, wantStatus: http.StatusOK, wantBody: "issuer"},
		{name: "JWKS", path: cryptoutilSharedMagic.PathJWKS, wantStatus: http.StatusOK, wantBody: "keys"},
		{name: "admin livez is not exposed publicly", path: cryptoutilSharedMagic.DefaultAdminLivezEndpoint, wantStatus: http.StatusNotFound},
		{name: "untrusted server certificate is rejected", path: cryptoutilSharedMagic.PathJWKS, untrustedTLS: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootCAs := testServer.TLSRootCAPool()
			if tc.untrustedTLS {
				rootCAs = x509.NewCertPool()
			}

			client := &http.Client{
				Timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						MinVersion: tls.VersionTLS13,
						RootCAs:    rootCAs,
					},
					DisableKeepAlives: true,
				},
			}
			url := testBaseURL + tc.path

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultHealthCheckTimeout)
			defer cancel()

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			require.NoError(t, err)

			response, err := client.Do(request)
			if tc.untrustedTLS {
				require.ErrorContains(t, err, cryptoutilSharedMagic.TLSName)

				return
			}

			require.NoError(t, err)

			body, err := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			require.Equal(t, tc.wantStatus, response.StatusCode)
			require.NoError(t, err)
			require.NoError(t, closeErr)

			if tc.wantBody != "" {
				require.Contains(t, string(body), tc.wantBody)
			}
		})
	}
}

func TestAuthzServer_AccessorMethods(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Test accessor methods.
	require.NotNil(t, testServer.Config())
	require.NotEmpty(t, testServer.PublicBaseURL())
	require.NotEmpty(t, testServer.AdminBaseURL())
}
