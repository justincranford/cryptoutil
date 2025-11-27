package rotate_secret

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	_ "modernc.org/sqlite"
)

func setupDatabase() (*gorm.DB, error) {
	// For CLI tool, use SQLite in-memory database.
	// Production deployment would use PostgreSQL DSN from config.
	const dsn = "file::memory:?cache=shared"

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Apply SQLite pragmas for concurrent operations.
	ctx := context.Background()
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	const busyTimeout = "PRAGMA busy_timeout = 30000;"
	if _, err := sqlDB.ExecContext(ctx, busyTimeout); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Configure connection pool for SQLite.
	const (
		maxOpenConns = 5
		maxIdleConns = 5
	)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// Create GORM instance.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM instance: %w", err)
	}

	// Run migrations.
	if err := db.AutoMigrate(
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
