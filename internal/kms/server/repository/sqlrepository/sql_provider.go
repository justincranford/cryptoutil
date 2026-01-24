// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver registration
)

// SupportedDBType represents a supported database type.
type SupportedDBType string

// ContainerMode represents the testcontainer mode for database initialization.
type ContainerMode string

const (
	// DBTypeSQLite represents SQLite database type.
	DBTypeSQLite SupportedDBType = "sqlite"
	// DBTypePostgres represents PostgreSQL database type.
	DBTypePostgres SupportedDBType = "pgx"

	// ContainerModeDisabled disables testcontainer mode.
	ContainerModeDisabled ContainerMode = "disabled"
	// ContainerModePreferred prefers testcontainer mode if available.
	ContainerModePreferred ContainerMode = "preferred"
	// ContainerModeRequired requires testcontainer mode.
	ContainerModeRequired ContainerMode = "required"

	firstDBPingAttemptWait = cryptoutilSharedMagic.DBPingFirstAttemptWait
	maxDBPingAttempts      = cryptoutilSharedMagic.DBMaxPingAttempts
	nextDBPingAttemptWait  = cryptoutilSharedMagic.DBPingNextAttemptWait
	sqliteBusyTimeout      = cryptoutilSharedMagic.DBSQLiteBusyTimeout
	// Local constants for this provider.
	sqliteMaxOpenConnections = cryptoutilSharedMagic.SQLiteMaxOpenConnections
)

// SQLRepository provides database operations using database/sql.
type SQLRepository struct {
	telemetryService    *cryptoutilSharedTelemetry.TelemetryService
	dbType              SupportedDBType // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	sqlDB               *sql.DB
	containerMode       ContainerMode
	verboseMode         bool
	shutdownDBContainer func()
}

// GetDBType returns the database type.
func (s *SQLRepository) GetDBType() SupportedDBType {
	return s.dbType
}

// HealthCheck performs a database connectivity check and returns detailed status.
func (s *SQLRepository) HealthCheck(ctx context.Context) (map[string]any, error) {
	if s.sqlDB == nil {
		return map[string]any{
			"status": "error",
			"error":  "database connection not initialized",
		}, fmt.Errorf("database connection not initialized")
	}

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.DBPingTimeout)
	defer cancel()

	err := s.sqlDB.PingContext(pingCtx)
	if err != nil {
		return map[string]any{
			"status":  "error",
			"error":   fmt.Sprintf("database ping failed: %v", err),
			"db_type": string(s.GetDBType()),
		}, fmt.Errorf("database ping failed: %w", err)
	}

	// Get connection pool stats
	stats := s.sqlDB.Stats()

	return map[string]any{
		"status":               "ok",
		"db_type":              string(s.GetDBType()),
		"open_connections":     stats.OpenConnections,
		"idle_connections":     stats.Idle,
		"in_use_connections":   stats.InUse,
		"max_open_connections": stats.MaxOpenConnections,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}, nil
}

func extractSchemaFromURL(databaseURL string) string {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return ""
	}

	searchPath := parsedURL.Query().Get("search_path")
	if searchPath == "" {
		return ""
	}

	// Extract first schema name (search_path can contain multiple schemas separated by comma).
	if commaIndex := strings.Index(searchPath, ","); commaIndex != -1 {
		return strings.TrimSpace(searchPath[:commaIndex])
	}

	return strings.TrimSpace(searchPath)
}

