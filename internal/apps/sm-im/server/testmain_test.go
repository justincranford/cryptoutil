// Copyright (c) 2025-2026 Justin Cranford.
//

package server_test

import (
	"context"
	"os"
	"testing"
	"time"

	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsFrameworkServiceTestutil "cryptoutil/internal/apps-framework/service/testutil"
	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm-im/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testSmIMServer        *cryptoutilAppsSmImServer.SmIMServer
	testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer
	baseURL               string
	adminURL              string
)

var (
	testMockServerOK     = cryptoutilAppsFrameworkServiceTestutil.NewMockServerOK()
	testMockServerError  = cryptoutilAppsFrameworkServiceTestutil.NewMockServerError()
	testMockServerCustom = cryptoutilAppsFrameworkServiceTestutil.NewMockServerCustom()
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsSmImServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testSmIMServer, err = cryptoutilAppsSmImServer.NewIMServerFromConfig(ctx, cfg)
	if err != nil {
		panic("TestMain: failed to create server: " + err.Error())
	}

	// Start server and wait for both ports to bind.
	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(ctx, testSmIMServer, nil)
	if err != nil {
		panic("TestMain: failed to start server: " + err.Error())
	}

	// Store base URLs for tests.
	baseURL = testSmIMServer.PublicBaseURL()
	adminURL = testSmIMServer.AdminBaseURL()

	defer testMockServerOK.Close()
	defer testMockServerError.Close()
	defer testMockServerCustom.Close()

	// Record start time for benchmark.
	startTime := time.Now().UTC()

	// Run all tests (defer statements will execute cleanup AFTER m.Run() completes).
	exitCode := m.Run()

	elapsed := time.Since(startTime)

	// Log timing for comparison (visible in test output).
	// IMPORTANT: This timing includes TestMain setup overhead, which is amortized across all tests.
	// Individual test functions no longer pay setup cost - they reuse shared resources.
	println("TestMain: All tests completed in", elapsed.String())

	// Cleanup: Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testIntegrationServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
