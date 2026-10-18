// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsFrameworkTestHelpAPI "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
)

func TestResourceServerPublicServer_JSONEncodingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		authorization string
		handler       func(*PublicServer) fiber.Handler
	}{
		{
			name: "health",
			path: testHealthRoutePath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleHealth
			},
		},
		{
			name: "resource response without token",
			path: testResourcesRoutePath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleListResources
			},
		},
		{
			name:          "resource response with token",
			path:          testResourcesRoutePath,
			authorization: "Bearer token",
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleListResources
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			publicServer := &PublicServer{cfg: cryptoutilAppsIdentityRsServerConfig.DefaultTestConfig()}

			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.authorization != "" {
				request.Header.Set("Authorization", tc.authorization)
			}

			cryptoutilAppsFrameworkTestHelpAPI.AssertJSONEncodingFailure(t, tc.path, request, tc.handler(publicServer))
		})
	}
}
