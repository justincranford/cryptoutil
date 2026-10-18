// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	authzTestHealthRouteName     = "health"
	authzTestHealthRoutePath     = "/health"
	authzTestHealthRouteResponse = `"status":"healthy"`
	authzTestLivezRouteName      = "livez"
	authzTestOIDCIssuerResponse  = "issuer"
	authzTestOKResponse          = "OK"
	authzTestReadyzRouteName     = "readyz"
)

func TestPublicServer_RoutesAndDiscoveryToggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		enableOIDC   bool
		path         string
		wantStatus   int
		wantResponse string
	}{
		{name: authzTestHealthRouteName, enableOIDC: true, path: authzTestHealthRoutePath, wantStatus: http.StatusOK, wantResponse: authzTestHealthRouteResponse},
		{name: authzTestLivezRouteName, enableOIDC: true, path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath, wantStatus: http.StatusOK, wantResponse: authzTestOKResponse},
		{name: authzTestReadyzRouteName, enableOIDC: true, path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, wantStatus: http.StatusOK, wantResponse: authzTestOKResponse},
		{name: "authorize page", enableOIDC: true, path: "/browser/api/v1/authorize", wantStatus: http.StatusOK, wantResponse: "OAuth 2.1 Authorization"},
		{name: "discovery enabled", enableOIDC: true, path: cryptoutilSharedMagic.PathDiscovery, wantStatus: http.StatusOK, wantResponse: authzTestOIDCIssuerResponse},
		{name: "jwks enabled", enableOIDC: true, path: cryptoutilSharedMagic.PathJWKS, wantStatus: http.StatusOK, wantResponse: `"keys":[]`},
		{name: "discovery disabled", path: cryptoutilSharedMagic.PathDiscovery, wantStatus: http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
			})
			require.NoError(t, err)

			publicServer := NewPublicServer(base, &cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings{
				Issuer:          "https://issuer.example.test",
				EnableDiscovery: tc.enableOIDC,
			})
			publicServer.registerRoutes()

			response, err := base.App().Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			if tc.wantResponse != "" {
				require.Contains(t, string(body), tc.wantResponse)
			}
		})
	}
}

func TestAuthzServer_NewFromConfigValidation(t *testing.T) {
	t.Parallel()

	var nilContext context.Context

	_, err := NewFromConfig(nilContext, cryptoutilAppsIdentityAuthzServerConfig.DefaultTestConfig())
	require.ErrorContains(t, err, "context cannot be nil")

	_, err = NewFromConfig(context.Background(), nil)
	require.ErrorContains(t, err, "config cannot be nil")

	adminChecks := []struct {
		name       string
		path       string
		ready      bool
		shutdown   bool
		wantStatus int
	}{
		{name: "livez while running", path: cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath, wantStatus: http.StatusOK},
		{name: "readyz before ready", path: cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, wantStatus: http.StatusServiceUnavailable},
		{name: "readyz after ready", path: cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, ready: true, wantStatus: http.StatusOK},
		{name: "livez during shutdown", path: cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath, shutdown: true, wantStatus: http.StatusServiceUnavailable},
		{name: "readyz during shutdown", path: cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, ready: true, shutdown: true, wantStatus: http.StatusServiceUnavailable},
	}
	for _, tc := range adminChecks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adminConfig := cryptoutilIdentityConfig.RequireNewForTest("authz-admin-health-routes")
			adminServer, serverErr := NewAdminHTTPServer(context.Background(), adminConfig)
			require.NoError(t, serverErr)

			adminServer.ready = tc.ready
			adminServer.shutdown = tc.shutdown

			response, requestErr := adminServer.app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, requestErr)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)
		})
	}
}
