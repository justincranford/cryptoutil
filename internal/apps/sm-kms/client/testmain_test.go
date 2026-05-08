// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-KMS client tests.

package client

import (
	"context"
	"crypto/x509"
	"log"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps-framework/service/testing/e2e_helpers"
	cryptoutilKmsServer "cryptoutil/internal/apps/sm-kms/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testSettings         = cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("application_test")
	testServerPublicURL  string
	testServerPrivateURL string
	testRootCAsPool      *x509.CertPool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	testSettings.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN // SQLite in-memory for fast tests.

	testServer, err := cryptoutilKmsServer.NewKMSServerFromConfig(ctx, testSettings)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testServer, func() error {
		return testServer.Start(ctx)
	})

	defer func() {
		_ = testServer.Shutdown(ctx)
	}()

	testServerPublicURL = testServer.PublicBaseURL()
	testServerPrivateURL = testServer.AdminBaseURL()
	testRootCAsPool = testServer.TLSRootCAPool()

	os.Exit(m.Run())
}
