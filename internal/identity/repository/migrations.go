// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies SQL migrations from embedded files.
func Migrate(db *sql.DB) error {
	// Enable SQLite pragmas for proper foreign key handling.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create schema_migrations table manually to avoid "no such table" error.
	// This is required because golang-migrate's sqlite3 driver tries to DELETE/INSERT
	// before the table exists on first run.
	createSchemaTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint not null primary key,
			dirty boolean not null
		);
	`
	if _, err := db.Exec(createSchemaTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Create iofs source driver from embedded filesystem.
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver for migrations: %w", err)
	}

	// Create SQLite database driver.
	databaseDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create sqlite driver: %w", err)
	}

	// Create migrate instance.
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Apply migrations.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
