// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsFrameworkTestHelpAPI "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSinglePageAppPublicServer_JSONEncodingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		handler func(*PublicServer) fiber.Handler
	}{
		{
			name: "health",
			path: "/health",
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleHealth
			},
		},
		{
			name: "livez",
			path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleLivez
			},
		},
		{
			name: "readyz",
			path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleReadyz
			},
		},
		{
			name: testSPAConfigCaseName,
			path: "/config",
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleConfig
			},
		},
		{
			name: "reserved route fallback",
			path: "/service/unknown",
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleSPAFallback
			},
		},
		{
			name: "client route fallback",
			path: "/profile",
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleSPAFallback
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			publicServer := &PublicServer{
				cfg: &cryptoutilAppsIdentitySpaServerConfig.IdentitySPAServerSettings{
					IndexFile:       testSPAIndexFile,
					StaticFilesPath: testSPAStaticFilesPath,
				},
			}
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			cryptoutilAppsFrameworkTestHelpAPI.AssertJSONEncodingFailure(t, tc.path, request, tc.handler(publicServer))
		})
	}
}
