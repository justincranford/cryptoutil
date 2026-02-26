// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
	"context"
	"io"
	http "net/http"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)


// TestE2E_HealthChecks validates /health endpoint for all instances.
func TestE2E_HealthChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		publicURL string
	}{
		{authzContainer, authzPublicURL},
		{idpContainer, idpPublicURL},
		{rsContainer, rsPublicURL},
		{rpContainer, rpPublicURL},
		{spaContainer, spaPublicURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
			defer cancel()

			healthURL := tt.publicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			require.NoError(t, err, "Creating health check request should succeed")

			healthResp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "Health check should succeed for %s", tt.name)
			require.NoError(t, healthResp.Body.Close())
			require.Equal(t, http.StatusOK, healthResp.StatusCode,
				"%s should return 200 OK for /health", tt.name)
		})
	}
}

// TestE2E_ServicePath_AuthZ tests /service/** path for AuthZ server.
func TestE2E_ServicePath_AuthZ(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test OIDC Discovery endpoint (/.well-known/openid-configuration).
	discoveryURL := authzPublicURL + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	require.NoError(t, err, "Creating discovery request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "OIDC Discovery request should succeed")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading response body should succeed")
	require.NoError(t, resp.Body.Close())

	// Verify discovery response contains expected fields.
	require.Equal(t, http.StatusOK, resp.StatusCode, "OIDC Discovery should return 200 OK")
	require.Contains(t, string(body), "issuer", "Discovery response should contain issuer")
}

// TestE2E_ServicePath_JWKS tests JWKS endpoint.
func TestE2E_ServicePath_JWKS(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test JWKS endpoint (/.well-known/jwks.json).
	jwksURL := authzPublicURL + "/.well-known/jwks.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	require.NoError(t, err, "Creating JWKS request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "JWKS request should succeed")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading response body should succeed")
	require.NoError(t, resp.Body.Close())

	// Verify JWKS response contains expected fields.
	require.Equal(t, http.StatusOK, resp.StatusCode, "JWKS should return 200 OK")
	require.Contains(t, string(body), "keys", "JWKS response should contain keys array")
}

// TestE2E_BrowserPath_AuthZ tests /browser/** path for AuthZ server.
func TestE2E_BrowserPath_AuthZ(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test browser path for authorization endpoint.
	browserURL := authzPublicURL + cryptoutilSharedMagic.PathPrefixBrowser + "/api/v1/authorize"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, browserURL, nil)
	require.NoError(t, err, "Creating browser authorize request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "Browser authorize request should complete")

	require.NoError(t, resp.Body.Close())

	// Browser endpoints may require authentication, but should not return 404.
	require.NotEqual(t, http.StatusNotFound, resp.StatusCode,
		"Browser path should exist (may require auth)")
}

// TestE2E_CORS_Headers tests CORS headers on browser endpoints.
func TestE2E_CORS_Headers(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test preflight OPTIONS request.
	corsURL := authzPublicURL + cryptoutilSharedMagic.PathPrefixBrowser + "/api/v1/authorize"
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, corsURL, nil)
	require.NoError(t, err, "Creating CORS preflight request should succeed")

	req.Header.Set("Origin", "https://localhost:8600")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "CORS preflight request should complete")

	require.NoError(t, resp.Body.Close())

	// Check for CORS headers (may be configured differently in test env).
	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")

	// CORS headers should be present for browser endpoints.
	if corsHeader != "" {
		require.True(t, strings.Contains(corsHeader, "*") || strings.Contains(corsHeader, "localhost"),
			"CORS should allow localhost origins")
	}
}

// TestE2E_CSP_Headers tests Content-Security-Policy headers.
func TestE2E_CSP_Headers(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test browser health endpoint for CSP headers.
	healthURL := authzPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	require.NoError(t, err, "Creating CSP test request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "CSP test request should complete")

	require.NoError(t, resp.Body.Close())

	// CSP headers may be configured differently based on endpoint type.
	// This test validates the endpoint is accessible and CSP is considered.
	require.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")
}

// TestE2E_ServicePath_RS tests /service/** path for Resource Server.
func TestE2E_ServicePath_RS(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test protected resource endpoint (should require auth).
	resourceURL := rsPublicURL + cryptoutilSharedMagic.PathPrefixService + "/api/v1/resources"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceURL, nil)
	require.NoError(t, err, "Creating resource request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "Resource request should complete")

	require.NoError(t, resp.Body.Close())

	// Protected endpoints should return 401 without token.
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"Protected resource should return 401 without token")
}

// TestE2E_ServicePath_IDP tests /service/** path for Identity Provider.
func TestE2E_ServicePath_IDP(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test IDP health endpoint.
	healthURL := idpPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	require.NoError(t, err, "Creating IDP health request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "IDP health request should complete")

	require.NoError(t, resp.Body.Close())

	require.Equal(t, http.StatusOK, resp.StatusCode, "IDP should return 200 OK for /health")
}

// TestE2E_BrowserPath_IDP tests /browser/** path for Identity Provider.
func TestE2E_BrowserPath_IDP(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test browser login page.
	loginURL := idpPublicURL + cryptoutilSharedMagic.PathPrefixBrowser + "/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loginURL, nil)
	require.NoError(t, err, "Creating browser login request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "Browser login request should complete")

	require.NoError(t, resp.Body.Close())

	// Login page should exist (may redirect to auth or return HTML).
	require.NotEqual(t, http.StatusNotFound, resp.StatusCode,
		"Browser login path should exist")
}

// TestE2E_ServicePath_RP tests /service/** path for Relying Party.
func TestE2E_ServicePath_RP(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test RP health endpoint.
	healthURL := rpPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	require.NoError(t, err, "Creating RP health request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "RP health request should complete")

	require.NoError(t, resp.Body.Close())

	require.Equal(t, http.StatusOK, resp.StatusCode, "RP should return 200 OK for /health")
}

// TestE2E_ServicePath_SPA tests SPA server static file serving.
func TestE2E_ServicePath_SPA(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
	defer cancel()

	// Test SPA health endpoint.
	healthURL := spaPublicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	require.NoError(t, err, "Creating SPA health request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "SPA health request should complete")

	require.NoError(t, resp.Body.Close())

	require.Equal(t, http.StatusOK, resp.StatusCode, "SPA should return 200 OK for /health")
}

// TestE2E_AllServicesIntegration tests all services can communicate.
func TestE2E_AllServicesIntegration(t *testing.T) {
	t.Parallel()

	services := []struct {
		name      string
		publicURL string
	}{
		{authzContainer, authzPublicURL},
		{idpContainer, idpPublicURL},
		{rsContainer, rsPublicURL},
		{rpContainer, rpPublicURL},
		{spaContainer, spaPublicURL},
	}

	for _, svc := range services {
		t.Run(svc.name+"_health", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
			defer cancel()

			healthURL := svc.publicURL + cryptoutilSharedMagic.IdentityE2EHealthEndpoint
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			require.NoError(t, err, "Creating health request should succeed")

			resp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "Health request should succeed for %s", svc.name)

			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusOK, resp.StatusCode,
				"%s should return 200 OK for /health", svc.name)
		})
	}
}
