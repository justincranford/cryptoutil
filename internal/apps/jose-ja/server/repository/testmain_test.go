// Copyright (c) 2025-2026 Justin Cranford.
//

package repository

import (
	"database/sql"
	"os"
	"testing"

	"gorm.io/gorm"

	cryptoutilTestDb "cryptoutil/internal/apps-framework/service/test_help_db"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	var (
		dbCleanup func()
		err       error
	)

	testDB, dbCleanup, err = cryptoutilTestDb.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		panic("TestMain: failed to create test DB: " + err.Error())
	}
	defer dbCleanup()

	// Run migrations using underlying sql.DB.
	testSQLDB, err := testDB.DB()
	if err != nil {
		panic("TestMain: failed to get sql.DB: " + err.Error())
	}

	if err := ApplyJoseJAMigrations(testSQLDB, DatabaseTypeSQLite); err != nil {
		panic("TestMain: failed to run migrations: " + err.Error())
	}

	// Run all tests.
	os.Exit(m.Run())
}

// newClosedDB creates a closed SQLite DB using the shared test_help_db helper.
// Used by error-path tests to force database errors.
func newClosedDB(t *testing.T) *gorm.DB {
	t.Helper()

	return cryptoutilTestDb.NewClosedSQLiteDB(t, func(sqlDB *sql.DB) error {
		return ApplyJoseJAMigrations(sqlDB, DatabaseTypeSQLite)
	})
}
