// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
// MigrationsFS contains embedded SQL migration files for JOSE repository.
var MigrationsFS embed.FS

// RunMigrations runs all database migrations for JOSE repository.
func RunMigrations(db *gorm.DB, driverName string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	sourceDriver, err := iofs.New(MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	var databaseDriver database.Driver

	switch driverName {
	case "postgres", "postgresql":
		databaseDriver, err = pgx.WithInstance(sqlDB, &pgx.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
	case "sqlite", "sqlite3":
		databaseDriver, err = sqlite.WithInstance(sqlDB, &sqlite.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported driver: %s", driverName)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, driverName, databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
