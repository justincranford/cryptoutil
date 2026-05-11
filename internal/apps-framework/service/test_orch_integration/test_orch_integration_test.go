// Copyright (c) 2025-2026 Justin Cranford.

package test_orch_integration

import (
	"context"
	"crypto/x509"
	"fmt"
	"sync"
	"testing"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeServiceServer struct {
	mu          sync.Mutex
	publicPort  int
	adminPort   int
	ready       bool
	startErr    error
	shutdownErr error
	shutdownHit bool
	setPorts    bool
}

func (f *fakeServiceServer) Start(context.Context) error {
	if f.startErr != nil {
		return f.startErr
	}

	if f.setPorts {
		f.mu.Lock()
		f.publicPort = 14080
		f.adminPort = 14090
		f.mu.Unlock()
	}

	return nil
}

func (f *fakeServiceServer) Shutdown(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.shutdownHit = true

	return f.shutdownErr
}

func (f *fakeServiceServer) DB() *gorm.DB                                           { return nil }
func (f *fakeServiceServer) App() *cryptoutilAppsFrameworkServiceServer.Application { return nil }

func (f *fakeServiceServer) PublicPort() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.publicPort
}

func (f *fakeServiceServer) AdminPort() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.adminPort
}

func (f *fakeServiceServer) SetReady(ready bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ready = ready
}

func (f *fakeServiceServer) PublicBaseURL() string                                  { return "" }
func (f *fakeServiceServer) AdminBaseURL() string                                   { return "" }
func (f *fakeServiceServer) TLSRootCAPool() *x509.CertPool                          { return nil }
func (f *fakeServiceServer) AdminTLSRootCAPool() *x509.CertPool                     { return nil }
func (f *fakeServiceServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService      { return nil }
func (f *fakeServiceServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService { return nil }
func (f *fakeServiceServer) Barrier() *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	return nil
}

func TestStartIntegrationServer_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		srv       cryptoutilAppsFrameworkServiceServer.ServiceServer
		ctx       context.Context
		wantError string
	}{
		{name: "nil server", srv: nil, ctx: context.Background(), wantError: "server cannot be nil"},
		{name: "start error", srv: &fakeServiceServer{startErr: fmt.Errorf("injected start failure")}, ctx: context.Background(), wantError: "server failed to start"},
		{name: "success", srv: &fakeServiceServer{setPorts: true}, ctx: context.Background()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			is, err := StartIntegrationServer(tc.ctx, t, tc.srv, nil)
			if tc.wantError != "" {
				require.Nil(t, is)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, is)
			require.Equal(t, "https://127.0.0.1:14080", is.PublicBaseURL())
			require.Equal(t, "https://127.0.0.1:14090", is.AdminBaseURL())
			require.Nil(t, is.DB())
			require.Equal(t, tc.srv, is.Server())
			require.NoError(t, is.Shutdown(context.Background()))
		})
	}
}

func TestStartIntegrationServerForTestMain_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		srv       cryptoutilAppsFrameworkServiceServer.ServiceServer
		ctx       context.Context
		wantError string
	}{
		{name: "nil server", srv: nil, ctx: context.Background(), wantError: "server cannot be nil"},
		{name: "start error", srv: &fakeServiceServer{startErr: fmt.Errorf("injected start failure")}, ctx: context.Background(), wantError: "server failed to start"},
		{name: "timeout waiting for ports", srv: &fakeServiceServer{setPorts: false}, ctx: func() context.Context {
			c, cancel := context.WithCancel(context.Background())
			cancel()

			return c
		}(), wantError: "timed out waiting for server ports"},
		{name: "success", srv: &fakeServiceServer{setPorts: true}, ctx: context.Background()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			is, err := StartIntegrationServerForTestMain(tc.ctx, tc.srv, nil)
			if tc.wantError != "" {
				require.Nil(t, is)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, is)
			require.NoError(t, is.Shutdown(context.Background()))
		})
	}
}

