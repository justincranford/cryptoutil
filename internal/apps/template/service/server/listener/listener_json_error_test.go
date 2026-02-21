// Copyright (c) 2025 Justin Cranford

package listener

import (
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
)

var errTestJSONFailure = errors.New("forced JSON failure")

func failingJSONEncoder(_ any) ([]byte, error) {
	return nil, errTestJSONFailure
}

// newAdminServerWithFailingJSON constructs an AdminServer with a JSON encoder that always fails.
func newAdminServerWithFailingJSON() *AdminServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Admin API Test",
		JSONEncoder:           failingJSONEncoder,
	})

	server := &AdminServer{
		app: app,
		settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
	}
	server.registerRoutes()

	return server
}

// newPublicServerWithFailingJSON constructs a PublicHTTPServer with a JSON encoder that always fails.
func newPublicServerWithFailingJSON() *PublicHTTPServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Public API Test",
		JSONEncoder:           failingJSONEncoder,
	})

	server := &PublicHTTPServer{
		app: app,
		settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{},
	}
	server.registerRoutes()

	return server
}

// TestAdminServer_HealthEndpoints_JSONSerializationErrors tests all JSON error paths in admin health handlers.
func TestAdminServer_HealthEndpoints_JSONSerializationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		setup   func(s *AdminServer)
		wantErr string
	}{
		{
			name: "livez alive JSON error",
			path: "/admin/api/v1/livez",
			setup: func(_ *AdminServer) {
				// Default: not shutdown, returns alive.
			},
			wantErr: "livez response",
		},
		{
			name: "livez shutdown JSON error",
			path: "/admin/api/v1/livez",
			setup: func(s *AdminServer) {
				s.shutdown = true
			},
			wantErr: "livez shutdown response",
		},
		{
			name: "readyz ready JSON error",
			path: "/admin/api/v1/readyz",
			setup: func(s *AdminServer) {
				s.ready = true
			},
			wantErr: "readyz response",
		},
		{
			name: "readyz not-ready JSON error",
			path: "/admin/api/v1/readyz",
			setup: func(_ *AdminServer) {
				// Default: not ready, not shutdown.
			},
			wantErr: "readyz not-ready response",
		},
		{
			name: "readyz shutdown JSON error",
			path: "/admin/api/v1/readyz",
			setup: func(s *AdminServer) {
				s.shutdown = true
			},
			wantErr: "readyz shutdown response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newAdminServerWithFailingJSON()
			tc.setup(server)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			resp, err := server.app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		})
	}
}

// TestPublicHTTPServer_HealthEndpoints_JSONSerializationErrors tests all JSON error paths in public health handlers.
func TestPublicHTTPServer_HealthEndpoints_JSONSerializationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		setup   func(s *PublicHTTPServer)
		wantErr string
	}{
		{
			name: "service health healthy JSON error",
			path: "/service/api/v1/health",
			setup: func(_ *PublicHTTPServer) {
				// Default: not shutdown.
			},
			wantErr: "service health response",
		},
		{
			name: "service health shutdown JSON error",
			path: "/service/api/v1/health",
			setup: func(s *PublicHTTPServer) {
				s.shutdown = true
			},
			wantErr: "service health shutdown response",
		},
		{
			name: "browser health healthy JSON error",
			path: "/browser/api/v1/health",
			setup: func(_ *PublicHTTPServer) {
				// Default: not shutdown.
			},
			wantErr: "browser health response",
		},
		{
			name: "browser health shutdown JSON error",
			path: "/browser/api/v1/health",
			setup: func(s *PublicHTTPServer) {
				s.shutdown = true
			},
			wantErr: "browser health shutdown response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newPublicServerWithFailingJSON()
			tc.setup(server)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			resp, err := server.app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		})
	}
}
