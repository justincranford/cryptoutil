// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"os"
	"testing"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps-framework/service/server/testutil"
	cryptoutilTestHelpDb "cryptoutil/internal/apps-framework/service/test_help_db"
)

var testGormDB *gorm.DB

func TestMain(m *testing.M) {
	if err := cryptoutilAppsFrameworkServiceServerTestutil.Initialize(); err != nil {
		panic("failed to initialize test fixtures: " + err.Error())
	}

	var (
		dbCleanup func()
		dbErr     error
	)

	testGormDB, dbCleanup, dbErr = cryptoutilTestHelpDb.NewInMemorySQLiteDBForTestMain()
	if dbErr != nil {
		panic("failed to create test database: " + dbErr.Error())
	}

	defer dbCleanup()

	sqlDB, err := testGormDB.DB()
	if err != nil {
		panic("failed to get sql.DB for migrations: " + err.Error())
	}

	if migrateErr := cryptoutilAppsFrameworkServiceServerRepository.ApplyMigrations(sqlDB, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite, cryptoutilAppsFrameworkServiceServerRepository.MigrationsFS); migrateErr != nil {
		panic("failed to apply migrations: " + migrateErr.Error())
	}

	os.Exit(m.Run())
}
