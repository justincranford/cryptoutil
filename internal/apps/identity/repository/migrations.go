// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration state tracking for parallel test safety.
var (
	migrationMutex sync.Mutex
	migratedDBs    = make(map[string]bool)
)

// ResetMigrationStateForTesting clears the migration state cache.
// ONLY for use in test setup to ensure clean migration state per test.
// NOT safe for production use - only call from test initialization code.
func ResetMigrationStateForTesting() {
	migrationMutex.Lock()
	defer migrationMutex.Unlock()

	migratedDBs = make(map[string]bool)
}

// Migrate applies SQL migrations from embedded files.
// Thread-safe for concurrent test execution via mutex protection.
func Migrate(db *sql.DB, dbType string) error {
	migrationMutex.Lock()
	defer migrationMutex.Unlock()

	ctx := context.TODO() // Migration runs during startup, no request context available

	// For in-memory databases, skip the cache entirely and always run migrations.
	// This handles parallel tests where each test should have its own database.
	// For persistent databases (file-based or postgres), use pointer caching.
	isMemoryDB := false

	if dbType == "sqlite" {
		// Check if this is an in-memory database by querying the database path.
		var dbPath string

		row := db.QueryRowContext(ctx, "PRAGMA database_list")

		var (
			seq  int
			name string
		)

		if err := row.Scan(&seq, &name, &dbPath); err == nil && (dbPath == "" || dbPath == cryptoutilSharedMagic.SQLiteMemoryPlaceholder) {
			isMemoryDB = true
		}
	}

	var dbKey string
	if !isMemoryDB {
		// For persistent databases, use pointer-based caching.
		dbKey = fmt.Sprintf("%p", db)
		if migratedDBs[dbKey] {
			return nil // Already migrated this DB instance
		}
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
			MigrationsTable: "schema_migrations",
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
			MigrationsTable: "schema_migrations",
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

	// Only cache persistent databases (not in-memory).
	if dbKey != "" {
		migratedDBs[dbKey] = true
	}

	return nil
}
