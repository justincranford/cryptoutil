// Copyright (c) 2025 Justin Cranford
//
//

// Package repository provides database repository implementations for identity services.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver for database/sql
	_ "modernc.org/sqlite"             // Register CGO-free SQLite driver

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

const (
	// Database type constants.
	dbTypePostgres = "postgres"
	dbTypeSQLite   = "sqlite"

	// SQLite connection pool settings for GORM transaction pattern.
	sqliteMaxOpenConns = 5 // Balance between concurrency and resource usage.
	sqliteMaxIdleConns = 5
)

// initializeDatabase initializes a GORM database connection based on configuration.
func initializeDatabase(ctx context.Context, cfg *cryptoutilIdentityConfig.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case dbTypePostgres:
		dialector = postgres.Open(cfg.DSN)
	case dbTypeSQLite:
		// Convert :memory: to shared cache mode for connection sharing.
		// INVESTIGATION: Shared cache disabled - MaxOpenConns=1 causes deadlock, MaxOpenConns=5 causes rollback visibility.
		// dsn := cfg.DSN
		// if dsn == dsnMemory {
		// 	dsn = dsnMemoryShared
		// }
		dsn := cfg.DSN // Open SQLite database with modernc driver (CGO-free) explicitly.

		sqlDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrDatabaseConnection,
				fmt.Errorf("failed to open SQLite database: %w", err),
			)
		}

		// Enable WAL mode for better concurrency (allows multiple readers + 1 writer).
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
			return nil, cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrDatabaseConnection,
				fmt.Errorf("failed to enable WAL mode: %w", err),
			)
		}

		// Set busy timeout for handling concurrent write operations (30 seconds).
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
			return nil, cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrDatabaseConnection,
				fmt.Errorf("failed to set busy timeout: %w", err),
			)
		}

		// Use GORM sqlite dialector with existing sql.DB connection from modernc driver.
		dialector = sqlite.Dialector{Conn: sqlDB}
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("unsupported database type: %s", cfg.Type),
		)
	}

	// Configure GORM logger (default to silent for production).
	gormLogger := logger.Default.LogMode(logger.Silent)

	// Open database connection.
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Disable automatic transactions for Create/Update/Delete (we manage transactions explicitly)
	})
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to connect to database: %w", err),
		)
	}

	// Configure connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to get database instance: %w", err),
		)
	}

	// Apply connection pool settings.
	// For SQLite: Limit pool size but allow enough connections for GORM transaction + operation pattern.
	// SQLite in WAL mode supports multiple readers + 1 writer.
	// GORM with SkipDefaultTransaction=true still needs 2+ connections: one for explicit transaction,
	// one for the CRUD operation inside the transaction.
	// busy_timeout makes SQLite retry when a connection is locked by another transaction.
	if cfg.Type == dbTypeSQLite {
		sqlDB.SetMaxOpenConns(sqliteMaxOpenConns) // Balance between concurrency and resource usage for GORM pattern.
		sqlDB.SetMaxIdleConns(sqliteMaxIdleConns)
		sqlDB.SetConnMaxLifetime(0) // Never close connections for in-memory DB.
		sqlDB.SetConnMaxIdleTime(0)
	} else {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// Verify connection.
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to ping database: %w", err),
		)
	}

	return db, nil
}
