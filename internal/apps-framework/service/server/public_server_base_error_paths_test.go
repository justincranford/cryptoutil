// Copyright (c) 2025-2026 Justin Cranford.

package server

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type testNonTCPAddress struct{}

func (testNonTCPAddress) Network() string { return "test" }

func (testNonTCPAddress) String() string { return "test-address" }

type testNonTCPListener struct{}

func (testNonTCPListener) Accept() (net.Conn, error) { return nil, errors.New("accept is unused") }

func (testNonTCPListener) Close() error { return nil }

func (testNonTCPListener) Addr() net.Addr { return testNonTCPAddress{} }

type testPoolPublicServer struct{}

func (testPoolPublicServer) Start(context.Context) error { return nil }

func (testPoolPublicServer) Shutdown(context.Context) error { return nil }

func (testPoolPublicServer) ActualPort() int { return 0 }

func (testPoolPublicServer) PublicBaseURL() string { return "" }

type testPoolAdminServer struct {
	pool *x509.CertPool
}

func (testPoolAdminServer) Start(context.Context) error { return nil }

func (testPoolAdminServer) Shutdown(context.Context) error { return nil }

func (testPoolAdminServer) ActualPort() int { return 0 }

func (testPoolAdminServer) SetReady(bool) {}

func (testPoolAdminServer) AdminBaseURL() string { return "" }

func (s testPoolAdminServer) AdminTLSRootCAPool() *x509.CertPool { return s.pool }

func TestPublicServerBase_Start_NonTCPListenerAddress(t *testing.T) {
	t.Parallel()

	base, err := NewPublicServerBase(&PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
	})
	require.NoError(t, err)

	base.listenFn = func(context.Context, string, string) (net.Listener, error) {
		return testNonTCPListener{}, nil
	}

	err = base.Start(context.Background())

	require.ErrorContains(t, err, "listener address is not *net.TCPAddr")
	require.NoError(t, base.Shutdown(context.Background()))
}

func TestPublicServerBase_Shutdown_AppError(t *testing.T) {
	t.Parallel()

	base, err := NewPublicServerBase(&PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		TLSMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{},
	})
	require.NoError(t, err)

	base.appShutdownFn = func(*fiber.App) error {
		return errors.New("injected Fiber shutdown failure")
	}

	err = base.Shutdown(context.Background())

	require.ErrorContains(t, err, "injected Fiber shutdown failure")
}

func TestApplication_TLSRootCAPoolAccessors(t *testing.T) {
	t.Parallel()

	pool := x509.NewCertPool()
	base := &PublicServerBase{
		tlsMaterial: &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{RootCAPool: pool},
	}
	admin := testPoolAdminServer{pool: pool}

	tests := []struct {
		name           string
		app            *Application
		wantPublic     *x509.CertPool
		wantAdmin      *x509.CertPool
		wantPublicBase bool
	}{
		{
			name: "public server base and admin pool",
			app: &Application{
				publicServer: base,
				adminServer:  admin,
			},
			wantPublic:     pool,
			wantAdmin:      pool,
			wantPublicBase: true,
		},
		{
			name: "public server is not a base and admin has no pool",
			app: &Application{
				publicServer: testPoolPublicServer{},
				adminServer:  testPoolAdminServer{},
			},
		},
		{
			name:           "nil admin server",
			app:            &Application{publicServer: base},
			wantPublic:     pool,
			wantPublicBase: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.wantPublicBase, testCase.app.PublicServerBase() != nil)

			if testCase.wantPublic == nil {
				require.Nil(t, testCase.app.TLSRootCAPool())
			} else {
				require.Same(t, testCase.wantPublic, testCase.app.TLSRootCAPool())
			}

			if testCase.wantAdmin == nil {
				require.Nil(t, testCase.app.AdminTLSRootCAPool())
			} else {
				require.Same(t, testCase.wantAdmin, testCase.app.AdminTLSRootCAPool())
			}
		})
	}

	require.Nil(t, (&PublicServerBase{}).TLSRootCAPool())
}
