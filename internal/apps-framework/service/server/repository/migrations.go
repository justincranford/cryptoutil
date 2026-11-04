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
	"io/fs"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	migrateSource "github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// MigrationsFS contains embedded base infrastructure migrations (1001-1004).
// Services that use service-framework MUST apply these migrations first,
// then apply their own app-specific migrations (1005+).
//
// Migration version numbering convention:
//   - 1001-1004: Service-framework base infrastructure (session mgmt, barrier, realms template, multi-tenancy)
//   - 1005+: App-specific tables (sm-kms users/messages, identity accounts, etc.)
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
	fsys                 interface{ fs.FS }
	migrationsPath       string
	migrateNewInstanceFn func(sourceName string, sourceDriver migrateSource.Driver, databaseName string, databaseDriver database.Driver) (*migrate.Migrate, error)
}

// DatabaseType represents supported database types.
type DatabaseType string

const (
	// DatabaseTypeSQLite represents SQLite database.
	DatabaseTypeSQLite DatabaseType = "sqlite3"
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	DatabaseTypePostgreSQL DatabaseType = cryptoutilSharedMagic.DockerServicePostgres
)

// NewMigrationRunner creates a new migration runner with filesystem.
//
// Parameters:
//   - fsys: Filesystem containing migration files (can be embed.FS or any fs.FS implementation)
//   - migrationsPath: Path within fsys where migrations are located (e.g., "migrations")
//
// Returns configured migration runner ready to apply migrations.
func NewMigrationRunner(fsys interface{ fs.FS }, migrationsPath string) *MigrationRunner {
	return newMigrationRunnerInternal(fsys, migrationsPath, migrate.NewWithInstance)
}

func newMigrationRunnerInternal(
	fsys interface{ fs.FS },
	migrationsPath string,
	migrateNewInstanceFn func(sourceName string, sourceDriver migrateSource.Driver, databaseName string, databaseDriver database.Driver) (*migrate.Migrate, error),
) *MigrationRunner {
	return &MigrationRunner{
		fsys:                 fsys,
		migrationsPath:       migrationsPath,
		migrateNewInstanceFn: migrateNewInstanceFn,
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
	return r.ApplyContext(context.Background(), db, dbType)
}

// ApplyContext reuses production migration drivers and signals graceful stop
// between embedded migrations on cancellation. In-flight SQL still obeys the
// driver's/database's statement and lock deadlines; it is not falsely reported
// as an immediately cancellable query.
func (r *MigrationRunner) ApplyContext(ctx context.Context, db *sql.DB, dbType DatabaseType) (result error) {
	if ctx == nil || db == nil {
		return errors.New("migration requires context and database")
	}

	view, closeView := migrationDatabaseView(ctx, db, nil, dbType == DatabaseTypePostgreSQL)
	defer func() { result = errors.Join(result, closeView()) }()

	return r.applyView(ctx, view, dbType)
}

// ApplyTransaction uses the same production migration drivers and embedded
// source, while nested driver transactions remain savepoints of the owner's
// transaction. The caller alone commits or rolls back that transaction.
func (r *MigrationRunner) ApplyTransaction(ctx context.Context, tx *sql.Tx, dbType DatabaseType) (result error) {
	if ctx == nil || tx == nil {
		return errors.New("migration requires context and transaction")
	}

	view, closeView := migrationDatabaseView(ctx, nil, tx, dbType == DatabaseTypePostgreSQL)
	defer func() { result = errors.Join(result, closeView()) }()

	return r.applyView(ctx, view, dbType)
}

func (r *MigrationRunner) applyView(ctx context.Context, db *sql.DB, dbType DatabaseType) (result error) {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("migration context: %w", err)
	}

	sourceDriver, err := iofs.New(r.fsys, r.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver: %w", err)
	}

	defer func() { result = errors.Join(result, sourceDriver.Close()) }()

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

	defer func() { result = errors.Join(result, databaseDriver.Close()) }()

	m, err := r.migrateNewInstanceFn("iofs", sourceDriver, string(dbType), databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("migration context: %w", err)
	}

	stop := context.AfterFunc(ctx, func() {
		select {
		case m.GracefulStop <- true:
		default:
		}
	})
	defer stop()

	applyErr := m.Up()

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("migration context: %w", err)
	}

	if err := applyErr; err != nil && !errors.Is(err, migrate.ErrNoChange) {
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
	case cryptoutilSharedMagic.DockerServicePostgres, "postgresql":
		dbType = DatabaseTypePostgreSQL
	default:
		return fmt.Errorf("unsupported database type: %s", databaseType)
	}

	runner := NewMigrationRunner(migrationFS, migrationsPath)

	return runner.Apply(db, dbType)
}
