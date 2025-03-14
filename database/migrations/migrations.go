package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sqlite/*.sql
var SQLiteMigrationsEmbedFS embed.FS

func ApplyMigrations(db *sql.DB) error {
	if db == nil {
		return errors.New("db can't be nil")
	}
	sourceDriver, err := iofs.New(SQLiteMigrationsEmbedFS, "sqlite")
	if err != nil {
		return fmt.Errorf("create iofs source driver failed: %w", err)
	}
	databaseDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("create driver failed: %w", err)
	}

	migrations, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", databaseDriver)
	if err != nil {
		return fmt.Errorf("create sqlite failed: %w", err)
	}

	if err = migrations.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run sqlite failed: %w", err)
	}
	return nil
}
