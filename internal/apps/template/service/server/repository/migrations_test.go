// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"database/sql"
	"embed"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

//go:embed test_migrations/*.sql
var testMigrationsFS embed.FS

// TestNewMigrationRunner tests migration runner creation.
func TestNewMigrationRunner(t *testing.T) {
	t.Parallel()

	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(testMigrationsFS, "test_migrations")

	require.NotNil(t, runner)
}

// TestMigrationRunner_Apply_SQLite tests applying migrations to SQLite.
func TestMigrationRunner_Apply_SQLite(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(testMigrationsFS, "test_migrations")
	err = runner.Apply(db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)

	require.NoError(t, err)

	// Verify table was created.
	var count int

	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM test_table").Scan(&count)
	require.NoError(t, err)
}

// TestMigrationRunner_Apply_InvalidPath tests migration with invalid path.
func TestMigrationRunner_Apply_InvalidPath(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(testMigrationsFS, "nonexistent")
	err = runner.Apply(db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create iofs source driver")
}

// TestMigrationRunner_Apply_UnsupportedDatabaseType tests unsupported database type.
func TestMigrationRunner_Apply_UnsupportedDatabaseType(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(testMigrationsFS, "test_migrations")
	err = runner.Apply(db, cryptoutilAppsTemplateServiceServerRepository.DatabaseType("unsupported"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}

// TestMigrationRunner_Apply_NoChanges tests applying migrations when already up-to-date.
func TestMigrationRunner_Apply_NoChanges(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	runner := cryptoutilAppsTemplateServiceServerRepository.NewMigrationRunner(testMigrationsFS, "test_migrations")

	// Apply migrations first time.
	err = runner.Apply(db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)

	// Apply again - should succeed with no changes.
	err = runner.Apply(db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)
}

// TestApplyMigrationsFromFS_SQLite tests convenience function with SQLite.
func TestApplyMigrationsFromFS_SQLite(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(db, testMigrationsFS, "test_migrations", cryptoutilSharedMagic.TestDatabaseSQLite)
	require.NoError(t, err)
}

// TestApplyMigrationsFromFS_Sqlite3 tests convenience function with sqlite3 alias.
func TestApplyMigrationsFromFS_Sqlite3(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(db, testMigrationsFS, "test_migrations", "sqlite3")
	require.NoError(t, err)
}

// TestApplyMigrationsFromFS_UnsupportedType tests unsupported database type.
func TestApplyMigrationsFromFS_UnsupportedType(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(db, testMigrationsFS, "test_migrations", "mysql")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}
