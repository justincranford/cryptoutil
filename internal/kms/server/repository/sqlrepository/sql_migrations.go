// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

var (
	//go:embed postgres/*.sql
	postgresMigrationsFS embed.FS // internal/kms/server/repository/sqlrepository/postgres/*.sql

	//go:embed sqlite/*.sql
	sqliteMigrationsFS embed.FS // internal/kms/server/repository/sqlrepository/sqlite/*.sql
)

// ApplyEmbeddedSQLMigrations applies embedded SQL migrations using default migration files.
func ApplyEmbeddedSQLMigrations(telemetryService *cryptoutilTelemetry.TelemetryService, db *sql.DB, dbType SupportedDBType) error {
	return ApplyEmbeddedSQLMigrationsForService(telemetryService, db, dbType, postgresMigrationsFS, sqliteMigrationsFS)
}

// ApplyEmbeddedSQLMigrationsForService applies embedded SQL migrations using custom migration files.
func ApplyEmbeddedSQLMigrationsForService(telemetryService *cryptoutilTelemetry.TelemetryService, db *sql.DB, dbType SupportedDBType, postgresMigrationsFS embed.FS, sqliteMigrationsFS embed.FS) error {
	telemetryService.Slogger.Debug("applying SQL migrations from embedded files", "driver", dbType)

	var sourceDriver source.Driver

	var databaseDriver database.Driver

	var err error

	switch dbType {
	case DBTypeSQLite:
		sourceDriver, err = iofs.New(sqliteMigrationsFS, "sqlite")
		if err != nil {
			return fmt.Errorf("failed to create iofs source driver for SQLite migration: %w", err)
		}

		databaseDriver, err = sqlite.WithInstance(db, &sqlite.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
	case DBTypePostgres:
		sourceDriver, err = iofs.New(postgresMigrationsFS, "postgres")
		if err != nil {
			return fmt.Errorf("failed to create migration source: %w", err)
		}

		databaseDriver, err = pgx.WithInstance(db, &pgx.Config{})
		if err != nil {
			return fmt.Errorf("failed to create iofs source driver for PostgreSQL migration: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", dbType)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, string(dbType), databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	telemetryService.Slogger.Debug("successfully applied migrations")

	return nil
}
