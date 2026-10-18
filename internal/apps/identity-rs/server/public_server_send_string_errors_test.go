// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestPublicServer_SendStringErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		handler func(*PublicServer) fiber.Handler
	}{
		{
			name: "liveness",
			path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleLivez
			},
		},
		{
			name: "readiness",
			path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleReadyz
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			server := NewPublicServer(nil, cryptoutilAppsIdentityRsServerConfig.DefaultTestConfig())
			server.sendStringFn = func(*fiber.Ctx, string) error {
				return errors.New("injected response write failure")
			}
			app.Get(tc.path, tc.handler(server))

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	}
}
