// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// DatabaseMode represents the database access pattern.
type DatabaseMode string

const (
	// DatabaseModeGORM uses GORM ORM for database access (default for template-based services).
	DatabaseModeGORM DatabaseMode = "gorm"
	// DatabaseModeRawSQL uses raw database/sql for database access (KMS-style services).
	DatabaseModeRawSQL DatabaseMode = "raw_sql"
	// DatabaseModeDual uses both GORM and raw SQL (hybrid services).
	DatabaseModeDual DatabaseMode = "dual"
)

// DatabaseConnection provides unified access to database connections.
// Supports GORM, raw SQL, or both depending on service requirements.
type DatabaseConnection struct {
	mode   DatabaseMode
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

// NewDatabaseConnectionGORM creates a GORM-only database connection.
func NewDatabaseConnectionGORM(db *gorm.DB) (*DatabaseConnection, error) {
	if db == nil {
		return nil, fmt.Errorf("gorm.DB is nil")
	}

	return &DatabaseConnection{
		mode:   DatabaseModeGORM,
		gormDB: db,
	}, nil
}

// NewDatabaseConnectionRawSQL creates a raw SQL-only database connection.
func NewDatabaseConnectionRawSQL(db *sql.DB) (*DatabaseConnection, error) {
	if db == nil {
		return nil, fmt.Errorf("sql.DB is nil")
	}

	return &DatabaseConnection{
		mode:  DatabaseModeRawSQL,
		sqlDB: db,
	}, nil
}

// NewDatabaseConnectionDual creates a dual-mode database connection with both GORM and raw SQL.
func NewDatabaseConnectionDual(gormDB *gorm.DB, sqlDB *sql.DB) (*DatabaseConnection, error) {
	if gormDB == nil {
		return nil, fmt.Errorf("gorm.DB is nil")
	}

	if sqlDB == nil {
		return nil, fmt.Errorf("sql.DB is nil")
	}

	return &DatabaseConnection{
		mode:   DatabaseModeDual,
		gormDB: gormDB,
		sqlDB:  sqlDB,
	}, nil
}

// Mode returns the database access mode.
func (d *DatabaseConnection) Mode() DatabaseMode {
	return d.mode
}

// GORM returns the GORM database connection.
// Returns nil if mode is DatabaseModeRawSQL.
func (d *DatabaseConnection) GORM() *gorm.DB {
	return d.gormDB
}

// SQL returns the raw SQL database connection.
// For GORM mode, extracts *sql.DB from GORM.
// For RawSQL mode, returns the direct connection.
// For Dual mode, returns the raw SQL connection.
func (d *DatabaseConnection) SQL() (*sql.DB, error) {
	switch d.mode {
	case DatabaseModeGORM:
		if d.gormDB == nil {
			return nil, fmt.Errorf("gorm.DB is nil")
		}

		sqlDB, err := d.gormDB.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
		}

		return sqlDB, nil
	case DatabaseModeRawSQL, DatabaseModeDual:
		if d.sqlDB == nil {
			return nil, fmt.Errorf("sql.DB is nil")
		}

		return d.sqlDB, nil
	default:
		return nil, fmt.Errorf("unknown database mode: %s", d.mode)
	}
}

// HasGORM returns true if GORM is available.
func (d *DatabaseConnection) HasGORM() bool {
	return d.gormDB != nil
}

// HasRawSQL returns true if raw SQL is available.
func (d *DatabaseConnection) HasRawSQL() bool {
	return d.sqlDB != nil || d.gormDB != nil // GORM can extract sql.DB
}

// Close closes the database connection(s).
func (d *DatabaseConnection) Close() error {
	var errs []error

	// Close raw SQL if present (and not managed by GORM).
	if d.mode == DatabaseModeRawSQL || d.mode == DatabaseModeDual {
		if d.sqlDB != nil {
			if err := d.sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close sql.DB: %w", err))
			}
		}
	}

	// For GORM mode, close via GORM.
	if d.mode == DatabaseModeGORM && d.gormDB != nil {
		sqlDB, err := d.gormDB.DB()
		if err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close gorm.DB underlying sql.DB: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database: %v", errs)
	}

	return nil
}

// DatabaseConfig provides configuration for database initialization.
type DatabaseConfig struct {
	// Mode specifies the database access pattern.
	Mode DatabaseMode

	// URL is the database connection URL.
	URL string

	// VerboseMode enables verbose logging.
	VerboseMode bool

	// ContainerMode specifies testcontainer behavior ("disabled", "preferred", "required").
	ContainerMode string

	// SkipTemplateMigrations skips template infrastructure migrations (1001-1004).
	// Set to true for services like KMS that have their own migration scheme.
	SkipTemplateMigrations bool
}

// NewDefaultDatabaseConfig creates a default GORM database configuration.
func NewDefaultDatabaseConfig(url string) *DatabaseConfig {
	return &DatabaseConfig{
		Mode:                   DatabaseModeGORM,
		URL:                    url,
		VerboseMode:            false,
		ContainerMode:          "disabled",
		SkipTemplateMigrations: false,
	}
}

// NewKMSDatabaseConfig creates a database configuration suitable for KMS-style services.
func NewKMSDatabaseConfig(url string) *DatabaseConfig {
	return &DatabaseConfig{
		Mode:                   DatabaseModeRawSQL,
		URL:                    url,
		VerboseMode:            false,
		ContainerMode:          "disabled",
		SkipTemplateMigrations: true,
	}
}
