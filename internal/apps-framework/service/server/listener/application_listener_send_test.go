// Copyright (c) 2025-2026 Justin Cranford.
//
//

package listener_test

import (
	"context"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceServerListener "cryptoutil/internal/apps-framework/service/server/listener"
	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps-framework/service/server/testutil"

	"github.com/stretchr/testify/require"
)

// startAdminServer starts an admin server with dynamic port allocation and returns
// the server, context cancel func, and bounded completion channel.
func startAdminServer(t *testing.T) (*cryptoutilAppsFrameworkServiceServerListener.AdminServer, func(), <-chan error) {
	t.Helper()

	tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PrivateTLS()
	settings := cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings()

	server, err := cryptoutilAppsFrameworkServiceServerListener.NewAdminHTTPServer(context.Background(), settings, tlsCfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	ctx, cancel := context.WithCancel(context.Background())

	startErr := make(chan error, 1)

	go func() {
		startErr <- server.Start(ctx)
	}()

	// Wait for dynamic port assignment.
	var port int

	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		time.Sleep(cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond)

		port = server.ActualPort()
		if port > 0 {
			break
		}
	}

	require.Greater(t, port, 0, "Expected dynamic port allocation")

	return server, cancel, startErr
}

func waitForListenerServerExit(t *testing.T, startErr <-chan error) error {
	t.Helper()

	select {
	case err := <-startErr:
		return err
	case <-time.After(cryptoutilSharedMagic.TestDefaultServerShutdownTimeout):
		require.FailNow(t, "admin server did not exit before shutdown deadline")

		return nil
	}
}

// TestSendLivenessCheck_Success tests the success path of SendLivenessCheck.
// Covers: return result, nil (the non-error path in SendLivenessCheck).
// Sequential: starts a real HTTPS server (port reuse issues).
func TestSendLivenessCheck_Success(t *testing.T) {
	// NOT parallel: starts admin server with dynamic port.
	server, cancel, startErr := startAdminServer(t)

	defer func() {
		cancel()
		require.ErrorIs(t, waitForListenerServerExit(t, startErr), context.Canceled)
	}()

	settings := cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings()
	settings.BindPrivatePort = uint16(server.ActualPort()) //nolint:gosec // Port validated by server.
	settings.DevMode = true                                // Enable InsecureSkipVerify for self-signed cert.

	result, err := cryptoutilAppsFrameworkServiceServerListener.SendLivenessCheck(settings)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestSendReadinessCheck_Success tests the success path of SendReadinessCheck.
// Covers: return result, nil (the non-error path in SendReadinessCheck).
// Sequential: starts a real HTTPS server (port reuse issues).
func TestSendReadinessCheck_Success(t *testing.T) {
	// NOT parallel: starts admin server with dynamic port.
	server, cancel, startErr := startAdminServer(t)

	defer func() {
		cancel()
		require.ErrorIs(t, waitForListenerServerExit(t, startErr), context.Canceled)
	}()

	// Mark server ready so readyz returns 200.
	server.SetReady(true)

	settings := cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings()
	settings.BindPrivatePort = uint16(server.ActualPort()) //nolint:gosec // Port validated by server.
	settings.DevMode = true

	result, err := cryptoutilAppsFrameworkServiceServerListener.SendReadinessCheck(settings)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestSendShutdownRequest_Success tests the success path of SendShutdownRequest.
// Covers: return nil (the non-error path in SendShutdownRequest).
// The endpoint must stop the listener without context cancellation.
// Sequential: starts a real HTTPS server (port reuse issues).
func TestSendShutdownRequest_Success(t *testing.T) {
	// NOT parallel: starts admin server with dynamic port.
	server, cancel, startErr := startAdminServer(t)
	defer cancel()

	settings := cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings()
	settings.BindPrivatePort = uint16(server.ActualPort()) //nolint:gosec // Port validated by server.
	settings.DevMode = true

	err := cryptoutilAppsFrameworkServiceServerListener.SendShutdownRequest(settings)
	require.NoError(t, err)

	require.NoError(t, waitForListenerServerExit(t, startErr))
}
