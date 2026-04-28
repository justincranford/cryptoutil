//go:build !integration

// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps-framework/service/testing/e2e_helpers"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestAuthzServer_StartupShutdown verifies independent server startup and graceful shutdown.
func TestAuthzServer_StartupShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := cryptoutilAppsIdentityAuthzServerConfig.DefaultTestConfig()

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
