//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for __PS_ID__ server integration tests.
// No //go:build directive: this file compiles for all build variants (unit, integration, e2e).

package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilTestHelpApi "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilTestDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutil__SERVICE__ServerConfig "cryptoutil/internal/apps/__PS_ID__/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer            *__SERVICE__Server
	testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer
	testHealthClient      *cryptoutilTestHelpApi.HealthClient
	testPublicBaseURL     string
	testAdminBaseURL      string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutil__SERVICE__ServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

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

	// Start integration server and wait for dual ports.
	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(
		ctx,
		testServer,
		testDB,
	)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to start integration server: %v", err))
	}

	// Expose resolved URLs for use in tests.
	testPublicBaseURL = testIntegrationServer.PublicBaseURL()
	testAdminBaseURL = testIntegrationServer.AdminBaseURL()
	testHealthClient = cryptoutilTestHelpApi.NewHealthClient(testPublicBaseURL, testAdminBaseURL)

	exitCode := m.Run()

	if err := testIntegrationServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: shutdown error: %v\n", err)
	}

	os.Exit(exitCode)
}
