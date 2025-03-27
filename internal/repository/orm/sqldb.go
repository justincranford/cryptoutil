package orm

import (
	"context"
	cryptoutilContainer "cryptoutil/internal/container"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type DBType string
type ContainerMode string

const (
	DBTypeSQLite   DBType = "sqlite"
	DBTypePostgres DBType = "postgres"

	ContainerModeRequired  ContainerMode = "required"
	ContainerModePreferred ContainerMode = "preferred"
	ContainerModeDisabled  ContainerMode = "disabled"

	maxDbConnectAttempts = 3
)

var (
	postgresContainerDbNameDefault = fmt.Sprintf("keyservice%04d", rand.Intn(10_000))
	postgresContainerDbUsername    = fmt.Sprintf("postgresUsername%04d", rand.Intn(10_000))
	postgresContainerDbPassword    = fmt.Sprintf("postgresPassword%04d", rand.Intn(10_000))
)

// TODO Move this to different package

func CreateSqlDB(ctx context.Context, dbType DBType, databaseUrl string, containerMode ContainerMode) (*sql.DB, func(), error) {
	var shutdownDBContainer func() = func() {} // no-op by default

	if containerMode != ContainerModeDisabled {
		log.Printf("Container mode is %s, trying to start a %s container", string(dbType), string(containerMode)) // containerMode is required or preferred
		var containerDatabaseUrl string
		var err error
		switch dbType {
		case DBTypeSQLite:
			return nil, nil, fmt.Errorf("there is no container option for sqlite")
		case DBTypePostgres:
			containerDatabaseUrl, shutdownDBContainer, err = cryptoutilContainer.StartPostgres(ctx, postgresContainerDbNameDefault, postgresContainerDbUsername, postgresContainerDbPassword)
		default:
			return nil, nil, fmt.Errorf("unsupported database type: %s", dbType)
		}
		if err == nil { // Example> Docker not installed, Docker Desktop not running, etc.
			log.Printf("containerMode was %s, and container started successfully, so using generated %s database URL: %s", string(containerMode), string(dbType), containerDatabaseUrl)
			databaseUrl = containerDatabaseUrl
		} else if containerMode == ContainerModeRequired { // give up and return the error
			return nil, nil, fmt.Errorf("containerMode was required, but failed to start %s container: %w", string(dbType), err)
		} else {
			log.Printf("containerMode was preferred, but failed to start, so use the provided %s database URL instead: %v", string(dbType), err)
		}
	}

	sqlDB, err := sql.Open(string(dbType), databaseUrl)
	if err != nil {
		shutdownDBContainer()
		return nil, nil, fmt.Errorf("failed to open %s database: %w", string(DBTypeSQLite), err)
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
