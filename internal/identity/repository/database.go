package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite" // Register CGO-free SQLite driver

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

const (
	// dsnMemory is the DSN for in-memory SQLite databases.
	dsnMemory = ":memory:"
	// dsnMemoryShared is the DSN for shared in-memory SQLite databases.
	dsnMemoryShared = "file::memory:?cache=shared"
)

// initializeDatabase initializes a GORM database connection based on configuration.
func initializeDatabase(ctx context.Context, cfg *cryptoutilIdentityConfig.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "sqlite":
		// For in-memory databases, use shared cache mode to ensure all connections share the same database.
		dsn := cfg.DSN
		if dsn == dsnMemory {
			dsn = dsnMemoryShared
		}

		// Open SQLite database with modernc driver (CGO-free).
		sqlDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrDatabaseConnection,
				fmt.Errorf("failed to open SQLite database: %w", err),
			)
		}

		// For in-memory databases with shared cache, keep at least one connection alive.
		if cfg.DSN == dsnMemory {
			sqlDB.SetMaxIdleConns(1)
			sqlDB.SetMaxOpenConns(1)
			sqlDB.SetConnMaxLifetime(0) // Never close connections for in-memory DB.
		}

		// Use GORM sqlite dialector with existing sql.DB connection.
		dialector = sqlite.Dialector{Conn: sqlDB}
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("unsupported database type: %s", cfg.Type),
		)
	}

	// Configure GORM logger (default to silent for production).
	gormLogger := logger.Default.LogMode(logger.Silent)

	// Open database connection.
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to connect to database: %w", err),
		)
	}

	// Configure connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to get database instance: %w", err),
		)
	}

	// Apply connection pool settings only if not using in-memory database.
	// For in-memory databases, we already set MaxIdleConns=1 and MaxOpenConns=1 above.
	if cfg.DSN != dsnMemory {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// Verify connection.
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to ping database: %w", err),
		)
	}

	return db, nil
}
