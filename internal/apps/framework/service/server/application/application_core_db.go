// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// provisionDatabase handles all database provisioning scenarios:
// 1. Internal managed SQLite instance (file::memory:?cache=shared)
// 2. Internal managed PostgreSQL testcontainer (DatabaseContainer=required/preferred)
// 3. External DB connection (postgres:// or file:// scheme).
func provisionDatabase(ctx context.Context, basic *Basic, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*gorm.DB, func(), error) {
	return provisionDatabaseInternal(ctx, basic, settings, cryptoutilSharedContainer.StartPostgres, sql.Open, func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
		return gorm.Open(dialector, config)
	})
}

func provisionDatabaseInternal(
	ctx context.Context,
	basic *Basic,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	startPostgresFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, dbName, username, password string) (string, func(), error),
	sqlOpenFn func(driverName, dataSourceName string) (*sql.DB, error),
	gormOpenSQLiteFn func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error),
) (*gorm.DB, func(), error) {
	databaseURL := settings.DatabaseURL
	containerMode := settings.DatabaseContainer

	shutdownContainer := func() {} // No-op by default.

	// Determine database type from URL.
	var isSQLite bool

	var isPostgres bool

	if databaseURL == "" || databaseURL == cryptoutilSharedMagic.SQLiteInMemoryDSN || databaseURL == cryptoutilSharedMagic.SQLiteMemoryPlaceholder {
		isSQLite = true
		databaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN // Normalize SQLite in-memory URL.
	} else if len(databaseURL) >= 9 && databaseURL[:9] == "postgres:" {
		isPostgres = true
	} else if len(databaseURL) >= 9 && databaseURL[:9] == "sqlite://" {
		// Handle sqlite:// scheme (e.g., sqlite://file::memory:?cache=shared).
		isSQLite = true
		databaseURL = databaseURL[9:] // Strip sqlite:// prefix -> file::memory:?cache=shared.
	} else if len(databaseURL) >= cryptoutilSharedMagic.GitRecentActivityDays && databaseURL[:cryptoutilSharedMagic.GitRecentActivityDays] == cryptoutilSharedMagic.FileURIScheme {
		isSQLite = true
	} else if len(databaseURL) >= 13 && databaseURL[:13] == "file::memory:" {
		// Handle file::memory:NAME?cache=shared format.
		// Normalize to file:NAME?mode=memory&cache=shared for modernc.org/sqlite compatibility.
		isSQLite = true

		name := strings.TrimPrefix(databaseURL, "file::memory:")
		if idx := strings.Index(name, "?"); idx != -1 {
			name = name[:idx]
		}

		if name != "" {
			databaseURL = fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
		}
	} else if strings.HasPrefix(databaseURL, "file:") && strings.Contains(databaseURL, "mode=memory") {
		// Handle file:NAME?mode=memory&cache=shared format (unique per-test in-memory SQLite).
		isSQLite = true
	} else {
		return nil, nil, fmt.Errorf("unsupported database URL scheme: %s", databaseURL)
	}

	// Handle PostgreSQL testcontainer provisioning.
	if isPostgres && containerMode != "" && containerMode != cryptoutilSharedMagic.DefaultDatabaseContainerDisabled {
		basic.TelemetryService.Slogger.Debug("attempting to start PostgreSQL testcontainer", "containerMode", containerMode)

		containerURL, cleanup, err := startPostgresSafely(ctx, basic, startPostgresFn, "test_db", "test_user", "test_password")
		if err == nil {
			basic.TelemetryService.Slogger.Info("successfully started PostgreSQL testcontainer", "containerURL", containerURL)
			databaseURL = containerURL
			shutdownContainer = cleanup
		} else if containerMode == "required" {
			basic.TelemetryService.Slogger.Error("failed to start required PostgreSQL testcontainer", cryptoutilSharedMagic.StringError, err)

			return nil, nil, fmt.Errorf("failed to start required PostgreSQL testcontainer: %w", err)
		} else {
			basic.TelemetryService.Slogger.Warn("failed to start preferred PostgreSQL testcontainer, falling back to external DB", cryptoutilSharedMagic.StringError, err)
		}
	}

	// Open database connection.
	var db *gorm.DB

	var err error

	if isSQLite {
		basic.TelemetryService.Slogger.Debug("opening SQLite database", "url", databaseURL)
		db, err = openSQLiteInternal(ctx, databaseURL, settings.VerboseMode, sqlOpenFn, gormOpenSQLiteFn)
	} else {
		basic.TelemetryService.Slogger.Debug("opening PostgreSQL database", "url", maskPassword(databaseURL))
		db, err = openPostgreSQL(ctx, databaseURL, settings.VerboseMode, settings.DatabaseSSLMode, settings.DatabaseSSLCert, settings.DatabaseSSLKey, settings.DatabaseSSLRootCert)
	}

	if err != nil {
		shutdownContainer()

		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	basic.TelemetryService.Slogger.Info("database connection established successfully")

	return db, shutdownContainer, nil
}

func startPostgresSafely(
	ctx context.Context,
	basic *Basic,
	startPostgresFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, dbName, username, password string) (string, func(), error),
	dbName, username, password string,
) (
	containerURL string,
	cleanup func(),
	err error,
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic while starting PostgreSQL testcontainer: %v", recovered)
		}
	}()

	return startPostgresFn(ctx, basic.TelemetryService, dbName, username, password)
}

