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
	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testHealthRoutePath        = "/health"
	testResourcesRoutePath     = "/service/api/v1/resources"
	testPublicServerOKResponse = "OK"
	testRouteNameHealth        = "health"
	testRouteNameLivez         = "livez"
	testRouteNameReadyz        = "readyz"
)

func TestPublicServer_RoutesAndResourceAuthorization(t *testing.T) {
	t.Parallel()

	base, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
	})
	require.NoError(t, err)

	publicServer := NewPublicServer(base, cryptoutilAppsIdentityRsServerConfig.DefaultTestConfig())
	require.NoError(t, publicServer.registerRoutes())

	tests := []struct {
		name          string
		path          string
		authorization string
		wantStatus    int
		wantResponse  string
	}{
		{name: testRouteNameHealth, path: testHealthRoutePath, wantStatus: http.StatusOK, wantResponse: `"status":"healthy"`},
		{name: testRouteNameLivez, path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath, wantStatus: http.StatusOK, wantResponse: testPublicServerOKResponse},
		{name: testRouteNameReadyz, path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, wantStatus: http.StatusOK, wantResponse: testPublicServerOKResponse},
		{name: "resources without token", path: testResourcesRoutePath, wantStatus: http.StatusUnauthorized, wantResponse: "Authorization header required"},
		{name: "resources with token", path: testResourcesRoutePath, authorization: "Bearer example", wantStatus: http.StatusOK, wantResponse: "token validation not yet implemented"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.authorization != "" {
				request.Header.Set("Authorization", tc.authorization)
			}

			response, requestErr := base.App().Test(request, -1)

			require.NoError(t, requestErr)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			body, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)
			require.Contains(t, string(body), tc.wantResponse)
		})
	}
}

func TestRSServer_NewFromConfigValidation(t *testing.T) {
	t.Parallel()

	var nilContext context.Context

	_, err := NewFromConfig(nilContext, cryptoutilAppsIdentityRsServerConfig.DefaultTestConfig())
	require.ErrorContains(t, err, "context cannot be nil")

	_, err = NewFromConfig(context.Background(), nil)
	require.ErrorContains(t, err, "config cannot be nil")
}
