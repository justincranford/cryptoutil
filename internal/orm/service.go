package orm

import (
	"context"
	openapiContainer "cryptoutil/internal/container"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type Service struct {
	GormDB              *gorm.DB
	SqlDB               *sql.DB
	shutdownDBContainer func()
}

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
	dbNameDefault = fmt.Sprintf("keyservice%04d", rand.Intn(10_000))
	dbUsername    = fmt.Sprintf("postgresUsername%04d", rand.Intn(10_000))
	dbPassword    = fmt.Sprintf("postgresPassword%04d", rand.Intn(10_000))
)

func NewService(ctx context.Context, dbType DBType, databaseUrl string, containerMode ContainerMode, applyMigrations bool) (*Service, error) {
	sqlDB, shutdownDBContainer, err := createSqlDB(ctx, dbType, databaseUrl, containerMode)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	gormDB, err := createGormDB(dbType, sqlDB)
	if err != nil {
		shutdownDBContainer()
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		log.Printf("Applying %s migrations", string(dbType))
		err = gormDB.AutoMigrate(ormTableStructs...)
		if err != nil {
			shutdownDBContainer()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		log.Printf("Skipping %s migrations", string(dbType))
	}

	return &Service{SqlDB: sqlDB, GormDB: gormDB, shutdownDBContainer: shutdownDBContainer}, nil
}

func (s *Service) Shutdown() {
	err := s.SqlDB.Close() // shutdown the underlying SQL DB connection
	if err == nil {
		log.Printf("failed to close DB: %v", err)
	}
	if s.shutdownDBContainer != nil { // terminate the DB container, if applicable
		s.shutdownDBContainer()
	}
}

func createSqlDB(ctx context.Context, dbType DBType, databaseUrl string, containerMode ContainerMode) (*sql.DB, func(), error) {
	var shutdownDBContainer func() = func() {} // no-op by default

	if containerMode != ContainerModeDisabled {
		log.Printf("Container mode is %s, trying to start a %s container", string(dbType), string(containerMode)) // containerMode is required or preferred
		var containerDatabaseUrl string
		var err error
		switch dbType {
		case DBTypeSQLite:
			return nil, nil, fmt.Errorf("there is no container option for sqlite")
		case DBTypePostgres:
			containerDatabaseUrl, shutdownDBContainer, err = openapiContainer.StartPostgres(ctx, dbNameDefault, dbUsername, dbPassword)
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

func createGormDB(dbType DBType, sqlDB *sql.DB) (*gorm.DB, error) {
	var gormDialector gorm.Dialector
	switch dbType {
	case DBTypeSQLite:
		gormDialector = sqlite.Dialector{Conn: sqlDB}
	case DBTypePostgres:
		gormDialector = postgres.New(postgres.Config{Conn: sqlDB})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	gormDB, err := gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create gormDB: %w", err)
	}
	return gormDB, nil
}
