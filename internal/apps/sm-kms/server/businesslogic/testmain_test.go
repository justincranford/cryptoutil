// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for SM-KMS businesslogic tests.

package businesslogic

import (
	"context"
	"database/sql"
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerApplication "cryptoutil/internal/apps-framework/service/server/application"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilKmsServerRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testCore *cryptoutilAppsFrameworkServiceServerApplication.Core
	testDB   *sql.DB
)

func TestMain(m *testing.M) {
	_ = os.Setenv("CRYPTOUTIL_DATABASE_URL", cryptoutilSharedMagic.SQLiteInMemoryDSN) //nolint:errcheck // TestMain cannot use t.Setenv

	ctx := os.Getenv("CRYPTOUTIL_DATABASE_URL")
	if ctx == "" {
		ctx = cryptoutilSharedMagic.SQLiteInMemoryDSN
	}

	// Initialize shared test fixture (Core + DB + migrations)
	// This runs ONCE for all tests in this package
	settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("businesslogic-shared")
	settings.DatabaseURL = ctx

	var err error

	testCore, err = cryptoutilAppsFrameworkServiceServerApplication.StartCore(context.Background(), settings)
	if err != nil {
		panic("TestMain: failed to start core: " + err.Error())
	}

	testDB, err = testCore.DB.DB()
	if err != nil {
		panic("TestMain: failed to get database: " + err.Error())
	}

	// Apply migrations (framework + domain)
	mergedFS := &testMergedMigrations{
		templateFS:   cryptoutilAppsFrameworkServiceServerRepository.MigrationsFS,
		templatePath: "migrations",
		domainFS:     cryptoutilKmsServerRepository.MigrationsFS,
		domainPath:   "migrations",
	}

	err = cryptoutilAppsFrameworkServiceServerRepository.ApplyMigrationsFromFS(
		testDB, mergedFS, "", cryptoutilSharedMagic.TestDatabaseSQLite,
	)
	if err != nil {
		panic("TestMain: failed to apply migrations: " + err.Error())
	}

	exitCode := m.Run()

	// Cleanup
	testCore.Shutdown()

	os.Exit(exitCode)
}