var (
	postgresContainerRandSuffix = func() string {
		postgresContainerRandSuffix, err := crand.Int(crand.Reader, big.NewInt(cryptoutilSharedMagic.DBContainerRandSuffixMax))
		if err != nil {
			panic(fmt.Sprintf("failed to generate random database name: %v", err))
		}

		return strconv.FormatInt(postgresContainerRandSuffix.Int64(), 36)
	}()
	postgresContainerDBName     = "postgresDatabase" + postgresContainerRandSuffix
	postgresContainerDBUsername = "postgresUsername" + postgresContainerRandSuffix
	postgresContainerDBPassword = "postgresPassword" + postgresContainerRandSuffix

	// ErrContainerOptionNotExist indicates container option is not available for SQLite.
	ErrContainerOptionNotExist = errors.New("container option not available for sqlite")
	// ErrUnsupportedDBType indicates an unsupported database type was specified.
	ErrUnsupportedDBType = errors.New("unsupported database type")
	// ErrContainerModeRequiredButContainerNotStarted indicates container was required but failed to start.
	ErrContainerModeRequiredButContainerNotStarted = errors.New("container mode required but container didn't start")
	// ErrContainerModePreferredButContainerNotStarted indicates container was preferred but failed to start.
	ErrContainerModePreferredButContainerNotStarted = errors.New("container mode preferred but container didn't start")
	// ErrOpenDatabaseFailed indicates the database connection could not be opened.
	ErrOpenDatabaseFailed = errors.New("failed to open database connection")
	// ErrPingDatabaseFailed indicates the database ping failed.
	ErrPingDatabaseFailed = errors.New("failed to ping database")
	// ErrFailedDBConnection indicates a failure to connect to the database.
	ErrFailedDBConnection = errors.New("failed to connect to the database")
	// ErrMaxPingAttemptsExceeded indicates the maximum ping attempts were exceeded.
	ErrMaxPingAttemptsExceeded = errors.New("exceeded maximum DB ping attempts")
)

// NewSQLRepository creates a new SQLRepository with the given configuration.
func NewSQLRepository(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*SQLRepository, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings must be non-nil")
	}

	dbType, databaseURL, err := mapDBTypeAndURL(telemetryService, settings.DevMode, settings.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to determine database type and URL: %w", err)
	}

	containerMode, err := mapContainerMode(telemetryService, settings.DatabaseContainer)
	if err != nil {
		return nil, fmt.Errorf("failed to determine container mode: %w", err)
	}

	shutdownDBContainer := func() {} // no-op by default

	if containerMode != ContainerModeDisabled { // containerMode is required or preferred
		telemetryService.Slogger.Debug("containerMode is not disabled, so trying to start a container", "dbType", string(dbType), "containerMode", string(containerMode))

		var containerDatabaseURL string

		var err error

		switch dbType {
		case DBTypeSQLite:
			return nil, ErrContainerOptionNotExist
		case DBTypePostgres:
			containerDatabaseURL, shutdownDBContainer, err = cryptoutilSharedContainer.StartPostgres(ctx, telemetryService, postgresContainerDBName, postgresContainerDBUsername, postgresContainerDBPassword)
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedDBType, dbType)
		}
		// Example errors: Rootless Docker not supported (Windows), Docker Desktop not running (Windows), Docker not installed (Linux/Mac), etc.
		if err == nil { // success
			telemetryService.Slogger.Debug("successfully started database container", "containerMode", string(containerMode), "dbType", string(dbType), "databaseUrl", containerDatabaseURL)
			databaseURL = containerDatabaseURL
		} else if containerMode == ContainerModeRequired { // container is required, so this error is fatal; give up and return the errors
			telemetryService.Slogger.Warn("failed to start database container", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrContainerModeRequiredButContainerNotStarted, err))

			return nil, fmt.Errorf("failed to start required database container: %w", errors.Join(ErrContainerModeRequiredButContainerNotStarted, fmt.Errorf("dbType: %s", string(dbType))))
		} else { // container was preferred, so this error not is fatal; fall through and try to connect with the provided databaseUrl parameter
			telemetryService.Slogger.Warn("failed to start database container", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrContainerModePreferredButContainerNotStarted, err))
		}
	}

	telemetryService.Slogger.Debug("loaded database drivers", "drivers", sql.Drivers())

	sqlDB, err := sql.Open(string(dbType), databaseURL)
	if err != nil {
		telemetryService.Slogger.Error("failed to open database", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))
		shutdownDBContainer()

		return nil, fmt.Errorf("failed to open database: %w", errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err)))
	}

	sqlRepository := &SQLRepository{telemetryService: telemetryService, dbType: dbType, sqlDB: sqlDB, containerMode: containerMode, shutdownDBContainer: shutdownDBContainer, verboseMode: settings.VerboseMode}

	if dbType == DBTypeSQLite {
		sqlDB.SetMaxOpenConns(sqliteMaxOpenConnections) // SQLite doesn't support concurrent writers; workaround is to limit the pool connections size, but not good for read concurrency

		if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
			telemetryService.Slogger.Error("failed to enable WAL mode", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))

			return nil, fmt.Errorf("failed to enable WAL mode: %w", errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err)))
		}

		if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d;", int(sqliteBusyTimeout.Milliseconds()))); err != nil { // 30 seconds for concurrent testing
			telemetryService.Slogger.Error("failed to set busy timeout", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))

			return nil, fmt.Errorf("failed to set busy timeout: %w", errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err)))
		}
	} else if firstDBPingAttemptWait > 0 {
		time.Sleep(firstDBPingAttemptWait)
	}

	sqlRepository.logConnectionPoolSettings()

	for attempt, attemptsRemaining := 1, maxDBPingAttempts; attemptsRemaining > 0; attemptsRemaining-- {
		err = sqlDB.PingContext(ctx)
		if err == nil {
			telemetryService.Slogger.Debug("successfully pinged database", "attempt", attempt, "containerMode", string(containerMode), "dbType", string(dbType))

			break
		}

		telemetryService.Slogger.Warn("failed to ping database", "attempt", attempt, "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrPingDatabaseFailed, err))

		attempt++

		if attemptsRemaining > 0 {
			time.Sleep(nextDBPingAttemptWait)
		}
	}

	if err != nil {
		telemetryService.Slogger.Warn("giving up trying to ping database", "attempts", maxDBPingAttempts, "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrPingDatabaseFailed, err))
		sqlRepository.Shutdown()

		return nil, fmt.Errorf("failed to ping database: %w", errors.Join(ErrPingDatabaseFailed, fmt.Errorf("dbType: %s", string(dbType))))
	}

	telemetryService.Slogger.Debug("applying migrations")

	// For PostgreSQL test isolation: create schema if search_path is specified.
	if dbType == DBTypePostgres {
		if schemaName := extractSchemaFromURL(databaseURL); schemaName != "" {
			telemetryService.Slogger.Debug("creating test schema for PostgreSQL", "schema", schemaName)

			if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)); err != nil {
				return nil, fmt.Errorf("failed to create test schema %s: %w", schemaName, err)
			}

			telemetryService.Slogger.Debug("test schema created successfully", "schema", schemaName)
		}
	}

	err = ApplyEmbeddedSQLMigrations(telemetryService, sqlDB, sqlRepository.GetDBType())
	if err != nil {
		return nil, fmt.Errorf("failed to apply SQL migrations: %w", err)
	}

	telemetryService.Slogger.Debug("migrations completed successfully")

	err = LogSchema(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to log schemas: %w", err)
	}

	return sqlRepository, nil
}

