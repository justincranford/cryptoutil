package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // Register CGO-free SQLite driver

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

// initializeDatabase initializes a GORM database connection based on configuration.
func initializeDatabase(ctx context.Context, cfg *cryptoutilIdentityConfig.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "sqlite":
		// Use modernc.org/sqlite (CGO-free) instead of mattn/go-sqlite3
		dialector = sqlite.New(sqlite.Config{
			DriverName: "sqlite",
			DSN:        cfg.DSN,
		})
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

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	// Verify connection.
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to ping database: %w", err),
		)
	}

	return db, nil
}
