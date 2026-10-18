// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerListener "cryptoutil/internal/apps-framework/service/server/listener"
	cryptoutilAppsFrameworkServiceTestHelpBootstrap "cryptoutil/internal/apps-framework/service/test_help_bootstrap"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestTemplateServer_PortConflict verifies public and admin listeners reject occupied dynamic loopback ports.
func TestTemplateServer_PortConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		admin     bool
		wantError string
	}{
		{name: "public listener", wantError: "public server failed: failed to create listener"},
		{name: "admin listener", admin: true, wantError: "admin server failed: failed to create admin listener"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			occupiedListener, err := net.Listen("tcp", net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, "0"))

			require.NoError(t, err)
			defer func() {
				require.NoError(t, occupiedListener.Close())
			}()

			tcpAddr, ok := occupiedListener.Addr().(*net.TCPAddr)
			require.True(t, ok)
			require.Greater(t, tcpAddr.Port, 0)

			occupiedPort := uint16(tcpAddr.Port) //nolint:gosec // Port range is guaranteed by the operating system.

			settings := cryptoutilAppsFrameworkServiceTestHelpBootstrap.NewTestServerSettings(t)
			if tc.admin {
				settings.BindPrivatePort = occupiedPort
			} else {
				settings.BindPublicPort = occupiedPort
			}

			httpsServers, err := cryptoutilAppsFrameworkServiceServerListener.NewHTTPServers(t.Context(), settings)
			require.NoError(t, err)

			app, err := cryptoutilAppsFrameworkServiceServer.NewApplication(t.Context(), httpsServers.PublicServer, httpsServers.AdminServer)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(t.Context(), cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
			defer cancel()

			err = app.Start(ctx)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}
