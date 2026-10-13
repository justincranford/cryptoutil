// Copyright (c) 2025-2026 Justin Cranford.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDatabaseInitialization_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configure  func(*databaseInitializationOperations)
		initialize func(databaseInitializationOperations) (*gorm.DB, error)
		wantErr    string
	}{
		{
			name: "SQLite SQL open error",
			configure: func(operations *databaseInitializationOperations) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					return nil, errors.New("injected SQL open error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: "failed to open SQLite database",
		},
		{
			name: "SQLite GORM open error",
			configure: func(operations *databaseInitializationOperations) {
				operations.openGORM = func(gorm.Dialector, ...gorm.Option) (*gorm.DB, error) {
					return nil, errors.New("injected GORM open error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: "failed to initialize GORM for SQLite",
		},
		{
			name: "SQLite migration error",
			configure: func(operations *databaseInitializationOperations) {
				operations.applyMigrations = func(*sql.DB, DatabaseType, fs.FS) error {
					return errors.New("injected migration error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initSQLiteWithOperations(context.Background(), cryptoutilSharedMagic.SQLiteInMemoryDSN, testDBMigrationsFS, operations)
			},
			wantErr: "failed to apply migrations",
		},
		{
			name: "PostgreSQL SQL open error",
			configure: func(operations *databaseInitializationOperations) {
				operations.openSQL = func(string, string) (*sql.DB, error) {
					return nil, errors.New("injected SQL open error")
				}
			},
			initialize: func(operations databaseInitializationOperations) (*gorm.DB, error) {
				return initPostgreSQLWithOperations(context.Background(), "postgres://invalid:5432/test", testDBMigrationsFS, operations)
			},
			wantErr: "failed to open PostgreSQL database",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := defaultDatabaseInitializationOperations()
			tc.configure(&operations)

			db, err := tc.initialize(operations)
			require.ErrorContains(t, err, tc.wantErr)
			require.Nil(t, db)
		})
	}
}

// TestMigrationRunner_MigrateNewWithInstanceError covers the migrate.NewWithInstance error path.
func TestMigrationRunner_MigrateNewWithInstanceError(t *testing.T) {
	t.Parallel()

	// Open a real SQLite database so that iofs and sqlite driver creation succeed,
	// then fail at the migrate.NewWithInstance step.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteInMemoryDSN)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, sqlDB.Close())
	}()

	stubMigrateNewInstanceFn := func(_ string, _ source.Driver, _ string, _ database.Driver) (*migrate.Migrate, error) {
		return nil, errors.New("injected migrate instance error")
	}

	runner := newMigrationRunnerInternal(testDBMigrationsFS, "test_migrations", stubMigrateNewInstanceFn)
	applyErr := runner.Apply(sqlDB, DatabaseTypeSQLite)
	require.Error(t, applyErr)
	require.Contains(t, applyErr.Error(), "failed to create migrate instance")
}
