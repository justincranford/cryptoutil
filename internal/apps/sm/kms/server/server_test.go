package server

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http/httptest"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm/kms/server/businesslogic"
	cryptoutilKmsServerMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerMiddleware "cryptoutil/internal/apps/template/service/server/middleware"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// stubPublicServer implements IPublicServer for testing accessor branches.
type stubPublicServer struct{ startErr error }

func (s *stubPublicServer) Start(context.Context) error    { return s.startErr }
func (s *stubPublicServer) Shutdown(context.Context) error { return nil }
func (s *stubPublicServer) ActualPort() int                { return 8443 }
func (s *stubPublicServer) PublicBaseURL() string          { return "https://localhost:8443" }

// stubAdminServer implements IAdminServer for testing accessor branches.
type stubAdminServer struct{}

func (s *stubAdminServer) Start(context.Context) error        { return nil }
func (s *stubAdminServer) Shutdown(context.Context) error     { return nil }
func (s *stubAdminServer) ActualPort() int                    { return cryptoutilSharedMagic.JoseJAAdminPort }
func (s *stubAdminServer) SetReady(bool)                      {}
func (s *stubAdminServer) AdminBaseURL() string               { return "https://localhost:9090" }
func (s *stubAdminServer) AdminTLSRootCAPool() *x509.CertPool { return nil }

func newTestApp(t *testing.T) *cryptoutilAppsTemplateServiceServer.Application {
	t.Helper()

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(
		context.Background(), &stubPublicServer{}, &stubAdminServer{},
	)
	require.NoError(t, err)

	return app
}

func TestNewKMSServer_NilChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr string
	}{
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: "context cannot be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, err := NewKMSServer(tc.ctx, nil) //nolint:staticcheck // SA1012: Intentionally testing nil context handling
			require.Error(t, err)
			require.Nil(t, server)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestNewKMSServer_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	server, err := NewKMSServer(ctx, nil)
	require.Error(t, err)
	require.Nil(t, server)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

func TestKMSServer_StartNotInitialized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *KMSServer
	}{
		{
			name:   "nil resources",
			server: &KMSServer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.server.Start(context.Background())
			require.Error(t, err)
			require.Contains(t, err.Error(), "server not initialized")
		})
	}
}

func TestKMSServer_ShutdownNilFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *KMSServer
	}{
		{
			name:   "all nil fields",
			server: &KMSServer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Should not panic with nil fields.
			require.NotPanics(t, func() {
				_ = tc.server.Shutdown(context.Background())
			})
		})
	}
}

func TestKMSServer_Accessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *KMSServer
	}{
		{
			name:   "zero value server",
			server: &KMSServer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.False(t, tc.server.IsReady())
			require.Equal(t, 0, tc.server.PublicPort())
			require.Equal(t, 0, tc.server.AdminPort())
			require.Empty(t, tc.server.PublicBaseURL())
			require.Empty(t, tc.server.AdminBaseURL())
			require.Nil(t, tc.server.Resources())

			require.Nil(t, tc.server.Settings())
		})
	}
}

func TestKMSServer_AccessorsWithResources(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	srv := &KMSServer{
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
			Application: app,
		},
	}

	require.Equal(t, 8443, srv.PublicPort())
	require.Equal(t, cryptoutilSharedMagic.JoseJAAdminPort, srv.AdminPort())
	require.Equal(t, "https://localhost:8443", srv.PublicBaseURL())
	require.Equal(t, "https://localhost:9090", srv.AdminBaseURL())
	require.NotNil(t, srv.Resources())
}

func TestKMSServer_StartError(t *testing.T) {
	t.Parallel()

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(
		context.Background(),
		&stubPublicServer{startErr: fmt.Errorf("bind failed")},
		&stubAdminServer{},
	)
	require.NoError(t, err)

	srv := &KMSServer{
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
			Application: app,
		},
	}

	err = srv.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to start KMS server")
	require.True(t, srv.IsReady()) // ready was set before Application.Start blocked
}

func TestKMSServer_ShutdownWithResources(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	shutdownCoreCalled := false
	shutdownContainerCalled := false

	srv := &KMSServer{
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
			Application:       app,
			ShutdownCore:      func() { shutdownCoreCalled = true },
			ShutdownContainer: func() { shutdownContainerCalled = true },
		},
	}

	srv.ready.Store(true)
	require.NotPanics(t, func() { _ = srv.Shutdown(context.Background()) })
	require.False(t, srv.IsReady())
	require.True(t, shutdownCoreCalled)
	require.True(t, shutdownContainerCalled)
}

func TestKMSServer_ShutdownWithPartialResources(t *testing.T) {
	t.Parallel()

	srv := &KMSServer{
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{},
	}

	require.NotPanics(t, func() { _ = srv.Shutdown(context.Background()) })
}

func TestRegisterKMSRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	}{
		{
			name: "default paths",
			settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				PublicBrowserAPIContextPath: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
				PublicServiceAPIContextPath: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
			},
		},
		{
			name: "custom paths",
			settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				PublicBrowserAPIContextPath: "/custom/browser",
				PublicServiceAPIContextPath: "/custom/service",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			err := registerKMSRoutes(app, (*cryptoutilKmsServerBusinesslogic.BusinessLogicService)(nil), tc.settings, nil)
			require.NoError(t, err)

			// Verify routes were registered (Fiber's route stack should be non-empty).
			routes := app.GetRoutes()
			require.NotEmpty(t, routes)
		})
	}
}

func TestRegisterKMSRoutes_WithSessionManager(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		PublicBrowserAPIContextPath: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
		PublicServiceAPIContextPath: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
	}
	res := &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
		SessionManager: &cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService{},
	}

	err := registerKMSRoutes(app, (*cryptoutilKmsServerBusinesslogic.BusinessLogicService)(nil), settings, res)

	require.NoError(t, err)
	require.NotEmpty(t, app.GetRoutes())
}

func TestKMSServer_SetReady(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server func(t *testing.T) *KMSServer
	}{
		{
			name:   "nil resources",
			server: func(_ *testing.T) *KMSServer { return &KMSServer{} },
		},
		{
			name: "with application",
			server: func(t *testing.T) *KMSServer {
				t.Helper()

				return &KMSServer{
					resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
						Application: newTestApp(t),
					},
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := tc.server(t)
			srv.SetReady(true)
			require.True(t, srv.IsReady())
			srv.SetReady(false)
			require.False(t, srv.IsReady())
		})
	}
}

func TestKMSServer_MissingAccessors(t *testing.T) {
	t.Parallel()

	srv := &KMSServer{}

	require.Nil(t, srv.DB())
	require.Nil(t, srv.App())
	require.Equal(t, 0, srv.PublicServerActualPort())
	require.Equal(t, 0, srv.AdminServerActualPort())
	require.Nil(t, srv.TLSRootCAPool())
	require.Nil(t, srv.AdminTLSRootCAPool())
	require.Nil(t, srv.JWKGen())
	require.Nil(t, srv.Telemetry())
	require.Nil(t, srv.Barrier())
}

func TestKMSServer_MissingAccessorsWithResources(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	srv := &KMSServer{
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
			Application: app,
		},
	}

	require.Nil(t, srv.DB())
	require.Equal(t, app, srv.App())
	require.Equal(t, 8443, srv.PublicServerActualPort())
	require.Equal(t, cryptoutilSharedMagic.JoseJAAdminPort, srv.AdminServerActualPort())
	require.Nil(t, srv.TLSRootCAPool())
	require.Nil(t, srv.AdminTLSRootCAPool())
	require.Nil(t, srv.JWKGen())
	require.Nil(t, srv.Telemetry())
	require.Nil(t, srv.Barrier())
}

func TestTenantContextBridgeMiddleware(t *testing.T) {
	t.Parallel()

	validTID := googleUuid.New()
	validRID := googleUuid.New()

	tests := []struct {
		name     string
		setup    func(c *fiber.Ctx)
		expectRC bool
	}{
		{
			name: "valid tenant and realm IDs",
			setup: func(c *fiber.Ctx) {
				c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyTenantID, validTID)
				c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyRealmID, validRID)
			},
			expectRC: true,
		},
		{
			name: "nil UUID",
			setup: func(c *fiber.Ctx) {
				c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyTenantID, googleUuid.UUID{})
			},
			expectRC: false,
		},
		{
			name: "wrong type in locals",
			setup: func(c *fiber.Ctx) {
				c.Locals(cryptoutilAppsTemplateServiceServerMiddleware.ContextKeyTenantID, "not-a-uuid")
			},
			expectRC: false,
		},
		{
			name:     "no tenant ID set",
			setup:    func(*fiber.Ctx) {},
			expectRC: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})

			var capturedRC *cryptoutilKmsServerMiddleware.RealmContext

			fiberApp.Use(func(c *fiber.Ctx) error {
				tc.setup(c)

				return c.Next()
			})
			fiberApp.Use(tenantContextBridgeMiddleware())
			fiberApp.Get("/test", func(c *fiber.Ctx) error {
				capturedRC = cryptoutilKmsServerMiddleware.GetRealmContext(c.UserContext())

				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
			resp, err := fiberApp.Test(req, -1)

			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			if tc.expectRC {
				require.NotNil(t, capturedRC)
				require.Equal(t, validTID, capturedRC.TenantID)
			} else {
				require.Nil(t, capturedRC)
			}
		})
	}
}
