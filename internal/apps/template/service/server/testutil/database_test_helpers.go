// Copyright (c) 2025 Justin Cranford

package testutil

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	serverTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilContainer "cryptoutil/internal/shared/container"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for test-containers.
)

// NewInitializedPostgresTestDatabase creates and initializes a PostgreSQL test database using test-containers.
// Returns: sqlDB (*sql.DB), closeDB (cleanup function), error.
// Usage: sqlDB, closeDB, err := NewInitializedPostgresTestDatabase(ctx, migrationsFS); defer closeDB().
func NewInitializedPostgresTestDatabase(ctx context.Context, migrationsFS embed.FS) (*sql.DB, func(), error) {
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
func NewInitializedSQLiteTestDatabase(ctx context.Context, migrationsFS embed.FS, inMemory bool) (*sql.DB, error) {
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
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	} else if gormDB == nil {
		return nil, fmt.Errorf("gormDB must be non-nil")
	}

	// Verify database connection.
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return sqlDB, nil
}

// InitDatabase is a database-agnostic initializer that dispatches to PostgreSQL or SQLite based on URL scheme.
// Supports: postgres:// (PostgreSQL), file: (SQLite).
// Returns: *gorm.DB, error.
func InitDatabase(ctx context.Context, databaseURL string, migrationsFS embed.FS) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

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
