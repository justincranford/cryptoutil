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
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// DatabaseType represents supported database types for learn-im.
type DatabaseType string

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite DatabaseType = "sqlite3"
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL DatabaseType = "postgres"
)

var (
	//go:embed migrations/*.sql
	migrationsFS embed.FS
)

// ApplyMigrations runs database migrations for learn-im service.
//
// Migrations are embedded in the binary and applied automatically on startup.
// Compatible with both PostgreSQL and SQLite (using TEXT type for cross-DB compatibility).
func ApplyMigrations(db *sql.DB, dbType DatabaseType) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver: %w", err)
	}

	var databaseDriver database.Driver

	switch dbType {
	case DatabaseTypeSQLite:
		databaseDriver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
	case DatabaseTypePostgreSQL:
		databaseDriver, err = pgx.WithInstance(db, &pgx.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, string(dbType), databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
