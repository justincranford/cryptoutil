// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-KMS server tests.
package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceTestHelperApi "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilTestDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsFrameworkServiceTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testIntegrationServer *cryptoutilAppsFrameworkServiceTestOrcIntegration.IntegrationServer
	testHealthClient      *cryptoutilAppsFrameworkServiceTestHelperApi.HealthClient
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	cfg.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

	var err error

	// Create test server
	testServer, err := NewKMSServerFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create KMS server: %v", err))
	}

	testDB, dbCleanup, err := cryptoutilTestDb.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create test database: %v", err))
	}
	defer dbCleanup()

	// Start integration server and wait for ports
	testIntegrationServer, err = cryptoutilAppsFrameworkServiceTestOrcIntegration.StartIntegrationServerForTestMain(
		ctx,
		testServer,
		testDB,
	)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to start integration server: %v", err))
	}

	// Mark server ready
	testServer.SetReady(true)

	// Create health client for use in tests
	testHealthClient = cryptoutilAppsFrameworkServiceTestHelperApi.NewHealthClient(
		testIntegrationServer.PublicBaseURL(),
		testIntegrationServer.AdminBaseURL(),
	)

	exitCode := m.Run()

	// Cleanup
	if err := testIntegrationServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to shutdown integration server: %v\n", err)
	}

	os.Exit(exitCode)
}
