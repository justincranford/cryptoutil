// Copyright (c) 2025 Justin Cranford

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"errors"
	"io/fs"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestInitSQLite_SqlOpenError covers the sqlOpenFn error path in InitSQLite.
// NOT parallel — modifies package-level injectable var.
func TestInitSQLite_SqlOpenError(t *testing.T) {
	original := sqlOpenFn
	sqlOpenFn = func(_, _ string) (*sql.DB, error) {
		return nil, errors.New("injected sql open error")
	}

	defer func() { sqlOpenFn = original }()

	db, err := InitSQLite(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "failed to open SQLite database")
}

// TestInitSQLite_GormOpenError covers the gormOpenFn error path in InitSQLite.
// NOT parallel — modifies package-level injectable var.
func TestInitSQLite_GormOpenError(t *testing.T) {
	original := gormOpenFn
	gormOpenFn = func(_ gorm.Dialector, _ ...gorm.Option) (*gorm.DB, error) {
		return nil, errors.New("injected gorm open error")
	}

	defer func() { gormOpenFn = original }()

	db, err := InitSQLite(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "failed to initialize GORM for SQLite")
}

// TestInitSQLite_ApplyMigrationsInjectedError covers the applyMigrationsFn error path in InitSQLite.
// NOT parallel — modifies package-level injectable var.
func TestInitSQLite_ApplyMigrationsInjectedError(t *testing.T) {
	original := applyMigrationsFn
	applyMigrationsFn = func(_ *sql.DB, _ DatabaseType, _ fs.FS) error {
		return errors.New("injected migration error")
	}

	defer func() { applyMigrationsFn = original }()

	db, err := InitSQLite(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "failed to apply migrations")
}

// TestInitPostgreSQL_SqlOpenError covers the sqlOpenFn error path in InitPostgreSQL.
// NOT parallel — modifies package-level injectable var.
func TestInitPostgreSQL_SqlOpenError(t *testing.T) {
	original := sqlOpenFn
	sqlOpenFn = func(_, _ string) (*sql.DB, error) {
		return nil, errors.New("injected sql open error")
	}

	defer func() { sqlOpenFn = original }()

	db, err := InitPostgreSQL(context.Background(), "postgres://invalid:5432/test", testDBMigrationsFS)
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "failed to open PostgreSQL database")
}

// TestMigrationRunner_MigrateNewWithInstanceError covers the migrateNewWithInstanceFn error path.
// NOT parallel — modifies package-level injectable var.
func TestMigrationRunner_MigrateNewWithInstanceError(t *testing.T) {
	original := migrateNewWithInstanceFn

	// Open a real SQLite database so that iofs and sqlite driver creation succeed,
	// then fail at the migrate.NewWithInstance step.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, sqlDB.Close())
	}()

	migrateNewWithInstanceFn = func(_ string, _ source.Driver, _ string, _ database.Driver) (*migrate.Migrate, error) {
		return nil, errors.New("injected migrate instance error")
	}

	defer func() { migrateNewWithInstanceFn = original }()

	runner := NewMigrationRunner(testDBMigrationsFS, "test_migrations")
	applyErr := runner.Apply(sqlDB, DatabaseTypeSQLite)
	require.Error(t, applyErr)
	require.Contains(t, applyErr.Error(), "failed to create migrate instance")
}
