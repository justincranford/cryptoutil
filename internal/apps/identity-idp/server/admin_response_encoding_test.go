// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	idpResponseEncoderFailure = "injected response encoder failure"
	idpPublicHealthPath       = "/health"
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

			var handledErr error

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
				JSONEncoder: func(any) ([]byte, error) {
					return nil, errors.New(idpResponseEncoderFailure)
				},
				ErrorHandler: func(ctx *fiber.Ctx, err error) error {
					handledErr = err

					return ctx.SendStatus(http.StatusInternalServerError)
				},
			})
			admin := &AdminServer{app: app, ready: tc.ready, shutdown: tc.shutdown}
			app.Get(tc.path, tc.handler(admin))

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
			require.ErrorContains(t, handledErr, tc.wantErr)
			require.ErrorContains(t, handledErr, idpResponseEncoderFailure)
		})
	}
}

func TestPublicServer_HealthHandlerReturnsEncodingError(t *testing.T) {
	t.Parallel()

	var handledErr error

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		JSONEncoder: func(any) ([]byte, error) {
			return nil, errors.New(idpResponseEncoderFailure)
		},
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			handledErr = err

			return ctx.SendStatus(http.StatusInternalServerError)
		},
	})
	publicServer := &PublicServer{}
	app.Get(idpPublicHealthPath, publicServer.handleHealth)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, idpPublicHealthPath, nil), -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	require.ErrorContains(t, handledErr, "failed to send health response")
	require.ErrorContains(t, handledErr, idpResponseEncoderFailure)
}
