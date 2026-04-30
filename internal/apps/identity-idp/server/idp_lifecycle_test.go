//go:build !integration

// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps-framework/service/testing/e2e_helpers"
	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity-idp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestIDPServer_StartupShutdown verifies server startup and graceful shutdown.
func TestIDPServer_StartupShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := cryptoutilAppsIdentityIdpServerConfig.DefaultTestConfig()

	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(server, func() error {
		return server.Start(ctx)
	})

	server.SetReady(true)

	require.Greater(t, server.PublicPort(), 0, "public port should be bound")
	require.Greater(t, server.AdminPort(), 0, "admin port should be bound")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	require.NoError(t, err)
}
