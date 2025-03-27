package sqlprovider

import (
	"context"
	cryptoutilContainer "cryptoutil/internal/container"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type SupportedSqlDB string
type ContainerMode string

const (
	SupportedSqlDBSQLite   SupportedSqlDB = "sqlite"
	SupportedSqlDBPostgres SupportedSqlDB = "postgres"

	ContainerModeDisabled  ContainerMode = "disabled"
	ContainerModePreferred ContainerMode = "preferred"
	ContainerModeRequired  ContainerMode = "required"

	maxDbConnectAttempts = 3
)

var (
	postgresContainerDbName     = fmt.Sprintf("keyservice%04d", rand.Intn(10_000))
	postgresContainerDbUsername = fmt.Sprintf("postgresUsername%04d", rand.Intn(10_000))
	postgresContainerDbPassword = fmt.Sprintf("postgresPassword%04d", rand.Intn(10_000))
)

func CreateSqlDB(ctx context.Context, dbType SupportedSqlDB, databaseUrl string, containerMode ContainerMode) (*sql.DB, func(), error) {
	var shutdownDBContainer func() = func() {} // no-op by default

	if containerMode != ContainerModeDisabled { // containerMode is required or preferred
		log.Printf("Container mode is %s, trying to start a %s container", string(dbType), string(containerMode))
		var containerDatabaseUrl string
		var err error
		switch dbType {
		case SupportedSqlDBSQLite:
			return nil, nil, fmt.Errorf("there is no container option for sqlite")
		case SupportedSqlDBPostgres:
			containerDatabaseUrl, shutdownDBContainer, err = cryptoutilContainer.StartPostgres(ctx, postgresContainerDbName, postgresContainerDbUsername, postgresContainerDbPassword)
		default:
			return nil, nil, fmt.Errorf("unsupported database type: %s", dbType)
		}
		// Example errors: Rootless Docker not supported (Windows), Docker Desktop not running (Windows), Docker not installed (Linux/Mac), etc.
		if err == nil { // success
			log.Printf("containerMode was %s, and container started successfully, so using generated %s database URL: %s", string(containerMode), string(dbType), containerDatabaseUrl)
			databaseUrl = containerDatabaseUrl
		} else if containerMode == ContainerModeRequired { // container is required, so this error is fatal; give up and return the errors
			return nil, nil, fmt.Errorf("containerMode was required, but failed to start %s container: %w", string(dbType), err)
		} else { // container was required, so this error not is fatal; fall through and try to connect with the provided databaseUrl parameter
			log.Printf("containerMode was preferred, but failed to start, so use the provided %s database URL instead: %v", string(dbType), err)
		}
	}

	sqlDB, err := sql.Open(string(dbType), databaseUrl)
	if err != nil {
		shutdownDBContainer()
		return nil, nil, fmt.Errorf("failed to open %s database: %w", string(dbType), err)
	}

	for attempt, attemptsRemaining := 1, maxDbConnectAttempts; attemptsRemaining > 0; attemptsRemaining-- {
		err = sqlDB.Ping()
		if err == nil {
			log.Printf("ping SQL DB attempt %d succeeded", attempt)
			break
		}
		log.Printf("ping SQL DB attempt %d failed: %v", attempt, err)
		attempt++
		if attemptsRemaining > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	if err != nil {
		log.Printf("giving up trying to get SQL")
		shutdownDBContainer()
		return nil, nil, fmt.Errorf("gave up trying to get SQL DB: %w", err)
	}

	return sqlDB, shutdownDBContainer, nil
}
