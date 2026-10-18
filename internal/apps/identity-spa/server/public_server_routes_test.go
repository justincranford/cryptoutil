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
	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testPublicServerHealthPath             = "/health"
	testPublicServerHealthCaseName         = "health"
	testPublicServerLivezCaseName          = "livez"
	testPublicServerReadyzCaseName         = "readyz"
	testPublicServerConfigPath             = "/config.json"
	testPublicServerEndpointNotFoundResult = "endpoint not found"
)

func TestPublicServer_Routes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		rpOrigin     string
		wantStatus   int
		wantResponse string
	}{
		{name: testPublicServerHealthCaseName, path: testPublicServerHealthPath, wantStatus: http.StatusOK, wantResponse: `"status":"healthy"`},
		{name: testPublicServerLivezCaseName, path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath, wantStatus: http.StatusOK, wantResponse: `"status":"live"`},
		{name: testPublicServerReadyzCaseName, path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, wantStatus: http.StatusOK, wantResponse: `"status":"ready"`},
		{name: "config without RP origin", path: testPublicServerConfigPath, wantStatus: http.StatusOK, wantResponse: `"service":"identity-spa"`},
		{name: "config with RP origin", path: testPublicServerConfigPath, rpOrigin: "https://rp.example.test", wantStatus: http.StatusOK, wantResponse: `"api_base_url":"https://rp.example.test/api"`},
		{name: "SPA fallback", path: "/account/profile", wantStatus: http.StatusOK, wantResponse: "SPA placeholder"},
		{name: "service API remains reserved", path: "/service/api/v1/unregistered", wantStatus: http.StatusNotFound, wantResponse: testPublicServerEndpointNotFoundResult},
		{name: "browser API remains reserved", path: "/browser/api/v1/unregistered", wantStatus: http.StatusNotFound, wantResponse: testPublicServerEndpointNotFoundResult},
		{name: "admin API remains reserved", path: "/admin/api/v1/unregistered", wantStatus: http.StatusNotFound, wantResponse: testPublicServerEndpointNotFoundResult},
		{name: "generic API remains reserved", path: "/api/unregistered", wantStatus: http.StatusNotFound, wantResponse: testPublicServerEndpointNotFoundResult},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
			})
			require.NoError(t, err)

			publicServer := NewPublicServer(base, &cryptoutilAppsIdentitySpaServerConfig.IdentitySPAServerSettings{
				RPOrigin:        tc.rpOrigin,
				IndexFile:       testSPAIndexFile,
				StaticFilesPath: testSPAStaticFilesPath,
			})
			publicServer.RegisterRoutes()

			response, err := base.App().Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tc.wantResponse)
		})
	}
}

func TestSPAServer_NewFromConfigValidation(t *testing.T) {
	t.Parallel()

	_, err := NewFromConfig(context.Background(), nil)
	require.ErrorContains(t, err, "config is nil")
}

func TestSPAServer_AccessorsWithoutApplication(t *testing.T) {
	t.Parallel()

	server := &SPAServer{}
	require.Equal(t, 0, server.PublicPort())
	require.Equal(t, 0, server.AdminPort())
	require.Empty(t, server.PublicBaseURL())
	require.Empty(t, server.AdminBaseURL())
	server.SetReady(true)
	require.NoError(t, server.Shutdown(context.Background()))
}
