// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io"
	http "net/http"
	"testing"
	"time"

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

	client := &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment uses self-signed certs.
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"health", "/health", http.StatusOK},
		{"livez", cryptoutilSharedMagic.PrivateAdminLivezRequestPath, http.StatusOK},
		{"readyz", cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := testBaseURL + tt.path

			ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() {
				_ = resp.Body.Close()
			}()

			require.Equal(t, tt.wantCode, resp.StatusCode)

			// Read body to verify response.
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)
		})
	}
}

func TestAuthzServer_OIDCDiscovery(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	client := &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment uses self-signed certs.
			},
		},
	}

	// Test OIDC discovery endpoint.
	url := testBaseURL + cryptoutilSharedMagic.PathDiscovery

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read body to verify JSON response.
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
	require.Contains(t, string(body), "issuer")
}

func TestAuthzServer_JWKS(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	client := &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment uses self-signed certs.
			},
		},
	}

	// Test JWKS endpoint.
	url := testBaseURL + cryptoutilSharedMagic.PathJWKS

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read body to verify JSON response.
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
	require.Contains(t, string(body), "keys")
}

func TestAuthzServer_AccessorMethods(t *testing.T) {
	t.Parallel()
	requireTestSetup(t)

	// Test accessor methods.
	require.NotNil(t, testServer.Config())
	require.NotEmpty(t, testServer.PublicBaseURL())
	require.NotEmpty(t, testServer.AdminBaseURL())
}

// HTTP timeout for test requests.
const httpTimeout = 5 * time.Second
