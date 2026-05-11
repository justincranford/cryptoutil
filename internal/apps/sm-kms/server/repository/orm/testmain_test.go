// Copyright (c) 2025-2026 Justin Cranford.
//
// Unified TestMain for orm package integration tests.
// No //go:build directive: TestMain must compile in all build modes so go test can
// discover it alongside the integration-tagged test functions.

package orm

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testOrmRepository    *OrmRepository
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		var err error

		// Initialize shared test fixture using SQLite for integration testing.
		// This runs ONCE for all tests in this package.
		testDB, dbCleanup, err := cryptoutilAppsFrameworkServiceTestHelpDb.NewInMemorySQLiteDBForTestMain()
		if err != nil {
			panic(fmt.Sprintf("failed to create test DB: %v", err))
		}
		defer dbCleanup()

		settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("orm-integration-tests")
		settings.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

		testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(testCtx, settings.ToTelemetrySettings())
		if err != nil {
			panic(fmt.Sprintf("failed to create telemetry service: %v", err))
		}
		defer testTelemetryService.Shutdown()

		testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(testCtx, testTelemetryService, false)
		if err != nil {
			panic(fmt.Sprintf("failed to create JWK service: %v", err))
		}
		defer testJWKGenService.Shutdown()

		if err := testDB.AutoMigrate(&ElasticKey{}, &MaterialKey{}); err != nil {
			panic(fmt.Sprintf("failed to apply KMS domain tables: %v", err))
		}

		testOrmRepository = RequireNewForTest(testCtx, testTelemetryService, testDB, testJWKGenService, settings.VerboseMode)
		defer testOrmRepository.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
