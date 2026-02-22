package server

import (
	"context"
	"fmt"
	"testing"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"

	"github.com/stretchr/testify/require"
)

// stubPublicServer implements IPublicServer for testing accessor branches.
type stubPublicServer struct{ startErr error }

func (s *stubPublicServer) Start(context.Context) error  { return s.startErr }
func (s *stubPublicServer) Shutdown(context.Context) error { return nil }
func (s *stubPublicServer) ActualPort() int                { return 8443 }
func (s *stubPublicServer) PublicBaseURL() string          { return "https://localhost:8443" }

// stubAdminServer implements IAdminServer for testing accessor branches.
type stubAdminServer struct{}

func (s *stubAdminServer) Start(context.Context) error  { return nil }
func (s *stubAdminServer) Shutdown(context.Context) error { return nil }
func (s *stubAdminServer) ActualPort() int                { return 9090 }
func (s *stubAdminServer) SetReady(bool)                  {}
func (s *stubAdminServer) AdminBaseURL() string           { return "https://localhost:9090" }

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

			err := tc.server.Start()
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
		{
			name: "nil kmsCore only",
			server: &KMSServer{
				kmsCore: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Should not panic with nil fields.
			require.NotPanics(t, func() {
				tc.server.Shutdown()
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
			require.Nil(t, tc.server.KMSCore())
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
	require.Equal(t, 9090, srv.AdminPort())
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
		ctx: context.Background(),
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
			Application: app,
		},
	}

	err = srv.Start()
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
		ctx: context.Background(),
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{
			Application:       app,
			ShutdownCore:      func() { shutdownCoreCalled = true },
			ShutdownContainer: func() { shutdownContainerCalled = true },
		},
	}

	srv.ready.Store(true)
	require.NotPanics(t, func() { srv.Shutdown() })
	require.False(t, srv.IsReady())
	require.True(t, shutdownCoreCalled)
	require.True(t, shutdownContainerCalled)
}

func TestKMSServer_ShutdownWithPartialResources(t *testing.T) {
	t.Parallel()

	srv := &KMSServer{
		resources: &cryptoutilAppsTemplateServiceServerBuilder.ServiceResources{},
	}

	require.NotPanics(t, func() { srv.Shutdown() })
}
