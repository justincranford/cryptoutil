// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"context"
"database/sql"
"testing"
"testing/fstest"

"github.com/stretchr/testify/require"
)

// testSuccessMigrationsFS is an in-memory filesystem with a valid "migrations"
// directory that ApplyMigrations expects.
var testSuccessMigrationsFS = fstest.MapFS{
"migrations/001_init.up.sql": &fstest.MapFile{
Data: []byte("CREATE TABLE IF NOT EXISTS test_init_table (id INTEGER PRIMARY KEY, name TEXT NOT NULL);"),
},
"migrations/001_init.down.sql": &fstest.MapFile{
Data: []byte("DROP TABLE IF EXISTS test_init_table;"),
},
}

// TestInitSQLite_Success verifies that InitSQLite succeeds with a valid
// in-memory DSN using an in-memory filesystem with valid migrations.
func TestInitSQLite_Success(t *testing.T) {
t.Parallel()

ctx := context.Background()

db, err := InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, testSuccessMigrationsFS)
require.NoError(t, err)
require.NotNil(t, db)

sqlDB, err := db.DB()
require.NoError(t, err)

defer func() {
_ = sqlDB.Close()
}()
}

// TestInitSQLite_ApplyMigrationsError verifies that InitSQLite returns an error
// when the migrations FS has no "migrations" sub-path (iofs.New fails inside
// Apply, propagates as "failed to apply migrations").
func TestInitSQLite_ApplyMigrationsError(t *testing.T) {
t.Parallel()

ctx := context.Background()

// testDBMigrationsFS only has "test_migrations/" not "migrations/" so
// ApplyMigrations -> NewMigrationRunner(fs, "migrations") -> iofs.New fails.
db, err := InitSQLite(ctx, "file::memory:?cache=shared&_initErr=1", testDBMigrationsFS)
require.Error(t, err)
require.Nil(t, db)
require.Contains(t, err.Error(), "failed to apply migrations")
}

// TestMigrationRunner_Apply_PostgreSQLDriverError verifies that Apply returns
// an error when the pgx driver rejects a non-PostgreSQL connection (SQLite DB).
// This covers the DatabaseTypePostgreSQL code path in Apply.
func TestMigrationRunner_Apply_PostgreSQLDriverError(t *testing.T) {
t.Parallel()

db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
require.NoError(t, err)

defer func() {
_ = db.Close()
}()

runner := NewMigrationRunner(testDBMigrationsFS, "test_migrations")
err = runner.Apply(db, DatabaseTypePostgreSQL)
require.Error(t, err)
}

// TestMigrationRunner_Apply_SQLiteDriverError verifies that Apply returns an
// error when the SQLite driver receives a closed sql.DB.
// This covers the sqlite.WithInstance error path in Apply.
func TestMigrationRunner_Apply_SQLiteDriverError(t *testing.T) {
t.Parallel()

db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
require.NoError(t, err)
require.NoError(t, db.Close())

runner := NewMigrationRunner(testDBMigrationsFS, "test_migrations")
err = runner.Apply(db, DatabaseTypeSQLite)
require.Error(t, err)
}

// TestApplyMigrationsFromFS_PostgreSQL verifies the "postgres" branch of
// ApplyMigrationsFromFS assigns DatabaseTypePostgreSQL and propagates errors.
func TestApplyMigrationsFromFS_PostgreSQL(t *testing.T) {
t.Parallel()

db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
require.NoError(t, err)

defer func() {
_ = db.Close()
}()

err = ApplyMigrationsFromFS(db, testDBMigrationsFS, "test_migrations", cryptoutilSharedMagic.DockerServicePostgres)
require.Error(t, err)
}

// TestMigrationRunner_Apply_UpError verifies that Apply returns an error when
// the migration SQL itself fails to execute (m.Up fails).
func TestMigrationRunner_Apply_UpError(t *testing.T) {
t.Parallel()

// Create an in-memory FS with a valid filename but invalid SQL.
invalidSQLFS := fstest.MapFS{
"bad_migrations/000001_bad.up.sql": &fstest.MapFile{
Data: []byte("THIS IS NOT VALID SQL @@@@;"),
},
"bad_migrations/000001_bad.down.sql": &fstest.MapFile{
Data: []byte("DROP TABLE IF EXISTS whatever;"),
},
}

db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
require.NoError(t, err)

defer func() {
_ = db.Close()
}()

runner := NewMigrationRunner(invalidSQLFS, "bad_migrations")
err = runner.Apply(db, DatabaseTypeSQLite)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to apply migrations")
}
