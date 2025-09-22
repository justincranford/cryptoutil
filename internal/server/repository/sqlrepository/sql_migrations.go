package sqlrepository

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

var (
	//go:embed postgres/*.sql
	postgresMigrationsFS embed.FS

	//go:embed sqlite/*.sql
	sqliteMigrationsFS embed.FS
)

func ApplyEmbeddedSQLMigrations(telemetryService *cryptoutilTelemetry.TelemetryService, db *sql.DB, dbType SupportedDBType) error {
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
		databaseDriver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
	case DBTypePostgres:
		sourceDriver, err = iofs.New(postgresMigrationsFS, "postgres")
		if err != nil {
			return fmt.Errorf("failed to create migration source: %w", err)
		}
		databaseDriver, err = postgres.WithInstance(db, &postgres.Config{})
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
