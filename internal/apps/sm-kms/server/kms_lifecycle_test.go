//go:build !integration

// Copyright (c) 2025-2026 Justin Cranford.
//

package server

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceTestOrchIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestKMSServer_Lifecycle verifies server startup and graceful shutdown.
func TestKMSServer_Lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	databaseID, err := googleUuid.NewV7()
	require.NoError(t, err)

	cfg.DatabaseURL = fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String())

	server, err := NewKMSServerFromConfig(ctx, cfg)
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
