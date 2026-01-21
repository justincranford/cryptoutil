// Copyright (c) 2025 Justin Cranford

// Package testutil provides test utilities for the template service server.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"testing"
	"time"

	serverTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilContainer "cryptoutil/internal/shared/container"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for test-containers.
)

// NewInitializedPostgresTestDatabase creates and initializes a PostgreSQL test database using test-containers.
// Returns: sqlDB (*sql.DB), closeDB (cleanup function), error.
// Usage: sqlDB, closeDB, err := NewInitializedPostgresTestDatabase(ctx, migrationsFS); defer closeDB().
func NewInitializedPostgresTestDatabase(ctx context.Context, migrationsFS fs.FS) (*sql.DB, func(), error) {
	postgresContainer, err := cryptoutilContainer.NewPostgresTestContainer(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create PostgreSQL container: %w", err)
	}

	closeDB := func() {
		err := postgresContainer.Terminate(ctx)
		if err != nil {
			fmt.Printf("failed to terminate PostgreSQL container: %v\n", err)
		}
	}

	databaseURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		closeDB()

		return nil, nil, fmt.Errorf("failed to get PostgreSQL connection string: %w", err)
	}

	// Initialize database with migrations.
	gormDB, err := serverTemplateRepository.InitPostgreSQL(ctx, databaseURL, migrationsFS)
	if err != nil {
		closeDB()

		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	} else if gormDB == nil {
		closeDB()

		return nil, nil, fmt.Errorf("gormDB must be non-nil")
	}

	// Verify database connection.
	sqlDB, err := gormDB.DB()
	if err != nil {
		closeDB()

		return nil, nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		closeDB()

		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return sqlDB, closeDB, nil
}

// NewInitializedSQLiteTestDatabase creates and initializes a SQLite test database (in-memory or file-based).
// Returns: sqlDB (*sql.DB), error.
// Usage: sqlDB, err := NewInitializedSQLiteTestDatabase(ctx, migrationsFS, true).
func NewInitializedSQLiteTestDatabase(ctx context.Context, migrationsFS fs.FS, inMemory bool) (*sql.DB, func(), error) {
	var databaseURL string
	if inMemory {
		// Unique in-memory database (prevents cross-test pollution).
		databaseURL = fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.NewString())
	} else {
		// Unique file-based database (no directory path - creates in current dir).
		databaseURL = fmt.Sprintf("file:test_%s.db?cache=shared", googleUuid.NewString())
	}

	// Initialize database with migrations.
	gormDB, err := serverTemplateRepository.InitSQLite(ctx, databaseURL, migrationsFS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	} else if gormDB == nil {
		return nil, nil, fmt.Errorf("gormDB must be non-nil")
	}

	// Verify database connection.
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	closeDB := func() {
		err := sqlDB.Close()
		if err != nil {
			fmt.Printf("Failed to close SQLite in-memory database: %v\n", err)
		}
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		closeDB()

		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return sqlDB, closeDB, nil
}

