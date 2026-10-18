// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsFrameworkTestHelpAPI "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity-rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestRelyingPartyPublicServer_JSONEncodingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		handler func(*PublicServer) fiber.Handler
	}{
		{
			name: testRPHealthCaseName,
			path: testRPHealthPath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleHealth
			},
		},
		{
			name: "readiness failure response",
			path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			handler: func(s *PublicServer) fiber.Handler {
				return s.handleReadyz
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := cryptoutilAppsIdentityRpServerConfig.DefaultTestConfig()
			if tc.path == cryptoutilSharedMagic.PrivateAdminReadyzRequestPath {
				cfg.AuthzServerURL = testRPInvalidAuthzServerURL
			}

			publicServer := &PublicServer{cfg: cfg}
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			cryptoutilAppsFrameworkTestHelpAPI.AssertJSONEncodingFailure(t, tc.path, request, tc.handler(publicServer))
		})
	}
}
