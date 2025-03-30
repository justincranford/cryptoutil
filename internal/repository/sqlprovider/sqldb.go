package sqlprovider

import (
	"context"
	cryptoutilContainer "cryptoutil/internal/container"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

type SqlProvider struct {
	telemetryService    *cryptoutilTelemetry.Service
	dbType              SupportedSqlDB
	sqlDB               *sql.DB
	containerMode       ContainerMode
	shutdownDBContainer func()
}

type SupportedSqlDB string
type ContainerMode string

const (
	SupportedSqlDBSQLite   SupportedSqlDB = "sqlite"
	SupportedSqlDBPostgres SupportedSqlDB = "postgres"

	ContainerModeDisabled  ContainerMode = "disabled"
	ContainerModePreferred ContainerMode = "preferred"
	ContainerModeRequired  ContainerMode = "required"

	maxDbConnectAttempts = 3
	waitBeforeNextPing   = 1 * time.Second
)

var (
	postgresContainerDbName     = fmt.Sprintf("keyservice%04d", rand.Intn(10_000))
	postgresContainerDbUsername = fmt.Sprintf("postgresUsername%04d", rand.Intn(10_000))
	postgresContainerDbPassword = fmt.Sprintf("postgresPassword%04d", rand.Intn(10_000))
)

func NewSqlProvider(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, dbType SupportedSqlDB, databaseUrl string, containerMode ContainerMode) (*SqlProvider, error) {
	var shutdownDBContainer func() = func() {} // no-op by default

	if containerMode != ContainerModeDisabled { // containerMode is required or preferred
		telemetryService.Slogger.Debug("containerMode is not disabled, so trying to start a container", "dbType", string(dbType), "containerMode", string(containerMode))
		var containerDatabaseUrl string
		var err error
		switch dbType {
		case SupportedSqlDBSQLite:
			return nil, fmt.Errorf("there is no container option for sqlite")
		case SupportedSqlDBPostgres:
			containerDatabaseUrl, shutdownDBContainer, err = cryptoutilContainer.StartPostgres(ctx, telemetryService, postgresContainerDbName, postgresContainerDbUsername, postgresContainerDbPassword)
		default:
			return nil, fmt.Errorf("unsupported database type: %s", dbType)
		}
		// Example errors: Rootless Docker not supported (Windows), Docker Desktop not running (Windows), Docker not installed (Linux/Mac), etc.
		if err == nil { // success
			telemetryService.Slogger.Debug("containerMode was %s, and container started successfully, so using generated %s database URL: %s", "containerMode", string(containerMode), "dbType", string(dbType), "containerDatabaseUrl", containerDatabaseUrl)
			databaseUrl = containerDatabaseUrl
		} else if containerMode == ContainerModeRequired { // container is required, so this error is fatal; give up and return the errors
			return nil, fmt.Errorf("containerMode was required, but failed to start %s container: %w", string(dbType), err)
		} else { // container was required, so this error not is fatal; fall through and try to connect with the provided databaseUrl parameter
			telemetryService.Slogger.Warn("containerMode was preferred, but failed to start, so use the provided %s database URL instead: %v", "dbType", string(dbType), "error", err)
		}
	}

	sqlDB, err := sql.Open(string(dbType), databaseUrl)
	if err != nil {
		shutdownDBContainer()
		return nil, fmt.Errorf("failed to open %s database: %w", string(dbType), err)
	}

	sqlProvider := &SqlProvider{telemetryService: telemetryService, dbType: dbType, sqlDB: sqlDB, containerMode: containerMode, shutdownDBContainer: shutdownDBContainer}

	for attempt, attemptsRemaining := 1, maxDbConnectAttempts; attemptsRemaining > 0; attemptsRemaining-- {
		err = sqlDB.Ping()
		if err == nil {
			telemetryService.Slogger.Debug("ping SQL DB succeeded", "attempt", attempt)
			break
		}
		telemetryService.Slogger.Warn("ping SQL DB failed", "attempt", attempt, "error", err)
		attempt++
		if attemptsRemaining > 0 {
			time.Sleep(waitBeforeNextPing)
		}
	}

	if err != nil {
		telemetryService.Slogger.Error("giving up trying to get SQL")
		sqlProvider.Shutdown()
		return nil, fmt.Errorf("gave up trying to get SQL DB: %w", err)
	}

	return sqlProvider, nil
}

func (s *SqlProvider) SqlDB() *sql.DB {
	return s.sqlDB
}

func (s *SqlProvider) Shutdown() {
	s.telemetryService.Slogger.Error("shutting down SQL Provider")
	s.shutdownDBContainer() // This call does it's own logging
	s.telemetryService.Slogger.Error("shutting down SQL connection")
	if err := s.sqlDB.Close(); err != nil {
		s.telemetryService.Slogger.Error("failed to close SQL DB", "error", err)
	}
}
