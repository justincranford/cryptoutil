// Copyright (c) 2025 Justin Cranford

// Package e2e_helpers provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from sm-im implementation to support 9-service migration.
package e2e_helpers

import (
	"context"
	"database/sql"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// InitTestDB creates an in-memory SQLite database with proper configuration.
// Reusable for all services using GORM with SQLite for testing.
//
// Parameters:
//   - applyMigrations: function that applies schema migrations using sql.DB
//
// Returns configured GORM DB instance ready for testing.
func InitTestDB(ctx context.Context, applyMigrations func(*sql.DB) error) (*gorm.DB, error) {
	// Create unique in-memory database per test to avoid table conflicts.
	dbID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Configure SQLite for concurrent operations (WAL mode + busy timeout).
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	if err != nil {
		return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	// Configure connection pool for GORM transactions (MaxOpenConns=5 required).
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: keep connections alive.

	// Wrap with GORM using sqlite Dialector.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM DB: %w", err)
	}

	// Apply service-specific migrations.
	if applyMigrations != nil {
		err = applyMigrations(sqlDB)
		if err != nil {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return db, nil
}
