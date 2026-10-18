// Copyright (c) 2025-2026 Justin Cranford.
//
//

package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sync"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var migrationMutex sync.Mutex

// ResetMigrationStateForTesting is retained for compatibility with existing tests.
// Migrate no longer stores process-global migration state, so there is nothing to reset.
func ResetMigrationStateForTesting() {}

// Migrate applies SQL migrations from embedded files.
// Thread-safe for concurrent test execution via mutex protection.
func Migrate(ctx context.Context, db *sql.DB, dbType string) error {
	migrationMutex.Lock()
	defer migrationMutex.Unlock()

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("migration context: %w", err)
	}

	// Enable SQLite pragmas for proper foreign key handling (SQLite only).
	// Postgres enables foreign keys by default and does not support PRAGMA syntax.
	if dbType == dbTypeSQLite {
		if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
			return fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}

	// Create schema_migrations table manually to avoid "no such table" error.
	// SQLite and Postgres use the same schema for the migrations table.
	const createSchemaTable = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint not null primary key,
			dirty boolean not null
		);
	`

	if _, err := db.ExecContext(ctx, createSchemaTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Create iofs source driver from embedded filesystem.
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver for migrations: %w", err)
	}

	// Create database driver and migrate instance based on database type.
	var m *migrate.Migrate

	if dbType == dbTypeSQLite {
		sqliteDriver, err := sqlite.WithInstance(db, &sqlite.Config{
			MigrationsTable: valueSchemaMigrations,
		})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}

		m, err = migrate.NewWithInstance("iofs", sourceDriver, dbTypeSQLite, sqliteDriver)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
	} else {
		postgresDriver, err := postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: valueSchemaMigrations,
		})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}

		m, err = migrate.NewWithInstance("iofs", sourceDriver, cryptoutilSharedMagic.DockerServicePostgres, postgresDriver)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
	}

	// Apply migrations.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
