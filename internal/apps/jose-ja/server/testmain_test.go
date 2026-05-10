// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for JOSE-JA server integration tests.
package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceTestHelperApi "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilTestDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsFrameworkServiceTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose-ja/server/config"
)

var (
	testServer            *JoseJAServer
	testIntegrationServer *cryptoutilAppsFrameworkServiceTestOrcIntegration.IntegrationServer
	testHealthClient      *cryptoutilAppsFrameworkServiceTestHelperApi.HealthClient
	testPublicBaseURL     string
	testAdminBaseURL      string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsJoseJaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
	}

	testDB, dbCleanup, err := cryptoutilTestDb.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create test database: %v", err))
	}
	defer dbCleanup()

	// Start integration server and wait for ports.
	testIntegrationServer, err = cryptoutilAppsFrameworkServiceTestOrcIntegration.StartIntegrationServerForTestMain(
		ctx,
		testServer,
		testDB,
	)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to start integration server: %v", err))
	}

	// Create health client for use in tests.
	testPublicBaseURL = testIntegrationServer.PublicBaseURL()
	testAdminBaseURL = testIntegrationServer.AdminBaseURL()
	testHealthClient = cryptoutilAppsFrameworkServiceTestHelperApi.NewHealthClient(
		testPublicBaseURL,
		testAdminBaseURL,
	)

	// Run all tests.
	exitCode := m.Run()

	// Cleanup.
	if err := testIntegrationServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to shutdown integration server: %v\n", err)
	}

	os.Exit(exitCode)
}