// InitDatabase is a database-agnostic initializer that dispatches to PostgreSQL or SQLite based on URL scheme.
// Supports: postgres:// (PostgreSQL), file: (SQLite).
// Returns: *gorm.DB, error.
func InitDatabase(ctx context.Context, databaseURL string, migrationsFS fs.FS) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	switch {
	case len(databaseURL) >= 11 && databaseURL[:11] == "postgres://":
		db, err = serverTemplateRepository.InitPostgreSQL(ctx, databaseURL, migrationsFS)
	case len(databaseURL) >= 5 && databaseURL[:5] == "file:":
		db, err = serverTemplateRepository.InitSQLite(ctx, databaseURL, migrationsFS)
	default:
		return nil, fmt.Errorf("unsupported database URL scheme: %s", databaseURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// HelpTestInitDatabaseHappyPaths tests successful database initialization for PostgreSQL and SQLite.
// It verifies that the database schema is correctly created by counting the number of tables.
// Expected table count and query must be provided for each database type.
func HelpTestInitDatabaseHappyPaths(t *testing.T, migrationsFS fs.FS, expectedTableCount int, countTablesQueryPostgres, countTablesQuerySQLite string) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, ctx context.Context) (*sql.DB, func(), string)
	}{
		{
			name: "PostgreSQL Container",
			setupFunc: func(t *testing.T, ctx context.Context) (*sql.DB, func(), string) {
				t.Parallel()

				// Start PostgreSQL container with randomized credentials.
				sqlDB, closeDB, err := NewInitializedPostgresTestDatabase(ctx, migrationsFS)
				require.NoError(t, err)

				return sqlDB, closeDB, countTablesQueryPostgres
			},
		},
		{
			name: "SQLite In-Memory",
			setupFunc: func(t *testing.T, ctx context.Context) (*sql.DB, func(), string) {
				t.Parallel()

				// Start SQLite in-memory database.
				sqlDB, closeDB, err := NewInitializedSQLiteTestDatabase(ctx, migrationsFS, true)
				require.NoError(t, err)

				return sqlDB, closeDB, countTablesQuerySQLite
			},
		},
		{
			name: "SQLite File-Based",
			setupFunc: func(t *testing.T, ctx context.Context) (*sql.DB, func(), string) {
				t.Parallel()

				// Start SQLite file database.
				sqlDB, closeDB, err := NewInitializedSQLiteTestDatabase(ctx, migrationsFS, false)
				require.NoError(t, err)

				return sqlDB, closeDB, countTablesQuerySQLite
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			sqlDB, cleanup, query := tt.setupFunc(t, ctx)
			defer cleanup()

			// Verify schema migration (tables exist).
			var tableCount int

			err := sqlDB.QueryRowContext(ctx, query).Scan(&tableCount)
			require.NoError(t, err)

			require.Equal(t, expectedTableCount, tableCount, "Expected %d tables to be created", expectedTableCount)
		})
	}
}

// HelpTestInitDatabaseSadPaths tests error handling for database initialization failures.
func HelpTestInitDatabaseSadPaths(t *testing.T, migrationsFS fs.FS) {
	tests := []struct {
		name           string
		setupFunc      func(ctx context.Context) error
		expectedErrMsg string
	}{
		{
			name: "Invalid Database Type",
			setupFunc: func(ctx context.Context) error {
				// Initialize database (should fail with unsupported scheme error).
				gormDB, err := InitDatabase(ctx, "mysql://user:pass@localhost:3306/dbname", migrationsFS)
				require.Error(t, err)
				require.Nil(t, gormDB)

				return err
			},
			expectedErrMsg: "unsupported database URL scheme",
		},
		{
			name: "PostgreSQL Connection Error",
			setupFunc: func(ctx context.Context) error {
				// Use 1-second timeout for fast failure (was 5.4s with no timeout).
				ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()

				// Use invalid connection string (nonexistent server).
				gormDB, err := serverTemplateRepository.InitPostgreSQL(ctx, "postgres://user:pass@nonexistent:5432/dbname", migrationsFS)
				if err != nil {
					err = fmt.Errorf("expected error from InitPostgreSQL: %w", err)
				}

				require.Error(t, err)
				require.Nil(t, gormDB)

				return err
			},
			expectedErrMsg: "ping",
		},
		{
			name: "SQLite Invalid Path",
			setupFunc: func(ctx context.Context) error {
				// Use invalid file path (directory doesn't exist).
				gormDB, err := serverTemplateRepository.InitSQLite(ctx, "file:/nonexistent/invalid/path.db", migrationsFS)
				if err != nil {
					err = fmt.Errorf("expected error from InitSQLite: %w", err)
				}

				require.Error(t, err)
				require.Nil(t, gormDB)

				return err
			},
			expectedErrMsg: "", // Error message varies by platform, just check it's an error.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := tt.setupFunc(ctx)

			require.Error(t, err)

			if tt.expectedErrMsg != "" {
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			}
		})
	}
}