func TestIntegrationServer_ShutdownAndAccessors(t *testing.T) {
	t.Parallel()

	t.Run("nil server shutdown is no-op", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, (&IntegrationServer{}).Shutdown(context.Background()))
	})

	t.Run("shutdown error propagated", func(t *testing.T) {
		t.Parallel()

		is := &IntegrationServer{srv: &fakeServiceServer{shutdownErr: fmt.Errorf("injected shutdown failure")}}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "shutdown failed")
	})

	t.Run("cleanup error propagated", func(t *testing.T) {
		t.Parallel()

		is := &IntegrationServer{
			srv:       &fakeServiceServer{},
			cleanupFn: func() error { return fmt.Errorf("injected cleanup failure") },
		}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "cleanup failed")
	})

	t.Run("shutdown error with testing tb logs and propagates", func(t *testing.T) {
		t.Parallel()
		is := &IntegrationServer{tb: t, srv: &fakeServiceServer{shutdownErr: fmt.Errorf("tb shutdown failure")}}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "shutdown failed")
	})

	t.Run("cleanup error with testing tb logs and propagates", func(t *testing.T) {
		t.Parallel()
		is := &IntegrationServer{
			tb:        t,
			srv:       &fakeServiceServer{},
			cleanupFn: func() error { return fmt.Errorf("tb cleanup failure") },
		}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "cleanup failed")
	})

	t.Run("empty urls when server nil", func(t *testing.T) {
		t.Parallel()

		is := &IntegrationServer{}
		require.Equal(t, "", is.PublicBaseURL())
		require.Equal(t, "", is.AdminBaseURL())
	})
}

func TestBuildBrokenFixtures_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		create        func(t *testing.T) (*IntegrationServer, error)
		wantBrokenDB  bool
		wantBrokenAPI bool
		wantErrPart   string
	}{
		{
			name: "broken db fixture",
			create: func(t *testing.T) (*IntegrationServer, error) {
				return BuildBrokenDBFixture(t, "db reason", &fakeServiceServer{})
			},
			wantBrokenDB: true,
			wantErrPart:  "db reason",
		},
		{
			name: "broken api fixture",
			create: func(t *testing.T) (*IntegrationServer, error) {
				return BuildBrokenAPIFixture(t, "api reason", &fakeServiceServer{}, nil)
			},
			wantBrokenAPI: true,
			wantErrPart:   "api reason",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			is, err := tc.create(t)
			require.NoError(t, err)
			require.NotNil(t, is)
			require.Equal(t, tc.wantBrokenDB, is.BrokenDB())
			require.Equal(t, tc.wantBrokenAPI, is.BrokenAPI())

			if tc.wantBrokenDB {
				require.ErrorContains(t, is.BrokenDBError(), tc.wantErrPart)
				require.Nil(t, is.BrokenAPIError())
			} else {
				require.ErrorContains(t, is.BrokenAPIError(), tc.wantErrPart)
				require.Nil(t, is.BrokenDBError())
			}
		})
	}
}

func TestBuildBrokenFixtures_CleanupClosurePaths(t *testing.T) {
	t.Parallel()

	t.Run("broken db cleanup closure logs cleanup error", func(t *testing.T) {
		t.Parallel()
		is, err := BuildBrokenDBFixture(t, "db cleanup path", &fakeServiceServer{})
		require.NoError(t, err)
		require.NotNil(t, is)
		is.cleanupFn = func() error { return fmt.Errorf("cleanup failure") }
	})

	t.Run("broken api cleanup closure logs cleanup error", func(t *testing.T) {
		t.Parallel()
		is, err := BuildBrokenAPIFixture(t, "api cleanup path", &fakeServiceServer{}, nil)
		require.NoError(t, err)
		require.NotNil(t, is)
		is.cleanupFn = func() error { return fmt.Errorf("cleanup failure") }
	})
}
