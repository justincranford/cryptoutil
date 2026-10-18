// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	authzResponseEncoderFailure = "injected response encoder failure"
	authzResponseTestIssuer     = "https://issuer.example.test"
)

func TestAdminServer_HealthHandlersReturnEncodingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		ready    bool
		shutdown bool
		handler  func(*AdminServer) fiber.Handler
		wantErr  string
	}{
		{
			name:    "livez running",
			path:    cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			handler: func(server *AdminServer) fiber.Handler { return server.handleLivez },
			wantErr: "failed to send livez response",
		},
		{
			name:     "livez shutting down",
			path:     cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			shutdown: true,
			handler:  func(server *AdminServer) fiber.Handler { return server.handleLivez },
			wantErr:  "failed to send livez shutdown response",
		},
		{
			name:     "readyz shutting down",
			path:     cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			ready:    true,
			shutdown: true,
			handler:  func(server *AdminServer) fiber.Handler { return server.handleReadyz },
			wantErr:  "failed to send readyz shutdown response",
		},
		{
			name:    "readyz not ready",
			path:    cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			handler: func(server *AdminServer) fiber.Handler { return server.handleReadyz },
			wantErr: "failed to send readyz not-ready response",
		},
		{
			name:    "readyz ready",
			path:    cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			ready:   true,
			handler: func(server *AdminServer) fiber.Handler { return server.handleReadyz },
			wantErr: "failed to send readyz response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
				JSONEncoder: func(any) ([]byte, error) {
					return nil, errors.New(authzResponseEncoderFailure)
				},
				ErrorHandler: func(ctx *fiber.Ctx, _ error) error {
					return ctx.SendStatus(http.StatusInternalServerError)
				},
			})
			admin := &AdminServer{app: app, ready: tc.ready, shutdown: tc.shutdown}
			app.Get(tc.path, tc.handler(admin))

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	}
}

func TestPublicServer_ConfigurationHandlersReturnEncodingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		handler func(*PublicServer) fiber.Handler
	}{
		{
			name: "health",
			path: "/health",
			handler: func(server *PublicServer) fiber.Handler {
				return server.handleHealth
			},
		},
		{
			name: "discovery",
			path: cryptoutilSharedMagic.PathDiscovery,
			handler: func(server *PublicServer) fiber.Handler {
				return server.handleOpenIDConfiguration
			},
		},
		{
			name: "jwks",
			path: cryptoutilSharedMagic.PathJWKS,
			handler: func(server *PublicServer) fiber.Handler {
				return server.handleJWKS
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
				JSONEncoder: func(any) ([]byte, error) {
					return nil, errors.New(authzResponseEncoderFailure)
				},
				ErrorHandler: func(ctx *fiber.Ctx, _ error) error {
					return ctx.SendStatus(http.StatusInternalServerError)
				},
			})
			server := &PublicServer{
				cfg: &cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings{
					Issuer: authzResponseTestIssuer,
				},
			}
			app.Get(tc.path, tc.handler(server))

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	}
}
