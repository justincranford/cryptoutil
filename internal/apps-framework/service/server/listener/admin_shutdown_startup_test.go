// Copyright (c) 2025-2026 Justin Cranford.
//

package listener

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAdminServer_ShutdownDuringStartup(t *testing.T) {
	t.Parallel()

	listenEntered := make(chan struct{})
	releaseListen := make(chan struct{})

	var releaseOnce sync.Once

	release := func() {
		releaseOnce.Do(func() {
			close(releaseListen)
		})
	}
	defer release()

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	server, err := newAdminHTTPServerInternal(
		context.Background(),
		settings,
		validAutoTLSSettings(t),
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			close(listenEntered)

			select {
			case <-releaseListen:
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		func(_ *fiber.App, _ net.Listener) error { return nil },
		os.ReadFile,
	)
	require.NoError(t, err)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start(context.Background())
	}()

	select {
	case <-listenEntered:
	case <-time.After(cryptoutilSharedMagic.TestTimeoutCryptoutilReady):
		require.FailNow(t, "admin server did not reach listener startup")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
	defer cancel()

	require.NoError(t, server.Shutdown(shutdownCtx))

	release()

	select {
	case err := <-startErr:
		require.ErrorContains(t, err, "admin server already shutdown")
	case <-time.After(cryptoutilSharedMagic.TestDefaultServerShutdownTimeout):
		require.FailNow(t, "admin server Start did not finish after shutdown during startup")
	}
}
