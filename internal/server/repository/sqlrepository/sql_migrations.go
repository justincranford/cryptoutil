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
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

var (
	//go:embed postgres/*.sql
	postgresMigrationsFS embed.FS

	//go:embed sqlite/*.sql
	sqliteMigrationsFS embed.FS
)

func ApplyEmbeddedSqlMigrations(telemetryService *cryptoutilTelemetry.TelemetryService, db *sql.DB, driverName string) error {
	telemetryService.Slogger.Debug("applying SQL migrations from embedded files", "driver", driverName)

	var driver database.Driver
	var err error
	var dbType string

	// Determine which driver to use based on the driver name
	var source source.Driver
	switch {
	case strings.Contains(driverName, "postgres"):
		driver, err = postgres.WithInstance(db, &postgres.Config{})
		dbType = "postgres"
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
		source, err = iofs.New(postgresMigrationsFS, "postgres")
		if err != nil {
			return fmt.Errorf("failed to create migration source: %w", err)
		}
	case strings.Contains(driverName, "sqlite"):
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		dbType = "sqlite"
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
		source, err = iofs.New(sqliteMigrationsFS, "sqlite")
		if err != nil {
			return fmt.Errorf("failed to create migration source: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", driverName)
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
