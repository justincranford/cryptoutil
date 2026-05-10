// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-KMS client tests.

package client

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilTestDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsFrameworkServiceTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilKmsServer "cryptoutil/internal/apps/sm-kms/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testIntegrationServer *cryptoutilAppsFrameworkServiceTestOrcIntegration.IntegrationServer
	testRootCAsPool       *x509.CertPool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("application_test")
	cfg.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

	// Create test server
	testServer, err := cryptoutilKmsServer.NewKMSServerFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
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

	// Extract TLS root CA pool from server
	testRootCAsPool = testServer.TLSRootCAPool()

	exitCode := m.Run()

	// Cleanup
	if err := testIntegrationServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to shutdown integration server: %v\n", err)
	}

	os.Exit(exitCode)
}
