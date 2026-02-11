// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// DatabaseConnection provides unified access to GORM database connections.
// All services MUST use GORM - there are no optional modes.
type DatabaseConnection struct {
	gormDB *gorm.DB
}

// NewDatabaseConnection creates a GORM database connection.
func NewDatabaseConnection(db *gorm.DB) (*DatabaseConnection, error) {
	if db == nil {
		return nil, fmt.Errorf("gorm.DB is nil")
	}

	return &DatabaseConnection{
		gormDB: db,
	}, nil
}

// GORM returns the GORM database connection.
func (d *DatabaseConnection) GORM() *gorm.DB {
	return d.gormDB
}

// SQL returns the underlying sql.DB from GORM.
// Useful for operations requiring raw database/sql access.
func (d *DatabaseConnection) SQL() (*sql.DB, error) {
	if d.gormDB == nil {
		return nil, fmt.Errorf("gorm.DB is nil")
	}

	sqlDB, err := d.gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	return sqlDB, nil
}

// Close closes the database connection.
func (d *DatabaseConnection) Close() error {
	if d.gormDB == nil {
		return nil
	}

	sqlDB, err := d.gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// DatabaseConfig provides configuration for database initialization.
type DatabaseConfig struct {
	// URL is the database connection URL.
	URL string

	// VerboseMode enables verbose logging.
	VerboseMode bool

	// ContainerMode specifies testcontainer behavior ("disabled", "preferred", "required").
	ContainerMode string
}

// NewDatabaseConfig creates a database configuration.
func NewDatabaseConfig(url string) *DatabaseConfig {
	return &DatabaseConfig{
		URL:           url,
		VerboseMode:   false,
		ContainerMode: "disabled",
	}
}
