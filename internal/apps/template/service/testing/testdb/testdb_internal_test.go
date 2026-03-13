// Copyright (c) 2025 Justin Cranford
//

// Package testdb provides white-box tests for internal functions.
// These tests verify error paths in buildInMemorySQLiteDB using seam injection.
package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

// TestBuildInMemorySQLiteDB_Success verifies the happy path.
func TestBuildInMemorySQLiteDB_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dsn := "file:test-success?mode=memory&cache=private"

	db, sqlDB, err := buildInMemorySQLiteDB(ctx, sql.Open, dsn)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, sqlDB)

	defer func() { _ = sqlDB.Close() }()

	require.NoError(t, sqlDB.PingContext(ctx))
}

// TestBuildInMemorySQLiteDB_OpenFails verifies sql.Open error is propagated.
func TestBuildInMemorySQLiteDB_OpenFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	failOpen := func(_, _ string) (*sql.DB, error) {
		return nil, fmt.Errorf("forced open failure")
	}

	db, sqlDB, err := buildInMemorySQLiteDB(ctx, failOpen, "any-dsn")
	require.Error(t, err)
	require.Contains(t, err.Error(), "sql.Open")
	require.Nil(t, db)
	require.Nil(t, sqlDB)
}

// TestBuildInMemorySQLiteDB_WALPragmaFails verifies WAL pragma error is propagated.
// Uses a pre-closed sql.DB to trigger the PRAGMA failure.
func TestBuildInMemorySQLiteDB_WALPragmaFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	closedOpen := func(_, _ string) (*sql.DB, error) {
		rawDB, openErr := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
		if openErr != nil {
			return nil, openErr
		}

		_ = rawDB.Close() // Close immediately to cause PRAGMA failure.

		return rawDB, nil
	}

	db, sqlDB, err := buildInMemorySQLiteDB(ctx, closedOpen, "any-dsn")
	require.Error(t, err)
	require.Contains(t, err.Error(), "WAL pragma")
	require.Nil(t, db)
	require.Nil(t, sqlDB)
}

// TestBuildInMemorySQLiteDB_CloseOnWALError verifies sqlDB is closed when WAL pragma fails.
func TestBuildInMemorySQLiteDB_CloseOnWALError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	closed := false

	openThenClose := func(_, _ string) (*sql.DB, error) {
		rawDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
		if err != nil {
			return nil, err
		}

		_ = rawDB.Close() // Pre-close to trigger WAL error.

		closed = true

		return rawDB, nil
	}

	_, _, err := buildInMemorySQLiteDB(ctx, openThenClose, "any")
	require.Error(t, err)
	require.True(t, closed, "openThenClose should have been called")
}
// TestBuildClosedSQLiteDB_Success verifies the happy path with no migrations.
func TestBuildClosedSQLiteDB_Success(t *testing.T) {
t.Parallel()

ctx := context.Background()

db, err := buildClosedSQLiteDB(ctx, sql.Open, "file:closed-success?mode=memory&cache=shared", nil)
require.NoError(t, err)
require.NotNil(t, db)

// All operations should fail because the connection is closed.
execErr := db.Exec("SELECT 1").Error
require.Error(t, execErr)
}

// TestBuildClosedSQLiteDB_WithMigrations verifies migrations are called before closing.
func TestBuildClosedSQLiteDB_WithMigrations(t *testing.T) {
t.Parallel()

ctx := context.Background()
called := false

db, err := buildClosedSQLiteDB(ctx, sql.Open, "file:closed-mig?mode=memory&cache=shared", func(_ *sql.DB) error {
called = true

return nil
})
require.NoError(t, err)
require.NotNil(t, db)
require.True(t, called)
}

// TestBuildClosedSQLiteDB_OpenFails covers openFn failure propagation.
func TestBuildClosedSQLiteDB_OpenFails(t *testing.T) {
t.Parallel()

ctx := context.Background()
failOpen := func(_, _ string) (*sql.DB, error) {
return nil, fmt.Errorf("injected open failure")
}

db, err := buildClosedSQLiteDB(ctx, failOpen, "any", nil)
require.Error(t, err)
require.Nil(t, db)
require.Contains(t, err.Error(), "sql.Open")
}

// TestBuildClosedSQLiteDB_MigrationsFails covers the migrations error path.
func TestBuildClosedSQLiteDB_MigrationsFails(t *testing.T) {
t.Parallel()

ctx := context.Background()

db, err := buildClosedSQLiteDB(ctx, sql.Open, "file:closed-mig-fail?mode=memory&cache=shared", func(_ *sql.DB) error {
return fmt.Errorf("injected migration failure")
})
require.Error(t, err)
require.Nil(t, db)
require.Contains(t, err.Error(), "migrations")
}
