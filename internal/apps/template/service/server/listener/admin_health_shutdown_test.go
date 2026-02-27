// Copyright (c) 2025 Justin Cranford
//
//

package listener_test

import (
"context"
"crypto/tls"
"fmt"
"io"
http "net/http"
"sync"
"testing"
"time"

cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"


"github.com/stretchr/testify/require"
)

// startAdminServerWithClient creates an admin server, starts it, and returns
// the server, an HTTPS client, the port, cancel func, and WaitGroup.
// The HTTP shutdown endpoint sets s.shutdown=true but leaves Fiber running
// (100ms delay goroutine is a no-op due to early return in Shutdown()).
// This gives a reliable window to test health endpoints during "shutting down" state.
func startAdminServerWithClient(t *testing.T) (*cryptoutilAppsTemplateServiceServerListener.AdminServer, *http.Client, int, context.CancelFunc, *sync.WaitGroup) {
t.Helper()

tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()

server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), settings, tlsCfg)
require.NoError(t, err)

ctx, cancel := context.WithCancel(context.Background())

var wg sync.WaitGroup

wg.Add(1)

go func() {
defer wg.Done()

_ = server.Start(ctx)
}()

time.Sleep(200 * time.Millisecond)

port := server.ActualPort()
require.Greater(t, port, 0, "Expected dynamic port allocation")

client := &http.Client{
Transport: &http.Transport{
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
},
Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
}

return server, client, port, cancel, &wg
}

// TestAdminServer_Livez_DuringHTTPShutdown covers handleLivez's "return nil" after
// the shutdown JSON response when s.shutdown=true via the HTTP endpoint.
// The HTTP /admin/api/v1/shutdown endpoint sets s.shutdown=true but leaves Fiber running
// (scheduled goroutine is a no-op), providing a reliable window to test livez.
func TestAdminServer_Livez_DuringHTTPShutdown(t *testing.T) {
	t.Parallel()

// NOT parallel: starts admin server.
_, client, port, cancel, wg := startAdminServerWithClient(t)

defer func() {
cancel()
wg.Wait()
}()

// POST to /admin/api/v1/shutdown - sets s.shutdown=true, Fiber still running.
shutdownURL := fmt.Sprintf("https://%s:%d/admin/api/v1/shutdown", cryptoutilSharedMagic.IPv4Loopback, port)
shutdownReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, shutdownURL, nil)
require.NoError(t, err)

shutdownResp, err := client.Do(shutdownReq)
require.NoError(t, err)

_, _ = io.ReadAll(shutdownResp.Body)
_ = shutdownResp.Body.Close()

// Immediately call livez - s.shutdown=true, Fiber still accepting (< 100ms).
livezURL := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, port)
livezReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, livezURL, nil)
require.NoError(t, err)

livezResp, err := client.Do(livezReq)
require.NoError(t, err)

defer func() { _ = livezResp.Body.Close() }()

require.Equal(t, http.StatusServiceUnavailable, livezResp.StatusCode)
}

// TestAdminServer_Readyz_DuringHTTPShutdown covers handleReadyz's "return nil" after
// the shutdown JSON response when s.shutdown=true via the HTTP endpoint.
func TestAdminServer_Readyz_DuringHTTPShutdown(t *testing.T) {
	t.Parallel()

// NOT parallel: starts admin server.
server, client, port, cancel, wg := startAdminServerWithClient(t)

defer func() {
cancel()
wg.Wait()
}()

// Mark server ready before testing shutdown response.
server.SetReady(true)

// POST to /admin/api/v1/shutdown - sets s.shutdown=true, Fiber still running.
shutdownURL := fmt.Sprintf("https://%s:%d/admin/api/v1/shutdown", cryptoutilSharedMagic.IPv4Loopback, port)
shutdownReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, shutdownURL, nil)
require.NoError(t, err)

shutdownResp, err := client.Do(shutdownReq)
require.NoError(t, err)

_, _ = io.ReadAll(shutdownResp.Body)
_ = shutdownResp.Body.Close()

// Immediately call readyz - s.shutdown=true, Fiber still accepting (< 100ms).
readyzURL := fmt.Sprintf("https://%s:%d/admin/api/v1/readyz", cryptoutilSharedMagic.IPv4Loopback, port)
readyzReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, readyzURL, nil)
require.NoError(t, err)

readyzResp, err := client.Do(readyzReq)
require.NoError(t, err)

defer func() { _ = readyzResp.Body.Close() }()

require.Equal(t, http.StatusServiceUnavailable, readyzResp.StatusCode)
}
