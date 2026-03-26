// Copyright (c) 2025 Justin Cranford
//
// Tests for RunHealthContracts and RunReadyzNotReadyContract.
package contract

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsSkeletonTemplateServer "cryptoutil/internal/apps/skeleton-template/server"
	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps/framework/service/testing/e2e_helpers"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestRunHealthContracts(t *testing.T) {
	t.Parallel()

	RunHealthContracts(t, testContractServer)
}

// TestRunReadyzNotReadyContract tests that readyz returns 503 when server is not ready.
// Uses a dedicated fresh server that starts NOT ready to avoid state pollution
// with the shared testContractServer used by other parallel tests.
func TestRunReadyzNotReadyContract(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := cryptoutilAppsSkeletonTemplateServerConfig.DefaultTestConfig()

	srv, err := cryptoutilAppsSkeletonTemplateServer.NewFromConfig(ctx, cfg)
	require.NoError(t, err, "failed to create dedicated not-ready server")

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(srv, func() error {
		return srv.Start(ctx)
	})

	// Intentionally NOT calling srv.SetReady(true) — server starts not ready.
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
		defer cancel()

		_ = srv.Shutdown(shutdownCtx)
	})

	RunReadyzNotReadyContract(t, srv)
}
