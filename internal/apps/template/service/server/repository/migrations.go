// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrateNewWithInstanceFn is injectable for testing the migrate.NewWithInstance error path.
var migrateNewWithInstanceFn = migrate.NewWithInstance

// DatabaseType represents supported database types.
type DatabaseType string

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite DatabaseType = "sqlite3"
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL DatabaseType = "postgres"
)

// MigrationsFS contains embedded base infrastructure migrations (1001-1004).
// Services that use service-template MUST apply these migrations first,
// then apply their own app-specific migrations (1005+).
//
// Migration version numbering convention:
//   - 1001-1004: Service-template base infrastructure (session mgmt, barrier, realms template, multi-tenancy)
//   - 1005+: App-specific tables (sm-im users/messages, identity accounts, etc.)
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// MigrationRunner applies database migrations from embedded filesystem.
//
// Supports both PostgreSQL and SQLite databases using golang-migrate.
// Migrations are embedded in the service binary and applied automatically.
//
// Usage:
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	runner := NewMigrationRunner(migrationsFS, "migrations")
//	err := runner.Apply(sqlDB, DatabaseTypePostgreSQL)
type MigrationRunner struct {
	fsys           interface{ fs.FS }
	migrationsPath string
}

// NewMigrationRunner creates a new migration runner with filesystem.
//
// Parameters:
//   - fsys: Filesystem containing migration files (can be embed.FS or any fs.FS implementation)
//   - migrationsPath: Path within fsys where migrations are located (e.g., "migrations")
//
// Returns configured migration runner ready to apply migrations.
func NewMigrationRunner(fsys interface{ fs.FS }, migrationsPath string) *MigrationRunner {
	return &MigrationRunner{
		fsys:           fsys,
		migrationsPath: migrationsPath,
	}
}

// Apply runs database migrations for the specified database type.
//
// Migrations are applied in order based on version numbers in filenames.
// If migrations are already up-to-date, returns nil (not an error).
//
// Parameters:
//   - db: SQL database connection
//   - dbType: Database type (DatabaseTypeSQLite or DatabaseTypePostgreSQL)
//
// Returns error if migrations fail to apply.
func (r *MigrationRunner) Apply(db *sql.DB, dbType DatabaseType) error {
	sourceDriver, err := iofs.New(r.fsys, r.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver: %w", err)
	}

	var databaseDriver database.Driver

	switch dbType {
	case DatabaseTypeSQLite:
		databaseDriver, err = sqlite.WithInstance(db, &sqlite.Config{})
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

	m, err := migrateNewWithInstanceFn("iofs", sourceDriver, string(dbType), databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// ApplyMigrationsFromFS is a convenience function for applying migrations from embed.FS.
// Automatically detects database type from provided string.
//
// Parameters:
//   - db: SQL database connection
//   - migrationFS: Embedded filesystem containing migration files
//   - migrationsPath: Path within filesystem (e.g., "migrations")
//   - databaseType: Either "sqlite" or "postgres"
//
// Returns error if migrations fail to apply.
func ApplyMigrationsFromFS(db *sql.DB, migrationFS fs.FS, migrationsPath string, databaseType string) error {
	var dbType DatabaseType

	switch databaseType {
	case "sqlite", "sqlite3":
		dbType = DatabaseTypeSQLite
	case "postgres", "postgresql":
		dbType = DatabaseTypePostgreSQL
	default:
		return fmt.Errorf("unsupported database type: %s", databaseType)
	}

	runner := NewMigrationRunner(migrationFS, migrationsPath)

	return runner.Apply(db, dbType)
}
