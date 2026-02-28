// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"database/sql"
	"io/fs"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigrationsFS_Embedded(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(MigrationsFS, "migrations")
	require.NoError(t, err)
	require.NotEmpty(t, entries, "should have embedded migration files")

	// Verify expected migration files exist.
	expectedFiles := map[string]bool{
		"2001_ca_items.up.sql":   false,
		"2001_ca_items.down.sql": false,
	}

	for _, entry := range entries {
		if _, ok := expectedFiles[entry.Name()]; ok {
			expectedFiles[entry.Name()] = true
		}
	}

	for name, found := range expectedFiles {
		require.True(t, found, "expected migration file %s not found", name)
	}
}

func TestGetMergedMigrationsFS(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()
	require.NotNil(t, mergedFS)

	// Verify we can read the migrations directory from merged FS.
	readDirFS, ok := mergedFS.(fs.ReadDirFS)
	require.True(t, ok, "merged FS should implement fs.ReadDirFS")

	entries, err := readDirFS.ReadDir("migrations")
	require.NoError(t, err)
	require.NotEmpty(t, entries, "merged FS should have migration files")

	// Should contain both template (1001-1004) and pki-ca (2001+) migrations.
	hasTemplate := false
	hasPkiCa := false

	for _, entry := range entries {
		name := entry.Name()

		if len(name) >= 4 && name[:4] == "1001" {
			hasTemplate = true
		}

		if len(name) >= 4 && name[:4] == "2001" {
			hasPkiCa = true
		}
	}

	require.True(t, hasTemplate, "merged FS should contain template migrations (1001+)")
	require.True(t, hasPkiCa, "merged FS should contain pki-ca migrations (2001+)")
}

func TestMergedFS_Open(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Test opening pki-ca migration file.
	file, err := mergedFS.Open("migrations/2001_ca_items.up.sql")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	// Test opening template migration file (from fallback).
	file, err = mergedFS.Open("migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	// Test opening non-existent file.
	_, err = mergedFS.Open("migrations/9999_nonexistent.up.sql")
	require.Error(t, err)
}

func TestMergedFS_ReadFile(t *testing.T) {
	t.Parallel()

	mergedFSRaw := GetMergedMigrationsFS()
	mergedFS, ok := mergedFSRaw.(fs.ReadFileFS)
	require.True(t, ok, "merged FS should implement fs.ReadFileFS")

	// Test reading pki-ca migration.
	data, err := mergedFS.ReadFile("migrations/2001_ca_items.up.sql")
	require.NoError(t, err)
	require.Contains(t, string(data), "ca_items")

	// Test reading template migration (fallback).
	data, err = mergedFS.ReadFile("migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Test reading non-existent file.
	_, err = mergedFS.ReadFile("migrations/9999_nonexistent.up.sql")
	require.Error(t, err)
}

func TestMergedFS_Stat(t *testing.T) {
	t.Parallel()

	mergedFSRaw := GetMergedMigrationsFS()
	mergedFS, ok := mergedFSRaw.(fs.StatFS)
	require.True(t, ok, "merged FS should implement fs.StatFS")

	// Test stat pki-ca migration.
	info, err := mergedFS.Stat("migrations/2001_ca_items.up.sql")
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))

	// Test stat template migration (fallback).
	info, err = mergedFS.Stat("migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))

	// Test stat non-existent file.
	_, err = mergedFS.Stat("migrations/9999_nonexistent.up.sql")
	require.Error(t, err)
}

func TestMergedFS_ReadDir_Empty(t *testing.T) {
	t.Parallel()

	mergedFSRaw := GetMergedMigrationsFS()
	mergedFS, ok := mergedFSRaw.(fs.ReadDirFS)
	require.True(t, ok, "merged FS should implement fs.ReadDirFS")

	// Test reading non-existent directory.
	_, err := mergedFS.ReadDir("nonexistent")
	require.Error(t, err)
}

func TestApplyPKICAMigrations(t *testing.T) {
	t.Parallel()

	// Open in-memory SQLite database.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	defer func() { require.NoError(t, sqlDB.Close()) }()

	// Configure SQLite for concurrent operations.
	ctx := context.Background()

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err, "WAL mode should be enabled")

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err, "busy timeout should be set")

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	// Apply migrations — should succeed.
	err = ApplyPKICAMigrations(sqlDB, DatabaseTypeSQLite)
	require.NoError(t, err, "migrations should apply successfully")

	// Verify ca_items table was created.
	var tableName string

	err = sqlDB.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='ca_items'").Scan(&tableName)
	require.NoError(t, err, "ca_items table should exist")
	require.Equal(t, "ca_items", tableName)
}

func TestApplyPKICAMigrations_Error(t *testing.T) {
	t.Parallel()

	// Open in-memory SQLite database and immediately close it.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, "file::memory:")
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	// Apply migrations on closed DB — should fail.
	err = ApplyPKICAMigrations(sqlDB, DatabaseTypeSQLite)
	require.Error(t, err, "migrations should fail on closed database")
	require.Contains(t, err.Error(), "failed to apply pki-ca migrations")
}
