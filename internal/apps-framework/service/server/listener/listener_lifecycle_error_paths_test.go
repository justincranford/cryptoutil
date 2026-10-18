// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var errTestListenerAccept = errors.New("test listener Accept called")

const testListenerNetwork = "test"

type testListenerAddr struct{}

func (testListenerAddr) Network() string { return testListenerNetwork }
func (testListenerAddr) String() string  { return testListenerNetwork }

type testNetListener struct {
	addr     net.Addr
	closeErr error
}

func (l testNetListener) Accept() (net.Conn, error) { return nil, errTestListenerAccept }
func (l testNetListener) Close() error              { return l.closeErr }
func (l testNetListener) Addr() net.Addr            { return l.addr }

type serverLifecycle interface {
	Start(context.Context) error
	Shutdown(context.Context) error
}

func TestStartAfterShutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		newServer func(*testing.T) serverLifecycle
	}{
		{
			name: cryptoutilSharedMagic.PublicLabel,
			newServer: func(t *testing.T) serverLifecycle {
				t.Helper()
				server, err := NewPublicHTTPServer(
					context.Background(),
					cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
					validAutoTLSSettings(t),
				)
				require.NoError(t, err)

				return server
			},
		},
		{
			name: testListenerAdmin,
			newServer: func(t *testing.T) serverLifecycle {
				t.Helper()
				server, err := NewAdminHTTPServer(
					context.Background(),
					cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
					validAutoTLSSettings(t),
				)
				require.NoError(t, err)

				return server
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := tc.newServer(t)
			require.NoError(t, server.Shutdown(context.Background()))
			require.ErrorContains(t, server.Start(context.Background()), "already shutdown")
		})
	}
}

func TestAdminStart_InvalidListenerAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		addr      net.Addr
		wantError string
	}{
		{name: "non-TCP listener", addr: testListenerAddr{}, wantError: "not a TCP address"},
		{name: "out-of-range TCP port", addr: &net.TCPAddr{Port: 65536}, wantError: "invalid port number"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, err := NewAdminHTTPServer(
				context.Background(),
				cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
				validAutoTLSSettings(t),
			)
			require.NoError(t, err)

			server.listenFn = func(context.Context, string, string) (net.Listener, error) {
				return testNetListener{addr: tc.addr}, nil
			}

			require.ErrorContains(t, server.Start(context.Background()), tc.wantError)
		})
	}
}

func TestPublicStart_NonTCPListenerAddress(t *testing.T) {
	t.Parallel()

	server, err := NewPublicHTTPServer(
		context.Background(),
		cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		validAutoTLSSettings(t),
	)
	require.NoError(t, err)

	server.listenFn = func(context.Context, string, string) (net.Listener, error) {
		return testNetListener{addr: testListenerAddr{}}, nil
	}

	require.ErrorContains(t, server.Start(context.Background()), "not *net.TCPAddr")
}

func TestShutdownDuringListenerCreation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		start func(*testing.T, net.Listener) error
	}{
		{
			name: "public listener close failure",
			start: func(t *testing.T, listener net.Listener) error {
				t.Helper()
				server, err := NewPublicHTTPServer(
					context.Background(),
					cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
					validAutoTLSSettings(t),
				)
				require.NoError(t, err)

				server.listenFn = func(context.Context, string, string) (net.Listener, error) {
					require.NoError(t, server.Shutdown(context.Background()))

					return listener, nil
				}

				return server.Start(context.Background())
			},
		},
		{
			name: "admin listener close failure",
			start: func(t *testing.T, listener net.Listener) error {
				t.Helper()
				server, err := NewAdminHTTPServer(
					context.Background(),
					cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
					validAutoTLSSettings(t),
				)
				require.NoError(t, err)

				server.listenFn = func(context.Context, string, string) (net.Listener, error) {
					require.NoError(t, server.Shutdown(context.Background()))

					return listener, nil
				}

				return server.Start(context.Background())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.start(t, testNetListener{
				addr:     &net.TCPAddr{Port: 1},
				closeErr: errors.New("injected close failure"),
			})
			require.ErrorContains(t, err, "already shutdown; failed to close listener")
			require.ErrorContains(t, err, "injected close failure")
		})
	}
}

func TestAdminShutdown_AlreadyStarted(t *testing.T) {
	t.Parallel()

	server, err := NewAdminHTTPServer(
		context.Background(),
		cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		validAutoTLSSettings(t),
	)
	require.NoError(t, err)

	server.shutdownStarted = true
	close(server.shutdownDone)

	require.NoError(t, server.Shutdown(context.Background()))
}

func TestAdminShutdown_AlreadyStartedTimeout(t *testing.T) {
	t.Parallel()

	server, err := NewAdminHTTPServer(
		context.Background(),
		cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		validAutoTLSSettings(t),
	)
	require.NoError(t, err)

	server.shutdownStarted = true

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.ErrorContains(t, server.Shutdown(ctx), "timed out waiting for admin server shutdown")
}

func TestPublicShutdown_AlreadyShutdown(t *testing.T) {
	t.Parallel()

	server, err := NewPublicHTTPServer(
		context.Background(),
		cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		validAutoTLSSettings(t),
	)
	require.NoError(t, err)
	require.NoError(t, server.Shutdown(context.Background()))

	require.ErrorContains(t, server.Shutdown(context.Background()), "already shutdown")
}

func TestAdminShutdown_ListenerCloseFailure(t *testing.T) {
	t.Parallel()

	server, err := NewAdminHTTPServer(
		context.Background(),
		cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		validAutoTLSSettings(t),
	)
	require.NoError(t, err)

	server.listener = testNetListener{
		addr:     &net.TCPAddr{Port: 1},
		closeErr: errors.New("injected listener close failure"),
	}

	require.ErrorContains(t, server.Shutdown(context.Background()), "failed to close admin listener")
}

func TestPublicShutdown_ListenerCloseFailure(t *testing.T) {
	t.Parallel()

	server, err := NewPublicHTTPServer(
		context.Background(),
		cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		validAutoTLSSettings(t),
	)
	require.NoError(t, err)

	server.listener = testNetListener{
		addr:     &net.TCPAddr{Port: 1},
		closeErr: errors.New("injected listener close failure"),
	}

	require.ErrorContains(t, server.Shutdown(context.Background()), "failed to close public listener")
}
