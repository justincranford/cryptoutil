package sqlrepository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilContainer "cryptoutil/internal/common/container"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

const (
	DBTypeSQLite   SupportedDBType = "sqlite"
	DBTypePostgres SupportedDBType = "postgres"

	ContainerModeDisabled  ContainerMode = "disabled"
	ContainerModePreferred ContainerMode = "preferred"
	ContainerModeRequired  ContainerMode = "required"

	firstDBPingAttemptWait = cryptoutilMagic.DBPingAttemptWait
	maxDBPingAttempts      = cryptoutilMagic.DBMaxPingAttempts
	nextDBPingAttemptWait  = cryptoutilMagic.DBNextPingAttemptWait
	sqliteBusyTimeout      = cryptoutilMagic.SQLiteBusyTimeout
	// Local constants for this provider.
	defaultPingTimeout       = cryptoutilMagic.DBPingTimeout
	sqliteMaxOpenConnections = cryptoutilMagic.SQLiteMaxOpenConnections
	randSuffixMax            = cryptoutilMagic.DBRandSuffixMax
)

type SQLRepository struct {
	telemetryService    *cryptoutilTelemetry.TelemetryService
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
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
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

type (
	SupportedDBType string
	ContainerMode   string
)

// extractSchemaFromURL extracts the schema name from PostgreSQL URL search_path parameter.
// Used for test isolation where each test gets its own schema.
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
	postgresContainerDBName = func() string {
		val, err := rand.Int(rand.Reader, big.NewInt(randSuffixMax))
		if err != nil {
			panic(fmt.Sprintf("failed to generate random database name: %v", err))
		}

		return fmt.Sprintf("keyservice%04d", val.Int64())
	}()
	postgresContainerDBUsername = func() string {
		val, err := rand.Int(rand.Reader, big.NewInt(randSuffixMax))
		if err != nil {
			panic(fmt.Sprintf("failed to generate random username: %v", err))
		}

		return fmt.Sprintf("postgresUsername%04d", val.Int64())
	}()
	postgresContainerDBPassword = func() string {
		val, err := rand.Int(rand.Reader, big.NewInt(randSuffixMax))
		if err != nil {
			panic(fmt.Sprintf("failed to generate random password: %v", err))
		}

		return fmt.Sprintf("postgresPassword%04d", val.Int64())
	}()

	ErrContainerOptionNotExist                      = errors.New("container option not available for sqlite")
	ErrUnsupportedDBType                            = errors.New("unsupported database type")
	ErrContainerModeRequiredButContainerNotStarted  = errors.New("container mode required but container didn't start")
	ErrContainerModePreferredButContainerNotStarted = errors.New("container mode preferred but container didn't start")
	ErrOpenDatabaseFailed                           = errors.New("failed to open database connection")
	ErrPingDatabaseFailed                           = errors.New("failed to ping database")
	ErrFailedDBConnection                           = errors.New("failed to connect to the database")
	ErrMaxPingAttemptsExceeded                      = errors.New("exceeded maximum DB ping attempts")
)

func NewSQLRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) (*SQLRepository, error) {
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
			containerDatabaseURL, shutdownDBContainer, err = cryptoutilContainer.StartPostgres(ctx, telemetryService, postgresContainerDBName, postgresContainerDBUsername, postgresContainerDBPassword)
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

	sqlDB, err := sql.Open(string(dbType), databaseURL)
	if err != nil {
		telemetryService.Slogger.Error("failed to open database", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))
		shutdownDBContainer()

		return nil, fmt.Errorf("failed to open database: %w", errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err)))
	}

	sqlRepository := &SQLRepository{telemetryService: telemetryService, dbType: dbType, sqlDB: sqlDB, containerMode: containerMode, shutdownDBContainer: shutdownDBContainer, verboseMode: settings.VerboseMode}

	if dbType == DBTypeSQLite {
		sqlDB.SetMaxOpenConns(sqliteMaxOpenConnections) // SQLite doesn't support concurrent writers; workaround is to limit the pool connections size, but not good for read concurrency

		if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			telemetryService.Slogger.Error("failed to enable WAL mode", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))

			return nil, fmt.Errorf("failed to enable WAL mode: %w", errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err)))
		}

		if _, err := sqlDB.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d;", int(sqliteBusyTimeout.Milliseconds()))); err != nil { // 30 seconds for concurrent testing
			telemetryService.Slogger.Error("failed to set busy timeout", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))

			return nil, fmt.Errorf("failed to set busy timeout: %w", errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err)))
		}
	} else if firstDBPingAttemptWait > 0 {
		time.Sleep(firstDBPingAttemptWait)
	}

	sqlRepository.logConnectionPoolSettings()

	for attempt, attemptsRemaining := 1, maxDBPingAttempts; attemptsRemaining > 0; attemptsRemaining-- {
		err = sqlDB.Ping()
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

			if _, err := sqlDB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)); err != nil {
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

	s.telemetryService.Slogger.Info("DB Pool Settings", "maxOpenConnections", maxOpenConnections, "openConnections", openConnections, "idle", idle, "isUse", isUse,
		"waitCount", waitCount, "waitDuration", waitDuration, "maxIdleClosed", maxIdleClosed, "maxIdleTimeClosed", maxIdleTimeClosed, "maxLifetimeClosed", maxLifetimeClosed)
}