// Shutdown closes the database connection and stops any container.
func (s *SQLRepository) Shutdown() {
	s.telemetryService.Slogger.Debug("shutting down SQL Provider")
	s.shutdownDBContainer() // This call does it's own logging
	s.telemetryService.Slogger.Debug("shutting down SQL Connection")

	if err := s.sqlDB.Close(); err != nil {
		s.telemetryService.Slogger.Error("failed to close SQL DB", "error", err)
	}
}

func (s *SQLRepository) logConnectionPoolSettings() {
	sqlDBStats := s.sqlDB.Stats()

	maxOpenConnections := sqlDBStats.MaxOpenConnections
	openConnections := sqlDBStats.OpenConnections
	idle := sqlDBStats.Idle
	isUse := sqlDBStats.InUse
	waitCount := sqlDBStats.WaitCount
	waitDuration := sqlDBStats.WaitDuration
	maxIdleClosed := sqlDBStats.MaxIdleClosed
	maxIdleTimeClosed := sqlDBStats.MaxIdleTimeClosed
	maxLifetimeClosed := sqlDBStats.MaxLifetimeClosed

	if s.verboseMode {
		s.telemetryService.Slogger.Info("DB Pool Settings", "maxOpenConnections", maxOpenConnections, "openConnections", openConnections, "idle", idle, "isUse", isUse,
			"waitCount", waitCount, "waitDuration", waitDuration, "maxIdleClosed", maxIdleClosed, "maxIdleTimeClosed", maxIdleTimeClosed, "maxLifetimeClosed", maxLifetimeClosed)
	}
}