// openSQLite opens a SQLite database connection with GORM and configures WAL mode.
func openSQLite(ctx context.Context, databaseURL string, debugMode bool) (*gorm.DB, error) {
	return openSQLiteInternal(ctx, databaseURL, debugMode, sql.Open, func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
		return gorm.Open(dialector, config)
	})
}

func openSQLiteInternal(
	ctx context.Context,
	databaseURL string,
	debugMode bool,
	sqlOpenFn func(driverName, dataSourceName string) (*sql.DB, error),
	gormOpenSQLiteFn func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error),
) (*gorm.DB, error) {
	// Open database connection using database/sql.
	sqlDB, err := sqlOpenFn("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure SQLite for concurrent operations.
	// Note: Skip WAL mode for in-memory databases as it's not supported.
	// Matches: ":memory:", "file::memory:?cache=shared", "file::memory:NAME?cache=shared" (unique per-test)
	isInMemory := databaseURL == cryptoutilSharedMagic.SQLiteMemoryPlaceholder || strings.HasPrefix(databaseURL, "file::memory:") ||
		strings.Contains(databaseURL, "mode=memory")

	if !isInMemory {
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
			_ = sqlDB.Close()

			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Wrap with GORM.
	dialector := sqlite.Dialector{Conn: sqlDB}

	gormConfig := &gorm.Config{SkipDefaultTransaction: true}
	if debugMode {
		gormConfig.Logger = logger.Default.LogMode(cryptoutilSharedMagic.GormLogModeInfo)
	}

	db, err := gormOpenSQLiteFn(dialector, gormConfig)
	if err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	// Configure connection pool for GORM transactions.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: never close connections.

	return db, nil
}

// openPostgreSQL opens a PostgreSQL database connection with GORM.
// sslMode, sslCert, sslKey, sslRootCert are optional mTLS parameters (Cat 10/14).
// When sslMode is non-empty it is appended to the DSN; cert/key/rootcert are appended if non-empty.
func openPostgreSQL(_ context.Context, databaseURL string, debugMode bool, sslMode, sslCert, sslKey, sslRootCert string) (*gorm.DB, error) {
	if sslMode != "" {
		sep := "?"
		if strings.Contains(databaseURL, "?") {
			sep = "&"
		}

		databaseURL += sep + "sslmode=" + sslMode

		if sslCert != "" {
			databaseURL += "&sslcert=" + sslCert
		}

		if sslKey != "" {
			databaseURL += "&sslkey=" + sslKey
		}

		if sslRootCert != "" {
			databaseURL += "&sslrootcert=" + sslRootCert
		}
	}

	gormConfig := &gorm.Config{SkipDefaultTransaction: true}
	if debugMode {
		gormConfig.Logger = logger.Default.LogMode(cryptoutilSharedMagic.GormLogModeInfo)
	}

	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Configure connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.PostgreSQLMaxOpenConns)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.PostgreSQLMaxIdleConns)

	return db, nil
}

// maskPassword masks the password in a PostgreSQL connection string for logging.
func maskPassword(dsn string) string {
	// Simple masking: replace password with "***".
	// Format: postgres://user:password@host:port/db
	// This is a naive implementation; production code should use url.Parse.
	start := 0

	for i := 0; i < len(dsn); i++ {
		if dsn[i] == ':' && i > 0 && dsn[i-1] == '/' {
			start = i + 1

			break
		}
	}

	if start == 0 {
		return dsn
	}

	end := start

	for i := start; i < len(dsn); i++ {
		if dsn[i] == '@' {
			end = i

			break
		}
	}

	if end == start {
		return dsn
	}

	return dsn[:start] + "***" + dsn[end:]
}
