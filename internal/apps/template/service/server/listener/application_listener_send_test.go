// Copyright (c) 2025 Justin Cranford
//
//

package listener_test

import (
"context"
"sync"
"testing"
"time"

cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"

"github.com/stretchr/testify/require"
)

// startAdminServer starts an admin server with dynamic port allocation and returns
// the server, context cancel func, and WaitGroup. Caller must call cancel() then wg.Wait()
// to clean up.
func startAdminServer(t *testing.T) (*cryptoutilAppsTemplateServiceServerListener.AdminServer, func(), *sync.WaitGroup) {
t.Helper()

tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()

server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), settings, tlsCfg)
require.NoError(t, err)
require.NotNil(t, server)

ctx, cancel := context.WithCancel(context.Background())

var wg sync.WaitGroup

wg.Add(1)

go func() {
defer wg.Done()

_ = server.Start(ctx)
}()

// Wait for dynamic port assignment.
var port int

for i := 0; i < 10; i++ {
time.Sleep(50 * time.Millisecond)

port = server.ActualPort()
if port > 0 {
break
}
}

require.Greater(t, port, 0, "Expected dynamic port allocation")

return server, cancel, &wg
}

// TestSendLivenessCheck_Success tests the success path of SendLivenessCheck.
// Covers: return result, nil (the non-error path in SendLivenessCheck).
func TestSendLivenessCheck_Success(t *testing.T) {
	t.Parallel()

// NOT parallel: starts admin server with dynamic port.
server, cancel, wg := startAdminServer(t)

defer func() {
// Cancel context to allow Start() to exit cleanly (matches TestAdminServer_Shutdown_Endpoint pattern).
cancel()
wg.Wait()
}()

settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
settings.BindPrivatePort = uint16(server.ActualPort()) //nolint:gosec // Port validated by server.
settings.DevMode = true                                // Enable InsecureSkipVerify for self-signed cert.

result, err := cryptoutilAppsTemplateServiceServerListener.SendLivenessCheck(settings)
require.NoError(t, err)
require.NotNil(t, result)
}

// TestSendReadinessCheck_Success tests the success path of SendReadinessCheck.
// Covers: return result, nil (the non-error path in SendReadinessCheck).
func TestSendReadinessCheck_Success(t *testing.T) {
	t.Parallel()

// NOT parallel: starts admin server with dynamic port.
server, cancel, wg := startAdminServer(t)

defer func() {
cancel()
wg.Wait()
}()

// Mark server ready so readyz returns 200.
server.SetReady(true)

settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
settings.BindPrivatePort = uint16(server.ActualPort()) //nolint:gosec // Port validated by server.
settings.DevMode = true

result, err := cryptoutilAppsTemplateServiceServerListener.SendReadinessCheck(settings)
require.NoError(t, err)
require.NotNil(t, result)
}

// TestSendShutdownRequest_Success tests the success path of SendShutdownRequest.
// Covers: return nil (the non-error path in SendShutdownRequest).
// Pattern matches TestAdminServer_Shutdown_Endpoint: explicit cancel() + wg.Wait().
func TestSendShutdownRequest_Success(t *testing.T) {
	t.Parallel()

// NOT parallel: starts admin server with dynamic port.
server, cancel, wg := startAdminServer(t)

settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
settings.BindPrivatePort = uint16(server.ActualPort()) //nolint:gosec // Port validated by server.
settings.DevMode = true

err := cryptoutilAppsTemplateServiceServerListener.SendShutdownRequest(settings)
require.NoError(t, err)

// The HTTP shutdown endpoint sets s.shutdown=true but the Fiber app requires the context
// to be cancelled for Start() to exit. This matches TestAdminServer_Shutdown_Endpoint pattern.
cancel()
wg.Wait()
}
