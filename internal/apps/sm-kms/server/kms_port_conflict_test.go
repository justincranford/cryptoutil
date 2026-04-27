//go:build !integration

// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps-framework/service/testing/e2e_helpers"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestKMSServer_PortConflict verifies that two servers can start on distinct dynamic ports.
func TestKMSServer_PortConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cfg1 := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	cfg1.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN
	cfg2 := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	cfg2.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

	server1, err := NewKMSServer(ctx, cfg1)
	require.NoError(t, err)
	require.NotNil(t, server1)

	server2, err := NewKMSServer(ctx, cfg2)
	require.NoError(t, err)
	require.NotNil(t, server2)

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(server1, func() error {
		return server1.Start(ctx)
	})

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(server2, func() error {
		return server2.Start(ctx)
	})

	require.Greater(t, server1.PublicPort(), 0)
	require.Greater(t, server2.PublicPort(), 0)
	require.NotEqual(t, server1.PublicPort(), server2.PublicPort(), "servers should use distinct public ports")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer cancel()

	_ = server1.Shutdown(shutdownCtx)
	_ = server2.Shutdown(shutdownCtx)
}
