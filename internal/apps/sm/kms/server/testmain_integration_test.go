//go:build integration

// Copyright (c) 2025 Justin Cranford
//
// TestMain for SM-KMS server integration tests.
package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps/framework/service/testing/e2e_helpers"
	cryptoutilTestingHealthclient "cryptoutil/internal/apps/framework/service/testing/healthclient"
)

var (
	testIntegrationServer       *KMSServer
	testIntegrationHealthClient *cryptoutilTestingHealthclient.HealthClient
	testIntegrationPublicURL    string
	testIntegrationAdminURL     string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration using template helper.
	cfg := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("kms-server-integration")

	// Create KMS server.
	var err error

	testIntegrationServer, err = NewKMSServer(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create KMS server: %v", err))
	}

	// Use generic template helper for goroutine start + dual port polling + panic-on-failure.
	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testIntegrationServer, func() error {
		return testIntegrationServer.Start(ctx)
	})

	// Store base URLs for tests.
	testIntegrationPublicURL, testIntegrationAdminURL = cryptoutilAppsFrameworkServiceTestingE2eHelpers.DualPortBaseURLs(testIntegrationServer)

	// Create shared health client.
	testIntegrationHealthClient = cryptoutilTestingHealthclient.NewHealthClient(testIntegrationPublicURL, testIntegrationAdminURL)

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Shutdown server.
	_ = testIntegrationServer.Shutdown(ctx)

	os.Exit(exitCode)
}
