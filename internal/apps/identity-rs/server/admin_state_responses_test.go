// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAdminServer_StateResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		shutdown   bool
		ready      bool
		statusCode int
		handler    func(*AdminServer) fiber.Handler
	}{
		{
			name:       "livez during shutdown",
			path:       cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			shutdown:   true,
			statusCode: http.StatusServiceUnavailable,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleLivez
			},
		},
		{
			name:       "readyz during shutdown",
			path:       cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			shutdown:   true,
			ready:      true,
			statusCode: http.StatusServiceUnavailable,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleReadyz
			},
		},
		{
			name:       "readyz before ready",
			path:       cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			statusCode: http.StatusServiceUnavailable,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleReadyz
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			server := &AdminServer{shutdown: tc.shutdown, ready: tc.ready}
			app.Get(tc.path, tc.handler(server))

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, tc.statusCode, response.StatusCode)
		})
	}
}
