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
	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity-rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testPublicServerOKResponse = "OK"

type authzRoundTripper func(*http.Request) (*http.Response, error)

func (fn authzRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestPublicServer_RoutesAndAuthorizationReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		authzURL     string
		transport    http.RoundTripper
		wantStatus   int
		wantResponse string
	}{
		{name: testRPHealthCaseName, path: testRPHealthPath, wantStatus: http.StatusOK, wantResponse: `"status":"healthy"`},
		{name: "livez", path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath, wantStatus: http.StatusOK, wantResponse: testPublicServerOKResponse},
		{name: "readyz without authorization server", path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, wantStatus: http.StatusOK, wantResponse: testPublicServerOKResponse},
		{
			name:     "readyz with healthy authorization server",
			path:     cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			authzURL: testRPAuthzServerURL,
			transport: authzRoundTripper(func(request *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: request}, nil
			}),
			wantStatus:   http.StatusOK,
			wantResponse: testPublicServerOKResponse,
		},
		{name: "readyz invalid URL", path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, authzURL: testRPInvalidAuthzServerURL, wantStatus: http.StatusServiceUnavailable, wantResponse: "failed to create request"},
		{
			name:     "readyz unavailable authorization server",
			path:     cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			authzURL: testRPAuthzServerURL,
			transport: authzRoundTripper(func(_ *http.Request) (*http.Response, error) {
				return nil, context.DeadlineExceeded
			}),
			wantStatus:   http.StatusServiceUnavailable,
			wantResponse: "failed to connect to authz server",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base, err := cryptoutilAppsFrameworkServiceServer.NewPublicServerBase(&cryptoutilAppsFrameworkServiceServer.PublicServerConfig{
				BindAddress: cryptoutilSharedMagic.IPv4Loopback,
				TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
			})
			require.NoError(t, err)

			cfg := cryptoutilAppsIdentityRpServerConfig.DefaultTestConfig()
			cfg.AuthzServerURL = tc.authzURL
			cfg.DevMode = true

			publicServer := NewPublicServer(base, cfg, nil, nil)
			if tc.transport != nil {
				publicServer.authzHTTPClient.Transport = tc.transport
			}

			require.NoError(t, publicServer.registerRoutes())

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

func TestRPServer_NewFromConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		config  func() *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings
		wantErr string
	}{
		{
			name: "nil context",
			config: func() *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings {
				return cryptoutilAppsIdentityRpServerConfig.DefaultTestConfig()
			},
			wantErr: "context cannot be nil",
		},
		{
			name: "nil config",
			ctx:  context.Background(),
			config: func() *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings {
				return nil
			},
			wantErr: "config cannot be nil",
		},
		{
			name: "nil framework settings",
			ctx:  context.Background(),
			config: func() *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings {
				cfg := cryptoutilAppsIdentityRpServerConfig.DefaultTestConfig()
				cfg.ServiceFrameworkServerSettings = nil

				return cfg
			},
			wantErr: "failed to build identity-rp service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewFromConfig(tc.ctx, tc.config())
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestRPServer_StartRejectsNilContext(t *testing.T) {
	t.Parallel()

	server, err := NewFromConfig(context.Background(), cryptoutilAppsIdentityRpServerConfig.DefaultTestConfig())
	require.NoError(t, err)

	var nilContext context.Context
	require.ErrorContains(t, server.Start(nilContext), "context cannot be nil")
	require.NoError(t, server.Shutdown(context.Background()))
}

func TestRPServer_CheckAuthZServer_NonOKStatus(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilAppsIdentityRpServerConfig.DefaultTestConfig()
	cfg.AuthzServerURL = testRPAuthzServerURL
	cfg.DevMode = true
	server := NewPublicServer(nil, cfg, nil, nil)
	server.authzHTTPClient.Transport = authzRoundTripper(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: http.NoBody, Request: request}, nil
	})

	require.ErrorContains(t, server.checkAuthZServer(), "authz server returned status")
}
