package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitPostgreSQL initializes a PostgreSQL database connection with GORM.
func InitPostgreSQL(ctx context.Context, databaseURL string, migrationsFS fs.FS) (*gorm.DB, error) {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Verify connection.
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	// Create GORM instance.
	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM for PostgreSQL: %w", err)
	}

	// Configure connection pool.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.PostgreSQLMaxOpenConns)       // 25
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.PostgreSQLMaxIdleConns)       // 10
	sqlDB.SetConnMaxLifetime(cryptoutilSharedMagic.PostgreSQLConnMaxLifetime) // 1 hour

	// Run migrations.
	if err := ApplyMigrations(sqlDB, DatabaseTypePostgreSQL, migrationsFS); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// InitSQLite initializes a SQLite database connection with GORM.
func InitSQLite(ctx context.Context, databaseURL string, migrationsFS fs.FS) (*gorm.DB, error) {
	// Open SQLite database.
	sqlDB, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Enable WAL mode for concurrent operations.
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout for concurrent write operations.
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Create GORM instance.
	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM for SQLite: %w", err)
	}

	// Configure connection pool for GORM transactions.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections) // 5
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections) // 5
	sqlDB.SetConnMaxLifetime(0)                                           // In-memory: never close

	// Run migrations.
	if err := ApplyMigrations(sqlDB, DatabaseTypeSQLite, migrationsFS); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// ApplyMigrations applies database migrations using the embedded migration files.
func ApplyMigrations(db *sql.DB, dbType DatabaseType, migrationsFS fs.FS) error {
	runner := NewMigrationRunner(migrationsFS, "migrations")

	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return runner.Apply(db, dbType)
}
