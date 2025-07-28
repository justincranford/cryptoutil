package sqlrepository

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsFS embed.FS

func ApplyEmbeddedSqlMigrations(telemetryService *cryptoutilTelemetry.TelemetryService, db *sql.DB, driverName string) error {
	telemetryService.Slogger.Debug("applying SQL migrations from embedded files", "driver", driverName)

	var driver database.Driver
	var err error
	var dbType string

	// Determine which driver to use based on the driver name
	switch {
	case strings.Contains(driverName, "postgres"):
		driver, err = postgres.WithInstance(db, &postgres.Config{})
		dbType = "postgres"
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
	case strings.Contains(driverName, "sqlite"):
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		dbType = "sqlite"
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", driverName)
	}

	source, err := iofs.New(migrationsFS, "migrations/"+dbType)
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, dbType, driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	telemetryService.Slogger.Debug("successfully applied SQL migrations", "driver", driverName)
	return nil
}
