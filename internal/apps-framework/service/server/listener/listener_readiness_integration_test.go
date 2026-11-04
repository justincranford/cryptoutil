// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"context"
	"errors"
	"io"
	"net"
	http "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type serverStartedNotifier interface {
	Started() <-chan struct{}
	signalStarted()
}

func TestStartedSignalsAreInitializedAndIdempotent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server serverStartedNotifier
	}{
		{name: "admin", server: &AdminServer{}},
		{name: cryptoutilSharedMagic.PublicLabel, server: &PublicHTTPServer{}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			started := testCase.server.Started()
			require.NotNil(t, started)

			select {
			case <-started:
				t.Fatal("started channel closed before the listener was bound")
			default:
			}

			testCase.server.signalStarted()
			require.Equal(t, started, testCase.server.Started())

			select {
			case <-started:
			default:
				t.Fatal("started channel was not closed")
			}

			testCase.server.signalStarted()
			require.Equal(t, started, testCase.server.Started())
		})
	}
}

func TestReadyzReportsStaleDependency(t *testing.T) {
	t.Parallel()

	observedAt := time.Now().UTC()

	var currentUnixNano atomic.Int64
	currentUnixNano.Store(observedAt.UnixNano())

	monitor, err := cryptoutilAppsFrameworkServiceServerReadiness.New(cryptoutilAppsFrameworkServiceServerReadiness.Config{
		Checks: []cryptoutilAppsFrameworkServiceServerReadiness.Check{{
			Name:       "stale_database",
			ReasonCode: "stale_database_unavailable",
			Run:        func(context.Context) error { return nil },
		}},
		Period:       time.Hour,
		MaxStaleness: time.Second,
		Now: func() time.Time {
			return time.Unix(0, currentUnixNano.Load()).UTC()
		},
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)
	require.NoError(t, monitor.Start(context.Background()))

	defer func() { require.NoError(t, monitor.Stop(context.Background())) }()

	require.NoError(t, monitor.WaitInitial(context.Background()))

	currentUnixNano.Store(observedAt.Add(2 * time.Second).UnixNano())

	server := &AdminServer{app: fiber.New()}
	server.registerRoutes()
	server.MarkStartupComplete()
	server.SetReady(true)
	server.SetReadinessMonitor(monitor)

	request := httptest.NewRequest(http.MethodGet,
		cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, nil)
	response, err := server.app.Test(request, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, response.Body.Close()) }()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, response.StatusCode)
	require.Contains(t, string(body), "stale")
	require.Contains(t, string(body), "stale_database_unavailable")
}

func TestStartAfterShutdownDuringListenerCreation(t *testing.T) {
	t.Parallel()

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		server, err := NewAdminHTTPServer(
			context.Background(),
			cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
			validAutoTLSSettings(t),
		)
		require.NoError(t, err)

		server.listenFn = func(ctx context.Context, _, _ string) (net.Listener, error) {
			require.NoError(t, server.Shutdown(ctx))

			return testNetListener{addr: &net.TCPAddr{Port: 1}}, nil
		}

		require.ErrorContains(t, server.Start(context.Background()), "admin server already shutdown")
	})

	t.Run(cryptoutilSharedMagic.PublicLabel, func(t *testing.T) {
		t.Parallel()

		server, err := NewPublicHTTPServer(
			context.Background(),
			cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
			validAutoTLSSettings(t),
		)
		require.NoError(t, err)

		server.listenFn = func(ctx context.Context, _, _ string) (net.Listener, error) {
			require.NoError(t, server.Shutdown(ctx))

			return testNetListener{addr: &net.TCPAddr{Port: 1}}, nil
		}

		require.ErrorContains(t, server.Start(context.Background()), "public server already shutdown")
	})
}

func TestStartReturnsNilAfterListenerClosesDuringShutdown(t *testing.T) {
	t.Parallel()

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		server, err := NewAdminHTTPServer(
			context.Background(),
			cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
			validAutoTLSSettings(t),
		)
		require.NoError(t, err)

		var shutdownErr error

		server.appListenerFn = func(_ *fiber.App, _ net.Listener) error {
			shutdownErr = server.Shutdown(context.Background())

			return net.ErrClosed
		}

		require.NoError(t, server.Start(context.Background()))
		require.NoError(t, shutdownErr)
	})

	t.Run(cryptoutilSharedMagic.PublicLabel, func(t *testing.T) {
		t.Parallel()

		server, err := NewPublicHTTPServer(
			context.Background(),
			cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
			validAutoTLSSettings(t),
		)
		require.NoError(t, err)

		var shutdownErr error

		server.appListenerFn = func(_ *fiber.App, _ net.Listener) error {
			shutdownErr = server.Shutdown(context.Background())

			return net.ErrClosed
		}

		require.NoError(t, server.Start(context.Background()))
		require.NoError(t, shutdownErr)
	})
}

func TestStartReturnsListenerErrorsOutsideShutdown(t *testing.T) {
	t.Parallel()

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		server, err := NewAdminHTTPServer(
			context.Background(),
			cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
			validAutoTLSSettings(t),
		)
		require.NoError(t, err)

		server.appListenerFn = func(_ *fiber.App, listener net.Listener) error {
			_ = listener.Close()

			return errors.New("injected non-shutdown listener error")
		}

		require.ErrorContains(t, server.Start(context.Background()), "injected non-shutdown listener error")
	})

	t.Run(cryptoutilSharedMagic.PublicLabel, func(t *testing.T) {
		t.Parallel()

		server, err := NewPublicHTTPServer(
			context.Background(),
			cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
			validAutoTLSSettings(t),
		)
		require.NoError(t, err)

		server.appListenerFn = func(_ *fiber.App, listener net.Listener) error {
			_ = listener.Close()

			return errors.New("injected non-shutdown listener error")
		}

		require.ErrorContains(t, server.Start(context.Background()), "injected non-shutdown listener error")
	})
}
