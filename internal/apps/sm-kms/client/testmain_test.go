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

	googleUuid "github.com/google/uuid"
)

var (
	testIntegrationServer *cryptoutilAppsFrameworkServiceTestOrcIntegration.IntegrationServer
	testRootCAsPool       *x509.CertPool
	testServerPublicURL   string
	testBrowserToken      string
	testServiceToken      string
)

// configureLightweightJWKPoolsForIntegrationTests reduces expensive keygen pool
// concurrency in this package's TestMain process so shutdown does not stall on
// long-running RSA workers.
func configureLightweightJWKPoolsForIntegrationTests() {
	cryptoutilSharedMagic.DefaultPoolConfigRSA4096 = cryptoutilSharedMagic.DefaultPoolConfig{NumWorkers: 1, MaxSize: 1}
	cryptoutilSharedMagic.DefaultPoolConfigRSA3072 = cryptoutilSharedMagic.DefaultPoolConfig{NumWorkers: 1, MaxSize: 1}
	cryptoutilSharedMagic.DefaultPoolConfigRSA2048 = cryptoutilSharedMagic.DefaultPoolConfig{NumWorkers: 1, MaxSize: 1}
}

func bootstrapTokens(testServer *cryptoutilKmsServer.KMSServer) (string, string, error) {
	resources := testServer.Resources()
	if resources == nil || resources.SessionManager == nil {
		return "", "", fmt.Errorf("session manager not initialized")
	}

	// Create ONE shared tenant/realm for both tokens to ensure test requests have consistent context
	sharedTenantID := googleUuid.New()
	sharedRealmID := googleUuid.New()

	userID := googleUuid.New().String()
	browserToken, err := resources.SessionManager.IssueBrowserSessionWithTenant(context.Background(), userID, sharedTenantID, sharedRealmID)
	if err != nil {
		return "", "", fmt.Errorf("issue browser session token: %w", err)
	}

	if browserToken == "" {
		return "", "", fmt.Errorf("issued empty browser session token")
	}

	clientID := googleUuid.New().String()
	serviceToken, err := resources.SessionManager.IssueServiceSessionWithTenant(context.Background(), clientID, sharedTenantID, sharedRealmID)
	if err != nil {
		return "", "", fmt.Errorf("issue service session token: %w", err)
	}

	if serviceToken == "" {
		return "", "", fmt.Errorf("issued empty service session token")
	}

	return browserToken, serviceToken, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	configureLightweightJWKPoolsForIntegrationTests()

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
	testServerPublicURL = testIntegrationServer.PublicBaseURL()

	testBrowserToken, testServiceToken, err = bootstrapTokens(testServer)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to bootstrap tokens: %v", err))
	}

	exitCode := m.Run()

	// Cleanup
	if err := testIntegrationServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to shutdown integration server: %v\n", err)
	}

	os.Exit(exitCode)
}
