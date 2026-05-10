//go:build integration
// +build integration

// Copyright (c) 2025-2026 Justin Cranford.
//
// Unified TestMain for orm package integration tests.
//

package orm

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerApplication "cryptoutil/internal/apps-framework/service/server/application"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testTemplateCore     *cryptoutilAppsFrameworkServiceServerApplication.Core
	testOrmRepository    *OrmRepository
	skipReadOnlyTxTests  = true // true for DBTypeSQLite, false for DBTypePostgres
	numMaterialKeys      = cryptoutilSharedMagic.JoseJADefaultMaxMaterials
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		var err error

		// Initialize shared test fixture using SQLite for integration testing
		// This runs ONCE for all tests in this package
		settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("orm-integration-tests")
		settings.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

		testTemplateCore, err = cryptoutilAppsFrameworkServiceServerApplication.StartCore(testCtx, settings)
		if err != nil {
			panic(fmt.Sprintf("failed to start template core: %v", err))
		}

		defer func() {
			if testTemplateCore.ShutdownDBContainer != nil {
				testTemplateCore.ShutdownDBContainer()
			}

			testTemplateCore.Basic.Shutdown()
		}()

		testTelemetryService = testTemplateCore.Basic.TelemetryService
		testJWKGenService = testTemplateCore.Basic.JWKGenService

		sqlDB, err := testTemplateCore.DB.DB()
		if err != nil {
			panic(fmt.Sprintf("failed to get sql.DB from GORM: %v", err))
		}

		err = cryptoutilAppsFrameworkServiceServerRepository.ApplyMigrationsFromFS(
			sqlDB,
			cryptoutilAppsFrameworkServiceServerRepository.MigrationsFS,
			"migrations",
			cryptoutilSharedMagic.TestDatabaseSQLite,
		)
		if err != nil {
			panic(fmt.Sprintf("failed to apply template migrations: %v", err))
		}

		err = testTemplateCore.DB.AutoMigrate(&ElasticKey{}, &MaterialKey{})
		if err != nil {
			panic(fmt.Sprintf("failed to apply KMS domain tables: %v", err))
		}

		testOrmRepository = RequireNewForTest(testCtx, testTelemetryService, testTemplateCore.DB, testJWKGenService, settings.VerboseMode)
		defer testOrmRepository.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
