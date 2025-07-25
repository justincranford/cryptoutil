package sqlrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilContainer "cryptoutil/internal/common/container"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

type SqlRepository struct {
	telemetryService    *cryptoutilTelemetry.TelemetryService
	dbType              SupportedDBType // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	sqlDB               *sql.DB
	containerMode       ContainerMode
	shutdownDBContainer func()
}

type (
	SupportedDBType string
	ContainerMode   string
)

const (
	DBTypeSQLite   SupportedDBType = "sqlite"
	DBTypePostgres SupportedDBType = "postgres"

	ContainerModeDisabled  ContainerMode = "disabled"
	ContainerModePreferred ContainerMode = "preferred"
	ContainerModeRequired  ContainerMode = "required"

	firstDbPingAttemptWait = 750 * time.Millisecond
	maxDbPingAttempts      = 3
	nextDbPingAttemptWait  = 1 * time.Second
)

var (
	postgresContainerDbName     = fmt.Sprintf("keyservice%04d", rand.Intn(10_000))
	postgresContainerDbUsername = fmt.Sprintf("postgresUsername%04d", rand.Intn(10_000))
	postgresContainerDbPassword = fmt.Sprintf("postgresPassword%04d", rand.Intn(10_000))

	ErrContainerOptionNotExist                      = errors.New("container option not available for sqlite")
	ErrUnsupportedDBType                            = errors.New("unsupported database type")
	ErrContainerModeRequiredButContainerNotStarted  = errors.New("container mode required but container didn't start")
	ErrContainerModePreferredButContainerNotStarted = errors.New("container mode preferred but container didn't start")
	ErrOpenDatabaseFailed                           = errors.New("failed to open database connection")
	ErrPingDatabaseFailed                           = errors.New("failed to ping database")
	ErrFailedDBConnection                           = errors.New("failed to connect to the database")
	ErrMaxPingAttemptsExceeded                      = errors.New("exceeded maximum DB ping attempts")
)

func NewSqlRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) (*SqlRepository, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings must be non-nil")
	}

	dbType, databaseUrl, err := mapDbTypeAndUrl(telemetryService, settings.DevMode, settings.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to determine database type and URL: %w", err)
	}

	containerMode, err := mapContainerMode(telemetryService, settings.DatabaseContainer)
	if err != nil {
		return nil, fmt.Errorf("failed to determine container mode: %w", err)
	}

	var shutdownDBContainer func() = func() {}  // no-op by default
	if containerMode != ContainerModeDisabled { // containerMode is required or preferred
		telemetryService.Slogger.Debug("containerMode is not disabled, so trying to start a container", "dbType", string(dbType), "containerMode", string(containerMode))
		var containerDatabaseUrl string
		var err error
		switch dbType {
		case DBTypeSQLite:
			return nil, ErrContainerOptionNotExist
		case DBTypePostgres:
			containerDatabaseUrl, shutdownDBContainer, err = cryptoutilContainer.StartPostgres(ctx, telemetryService, postgresContainerDbName, postgresContainerDbUsername, postgresContainerDbPassword)
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedDBType, dbType)
		}
		// Example errors: Rootless Docker not supported (Windows), Docker Desktop not running (Windows), Docker not installed (Linux/Mac), etc.
		if err == nil { // success
			telemetryService.Slogger.Debug("successfully started database container", "containerMode", string(containerMode), "dbType", string(dbType), "databaseUrl", containerDatabaseUrl)
			databaseUrl = containerDatabaseUrl
		} else if containerMode == ContainerModeRequired { // container is required, so this error is fatal; give up and return the errors
			telemetryService.Slogger.Warn("failed to start database container", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrContainerModeRequiredButContainerNotStarted, err))
			return nil, errors.Join(ErrContainerModeRequiredButContainerNotStarted, fmt.Errorf("dbType: %s", string(dbType)))
		} else { // container was preferred, so this error not is fatal; fall through and try to connect with the provided databaseUrl parameter
			telemetryService.Slogger.Warn("failed to start database container", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrContainerModePreferredButContainerNotStarted, err))
		}
	}

	sqlDB, err := sql.Open(string(dbType), databaseUrl)
	if err != nil {
		telemetryService.Slogger.Error("failed to open database", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))
		shutdownDBContainer()
		return nil, errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err))
	}

	sqlRepository := &SqlRepository{telemetryService: telemetryService, dbType: dbType, sqlDB: sqlDB, containerMode: containerMode, shutdownDBContainer: shutdownDBContainer}

	if dbType == DBTypeSQLite {
		sqlDB.SetMaxOpenConns(1) // SQLite doesn't support concurrent writers; workaround is to limit the pool connections size to 1, but not good for read concurrency
		if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			telemetryService.Slogger.Error("failed to enable WAL mode", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))
			return nil, errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err))
		}
		if _, err := sqlDB.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
			telemetryService.Slogger.Error("failed to set busy timeout", "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrOpenDatabaseFailed, err))
			return nil, errors.Join(ErrOpenDatabaseFailed, fmt.Errorf("dbType: %s, %w", string(dbType), err))
		}
	} else if firstDbPingAttemptWait > 0 {
		time.Sleep(firstDbPingAttemptWait)
	}
	sqlRepository.logConnectionPoolSettings()

	for attempt, attemptsRemaining := 1, maxDbPingAttempts; attemptsRemaining > 0; attemptsRemaining-- {
		err = sqlDB.Ping()
		if err == nil {
			telemetryService.Slogger.Debug("successfully pinged database", "attempt", attempt, "containerMode", string(containerMode), "dbType", string(dbType))
			break
		}
		telemetryService.Slogger.Warn("failed to ping database", "attempt", attempt, "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrPingDatabaseFailed, err))
		attempt++
		if attemptsRemaining > 0 {
			time.Sleep(nextDbPingAttemptWait)
		}
	}

	if err != nil {
		telemetryService.Slogger.Warn("giving up trying to ping database", "attempts", maxDbPingAttempts, "containerMode", string(containerMode), "dbType", string(dbType), "error", errors.Join(ErrPingDatabaseFailed, err))
		sqlRepository.Shutdown()
		return nil, errors.Join(ErrPingDatabaseFailed, fmt.Errorf("dbType: %s", string(dbType)))
	}

	return sqlRepository, nil
}

func (s *SqlRepository) Shutdown() {
	s.telemetryService.Slogger.Debug("shutting down SQL Provider")
	s.shutdownDBContainer() // This call does it's own logging
	s.telemetryService.Slogger.Debug("shutting down SQL Connection")
	if err := s.sqlDB.Close(); err != nil {
		s.telemetryService.Slogger.Error("failed to close SQL DB", "error", err)
	}
}

func (s *SqlRepository) logConnectionPoolSettings() {
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
