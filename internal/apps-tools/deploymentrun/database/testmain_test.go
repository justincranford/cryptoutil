// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"database/sql"
	"os"
	"sync"
	"testing"

	dbhelp "cryptoutil/internal/apps-framework/service/test_help_db"
)

var (
	sharedSQLite *sql.DB
	sqliteMu     sync.Mutex
)

func TestMain(m *testing.M) {
	gormDB, cleanup, err := dbhelp.NewInMemorySQLiteDBForTestMain()
	if err != nil {
		panic(err)
	}

	sharedSQLite, err = gormDB.DB()
	if err != nil {
		panic(err)
	}

	sharedSQLite.SetMaxOpenConns(1)

	code := m.Run()
	if err := cleanup(); err != nil {
		code = 1
	}

	os.Exit(code)
}
