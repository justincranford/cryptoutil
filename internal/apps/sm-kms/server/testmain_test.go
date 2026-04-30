//go:build !integration

// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-KMS server unit tests.
package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps-framework/service/testing/e2e_helpers"
	cryptoutilTestingHealthclient "cryptoutil/internal/apps-framework/service/testing/healthclient"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer       *KMSServer
	testHealthClient *cryptoutilTestingHealthclient.HealthClient
	testPublicURL    string
	testAdminURL     string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	cfg.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

	var err error

	testServer, err = NewKMSServer(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create KMS server: %v", err))
	}

	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testServer, func() error {
		return testServer.Start(ctx)
	})

	testServer.SetReady(true)

	testPublicURL, testAdminURL = cryptoutilAppsFrameworkServiceTestingE2eHelpers.DualPortBaseURLs(testServer)
	testHealthClient = cryptoutilTestingHealthclient.NewHealthClient(testPublicURL, testAdminURL)

	exitCode := m.Run()

	_ = testServer.Shutdown(ctx)

	os.Exit(exitCode)
}
