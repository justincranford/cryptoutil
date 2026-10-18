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
	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity-idp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	idpTestHealthRouteName     = "health"
	idpTestHealthRoutePath     = "/health"
	idpTestHealthRouteResponse = `"status":"healthy"`
	idpTestLivezRouteName      = "livez"
	idpTestOKResponse          = "OK"
	idpTestReadyzRouteName     = "readyz"
)

func TestPublicServer_Routes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantResponse string
	}{
		{name: idpTestHealthRouteName, path: idpTestHealthRoutePath, wantStatus: http.StatusOK, wantResponse: idpTestHealthRouteResponse},
		{name: idpTestLivezRouteName, path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath, wantStatus: http.StatusOK, wantResponse: idpTestOKResponse},
		{name: idpTestReadyzRouteName, path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, wantStatus: http.StatusOK, wantResponse: idpTestOKResponse},
		{name: "login page", path: "/browser/login", wantStatus: http.StatusOK, wantResponse: "Identity Provider Login"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
			})
			require.NoError(t, err)

			publicServer := NewPublicServer(base, cryptoutilAppsIdentityIdpServerConfig.DefaultTestConfig())
			publicServer.registerRoutes()

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

func TestIDPServer_NewFromConfigValidation(t *testing.T) {
	t.Parallel()

	var nilContext context.Context

	_, err := NewFromConfig(nilContext, cryptoutilAppsIdentityIdpServerConfig.DefaultTestConfig())
	require.ErrorContains(t, err, "context cannot be nil")

	_, err = NewFromConfig(context.Background(), nil)
	require.ErrorContains(t, err, "config cannot be nil")
}
