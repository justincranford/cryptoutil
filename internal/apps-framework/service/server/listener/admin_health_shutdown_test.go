// Copyright (c) 2025-2026 Justin Cranford.
//
//

package listener_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	http "net/http"
	"strconv"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceServerListener "cryptoutil/internal/apps-framework/service/server/listener"
	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps-framework/service/server/testutil"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// startAdminServerWithClient creates an admin server, starts it, and returns
// the server, an HTTPS client, the port, cancel func, and completion channel.
func startAdminServerWithClient(t *testing.T) (*cryptoutilAppsFrameworkServiceServerListener.AdminServer, *http.Client, int, context.CancelFunc, <-chan error) {
	t.Helper()

	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	settings := cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings()

	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	startErr := make(chan error, 1)

	go func() {
		startErr <- server.Start(ctx)
	}()

	require.Eventually(t, func() bool { return server.ActualPort() > 0 },
		cryptoutilSharedMagic.DefaultHealthCheckTimeout, cryptoutilSharedMagic.DefaultHealthCheckInterval,
		"server should allocate a dynamic port")

	port := server.ActualPort()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    cryptoutilAppsFrameworkServiceServerTestutil.PrivateRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	return server, client, port, cancel, startErr
}

func requireAdminServerStoppedOnShutdown(t *testing.T, startErr <-chan error) {
	t.Helper()

	// The shutdown endpoint may win Start's select before context cancellation.
	if err := waitForListenerServerExit(t, startErr); err != nil {
		require.ErrorIs(t, err, context.Canceled)
	}
}

// TestAdminServer_Livez_DuringHTTPShutdown covers handleLivez's "return nil" after
// the shutdown JSON response while the listener is draining.
// Sequential: starts a real HTTPS server (port reuse issues).
func TestAdminServer_Livez_DuringHTTPShutdown(t *testing.T) {
	// NOT parallel: starts admin server.
	_, client, port, cancel, startErr := startAdminServerWithClient(t)

	defer func() {
		cancel()
		requireAdminServerStoppedOnShutdown(t, startErr)
	}()

	// POST to /admin/api/v1/shutdown and probe liveness during the drain delay.
	shutdownURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + "/admin/api/v1/shutdown"
	shutdownReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, shutdownURL, nil)
	require.NoError(t, err)

	shutdownResp, err := client.Do(shutdownReq)
	require.NoError(t, err)

	_, _ = io.ReadAll(shutdownResp.Body)
	_ = shutdownResp.Body.Close()

	// During the drain delay, the listener remains available but reports shutdown.
	livezURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + "/admin/api/v1/livez"
	livezReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, livezURL, nil)
	require.NoError(t, err)

	livezResp, err := client.Do(livezReq)
	require.NoError(t, err)

	defer func() { _ = livezResp.Body.Close() }()

	require.Equal(t, http.StatusServiceUnavailable, livezResp.StatusCode)
}

// TestAdminServer_Readyz_DuringHTTPShutdown covers handleReadyz's "return nil" after
// the shutdown JSON response while the listener is draining.
// Sequential: starts a real HTTPS server (port reuse issues).
func TestAdminServer_Readyz_DuringHTTPShutdown(t *testing.T) {
	// NOT parallel: starts admin server.
	server, client, port, cancel, startErr := startAdminServerWithClient(t)

	defer func() {
		cancel()
		requireAdminServerStoppedOnShutdown(t, startErr)
	}()

	// Mark server ready before testing shutdown response.
	server.SetReady(true)

	// POST to /admin/api/v1/shutdown and probe readiness during the drain delay.
	shutdownURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + "/admin/api/v1/shutdown"
	shutdownReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, shutdownURL, nil)
	require.NoError(t, err)

	shutdownResp, err := client.Do(shutdownReq)
	require.NoError(t, err)

	_, _ = io.ReadAll(shutdownResp.Body)
	_ = shutdownResp.Body.Close()

	// During the drain delay, the listener remains available but reports shutdown.
	readyzURL := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + "/admin/api/v1/readyz"
	readyzReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, readyzURL, nil)
	require.NoError(t, err)

	readyzResp, err := client.Do(readyzReq)
	require.NoError(t, err)

	defer func() { _ = readyzResp.Body.Close() }()

	require.Equal(t, http.StatusServiceUnavailable, readyzResp.StatusCode)
}
