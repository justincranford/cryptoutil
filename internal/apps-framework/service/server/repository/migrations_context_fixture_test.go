// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"database/sql"
	"sync"

	dbhelp "cryptoutil/internal/apps-framework/service/test_help_db"
)

var (
	contextMigrationDB *sql.DB
	contextMigrationMu sync.Mutex
)

// InitContextMigrationFixture initializes the shared cancellation test database.
func InitContextMigrationFixture() (func() error, error) {
	db, cleanup, err := dbhelp.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		return nil, err
	}

	contextMigrationDB, err = db.DB()
	if err != nil {
		_ = cleanup()

		return nil, err
	}

	contextMigrationDB.SetMaxOpenConns(1)

	return cleanup, nil
}
