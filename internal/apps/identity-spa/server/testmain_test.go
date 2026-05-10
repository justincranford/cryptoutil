// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"os"
	"testing"

	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer            *SPAServer
	testBaseURL           string
	testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsIdentitySpaServerConfig.DefaultTestConfig()

	var err error

	testServer, err = NewFromConfig(ctx, cfg)
	if err != nil {
		panic("TestMain: failed to create server: " + err.Error())
	}

	// Start server and wait for both ports to bind.
	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(ctx, testServer, nil)
	if err != nil {
		panic("TestMain: failed to start server: " + err.Error())
	}

	// Set base URL after server starts (uses dynamic port).
	testBaseURL = testServer.PublicBaseURL()

	// Run tests.
	exitCode := m.Run()

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testIntegrationServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}

// requireTestSetup checks that the test server is properly initialized.
func requireTestSetup(t *testing.T) {
	t.Helper()

	if testServer == nil {
		t.Fatal("Test server is nil")
	}
}
