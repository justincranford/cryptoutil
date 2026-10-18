// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestNewHTTPServers_StartAndShutdown(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	servers, err := NewHTTPServers(context.Background(), settings)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	startErrors := make(chan error, 2)

	go func() {
		startErrors <- servers.PublicServer.Start(ctx)
	}()
	go func() {
		startErrors <- servers.AdminServer.Start(ctx)
	}()

	require.Eventually(t, func() bool {
		return servers.PublicServer.ActualPort() > 0 && servers.AdminServer.ActualPort() > 0
	}, cryptoutilSharedMagic.TestTimeoutCryptoutilReady, cryptoutilSharedMagic.TestTimeoutHTTPRetryInterval)

	cancel()

	for range 2 {
		select {
		case err := <-startErrors:
			require.ErrorIs(t, err, context.Canceled)
		case <-time.After(cryptoutilSharedMagic.TestDefaultServerShutdownTimeout):
			require.FailNow(t, "HTTP server did not stop after context cancellation")
		}
	}
}
