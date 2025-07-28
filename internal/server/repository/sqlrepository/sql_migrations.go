package sqlrepository

import (
	"database/sql"
	"embed"
	"fmt"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsFS embed.FS

func ApplyEmbeddedSqlMigrations(telemetryService *cryptoutilTelemetry.TelemetryService, db *sql.DB) error {
	telemetryService.Slogger.Debug("applying SQL migrations from embedded files")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	telemetryService.Slogger.Debug("successfully applied SQL migrations")
	return nil
}
