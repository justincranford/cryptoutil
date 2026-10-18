// Copyright (c) 2025-2026 Justin Cranford.
//

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Sequential: resets and inspects the package migration cache.
func TestMigrate_PersistentDatabaseAndCachedRepeat(t *testing.T) {
	memoryDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = memoryDB.Close()
	})

	require.NoError(t, Migrate(context.Background(), memoryDB, dbTypeSQLite))
	require.NoError(t, Migrate(context.Background(), memoryDB, dbTypeSQLite))

	dbPath := filepath.Join(t.TempDir(), "identity.db")
	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, Migrate(context.Background(), db, dbTypeSQLite))
	require.NoError(t, Migrate(context.Background(), db, dbTypeSQLite))
}

func TestMigrate_InvalidDatabaseAndCancelledContext(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	err = Migrate(context.Background(), db, testUnsupportedDatabaseType)
	require.ErrorContains(t, err, testFailedToCreatePostgresDriverError)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = Migrate(ctx, db, dbTypeSQLite)
	require.ErrorIs(t, err, context.Canceled)

	require.NoError(t, db.Close())

	closedDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	require.NoError(t, closedDB.Close())
	require.ErrorContains(t, Migrate(context.Background(), closedDB, testUnsupportedDatabaseType), "failed to create schema_migrations table")
}

func TestMigrate_FailedSQLMigration(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	_, err = db.Exec("CREATE TABLE users (id TEXT PRIMARY KEY)")
	require.NoError(t, err)
	require.ErrorContains(t, Migrate(context.Background(), db, dbTypeSQLite), "failed to apply migrations")
}

func TestMigrate_DatabaseDriverErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		database string
		wantErr  string
	}{
		{name: "SQLite", database: dbTypeSQLite, wantErr: "failed to apply migrations"},
		{name: "PostgreSQL", database: testUnsupportedDatabaseType, wantErr: testFailedToCreatePostgresDriverError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := sql.OpenDB(closeErrorConnector{})

			t.Cleanup(func() {
				_ = db.Close()
			})

			require.ErrorContains(t, Migrate(context.Background(), db, tc.database), tc.wantErr)
		})
	}
}

func TestMigrate_DriverInitializationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dbType  string
		wantErr string
	}{
		{name: "SQLite ping failure", dbType: dbTypeSQLite, wantErr: "failed to create sqlite driver"},
		{name: "PostgreSQL ping failure", dbType: testUnsupportedDatabaseType, wantErr: testFailedToCreatePostgresDriverError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := sql.OpenDB(migrationPingErrorConnector{})

			t.Cleanup(func() {
				_ = db.Close()
			})

			require.ErrorContains(t, Migrate(context.Background(), db, tc.dbType), tc.wantErr)
		})
	}
}

// Sequential: replaces the embedded migrations filesystem to test source-driver initialization errors.
func TestMigrate_InvalidEmbeddedFilesystem(t *testing.T) {
	originalFS := migrationsFS

	migrationsFS = embed.FS{}
	defer func() {
		migrationsFS = originalFS
	}()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.ErrorContains(t, Migrate(context.Background(), db, dbTypeSQLite), "failed to create iofs source driver")
}

type migrationPingErrorConnector struct{}

func (migrationPingErrorConnector) Connect(context.Context) (driver.Conn, error) {
	return migrationPingErrorConnection{}, nil
}

func (migrationPingErrorConnector) Driver() driver.Driver {
	return migrationPingErrorDriver{}
}

type migrationPingErrorDriver struct{}

func (migrationPingErrorDriver) Open(string) (driver.Conn, error) {
	return migrationPingErrorConnection{}, nil
}

type migrationPingErrorConnection struct{}

func (migrationPingErrorConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is unsupported")
}

func (migrationPingErrorConnection) Close() error {
	return nil
}

func (migrationPingErrorConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported")
}

func (migrationPingErrorConnection) Ping(context.Context) error {
	return errors.New("injected ping failure")
}

func (migrationPingErrorConnection) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (migrationPingErrorConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("injected query failure")
}
