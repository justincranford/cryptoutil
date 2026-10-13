//go:build !integration

// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceTestOrchIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestRSServer_StartupShutdown verifies server startup and graceful shutdown.
func TestRSServer_StartupShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := cryptoutilAppsIdentityRsServerConfig.DefaultTestConfig()

	server, err := NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	integrationServer, err := cryptoutilAppsFrameworkServiceTestOrchIntegration.StartIntegrationServer(ctx, t, server, nil)
	require.NoError(t, err)

	require.Greater(t, server.PublicPort(), 0, "public port should be bound")
	require.Greater(t, server.AdminPort(), 0, "admin port should be bound")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
	defer cancel()

	err = integrationServer.Shutdown(shutdownCtx)
	require.NoError(t, err)
}
