// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// fakeAddr implements net.Addr but is not *net.TCPAddr. Used to test non-TCP address guard.
type fakeAddr struct{}

func (f *fakeAddr) Network() string { return "fake" }
func (f *fakeAddr) String() string  { return "fake:0" }

// fakeListener wraps a real listener but returns fakeAddr from Addr().
type fakeListener struct {
	inner net.Listener
}

func (f *fakeListener) Accept() (net.Conn, error) { return f.inner.Accept() }
func (f *fakeListener) Close() error              { return f.inner.Close() }
func (f *fakeListener) Addr() net.Addr            { return &fakeAddr{} }

// invalidPortListener wraps a real listener but returns an out-of-range port.
type invalidPortListener struct {
	inner net.Listener
}

func (f *invalidPortListener) Accept() (net.Conn, error) { return f.inner.Accept() }
func (f *invalidPortListener) Close() error              { return f.inner.Close() }
func (f *invalidPortListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 99999}
}

type stubTCPListener struct{}

func (stubTCPListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (stubTCPListener) Close() error              { return nil }
func (stubTCPListener) Addr() net.Addr {
	return &net.TCPAddr{Port: cryptoutilSharedMagic.TestAdminPort}
}

// TestAdminServer_Start_NonTCPAddr tests Start when listener returns a non-TCP address.
// Covers admin.go:208-214 (TCP addr type assertion failure guard).
func TestAdminServer_Start_NonTCPAddr(t *testing.T) {
	t.Parallel()

	stubListenFn := func(ctx context.Context, network, address string) (net.Listener, error) {
		realLn, err := (&net.ListenConfig{}).Listen(ctx, network, address)
		if err != nil {
			return nil, err
		}

		return &fakeListener{inner: realLn}, nil
	}

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := newAdminHTTPServerInternal(context.Background(), settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		stubListenFn,
		func(app *fiber.App, ln net.Listener) error { return app.Listener(ln) },
		os.ReadFile,
	)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "listener address is not a TCP address")
}

// TestAdminServer_Start_InvalidPort tests Start when listener returns an invalid port.
// Covers admin.go:216-222 (port range validation guard).
func TestAdminServer_Start_InvalidPort(t *testing.T) {
	t.Parallel()

	stubListenFn := func(ctx context.Context, network, address string) (net.Listener, error) {
		realLn, err := (&net.ListenConfig{}).Listen(ctx, network, address)
		if err != nil {
			return nil, err
		}

		return &invalidPortListener{inner: realLn}, nil
	}

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := newAdminHTTPServerInternal(context.Background(), settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		stubListenFn,
		func(app *fiber.App, ln net.Listener) error { return app.Listener(ln) },
		os.ReadFile,
	)
	require.NoError(t, err)
	err = server.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port number")
}

// TestAdminServer_Start_AppListenerError tests Start when app.Listener returns an error.
// Covers admin.go:240-242 (app.Listener error in goroutine).
func TestAdminServer_Start_AppListenerError(t *testing.T) {
	t.Parallel()

	stubAppListenerFn := func(_ *fiber.App, ln net.Listener) error {
		_ = ln.Close()

		return fmt.Errorf("forced admin listener error")
	}

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := newAdminHTTPServerInternal(context.Background(), settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		stubAppListenerFn,
		os.ReadFile,
	)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "admin server error")
	require.Contains(t, err.Error(), "forced admin listener error")
}

// TestAdminServer_ErrChanPath tests Start returns via errChan on clean Fiber shutdown.
// Covers admin.go errChan select case (without context cancellation).
func TestAdminServer_ErrChanPath(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := newAdminHTTPServerInternal(context.Background(), settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		func(_ context.Context, _, _ string) (net.Listener, error) {
			return stubTCPListener{}, nil
		},
		func(_ *fiber.App, _ net.Listener) error {
			return nil
		},
		os.ReadFile,
	)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.TestAdminPort, server.ActualPort())
}

// TestPublicServer_Start_NonTCPAddr tests Start when listener returns a non-TCP address.
// Covers public.go TCPAddr type assertion failure guard.
func TestPublicServer_Start_NonTCPAddr(t *testing.T) {
	t.Parallel()

	stubListenFn := func(ctx context.Context, network, address string) (net.Listener, error) {
		realLn, err := (&net.ListenConfig{}).Listen(ctx, network, address)
		if err != nil {
			return nil, err
		}

		return &fakeListener{inner: realLn}, nil
	}

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := newPublicHTTPServerInternal(context.Background(), settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		stubListenFn,
		func(app *fiber.App, ln net.Listener) error { return app.Listener(ln) },
		func(_ string) ([]byte, error) { return nil, nil },
	)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "listener address is not *net.TCPAddr")
}

// TestPublicServer_Start_AppListenerError tests Start when app.Listener returns an error.
// Covers public.go app.Listener error in goroutine.
func TestPublicServer_Start_AppListenerError(t *testing.T) {
	t.Parallel()

	stubAppListenerFn := func(_ *fiber.App, ln net.Listener) error {
		_ = ln.Close()

		return fmt.Errorf("forced public listener error")
	}

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	tlsCfg := validAutoTLSSettings(t)

	server, err := newPublicHTTPServerInternal(context.Background(), settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		stubAppListenerFn,
		func(_ string) ([]byte, error) { return nil, nil },
	)
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server error")
	require.Contains(t, err.Error(), "forced public listener error")
}
