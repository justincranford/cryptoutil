package orm

import (
	"context"
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
	GormDB            *gorm.DB
	SqlDB             *sql.DB
	shutdownContainer func()
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
	dbNameDefault = fmt.Sprintf("kekservice%04d", rand.Intn(10_000))
	dbUsername    = fmt.Sprintf("postgresUsername%04d", rand.Intn(10_000))
	dbPassword    = fmt.Sprintf("postgresPassword%04d", rand.Intn(10_000))
)

func NewService(ctx context.Context, dbType DBType, databaseUrl string, containerMode ContainerMode, applyMigrations bool) (*Service, error) {
	// create gormDialector using provided dsn, or the dsn from a required||preferred container
	sqlDB, shutdownContainer, err := createSqlDB(ctx, dbType, databaseUrl, containerMode)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	gormDB, err := createGormDB(dbType, sqlDB)
	if err != nil {
		shutdownContainer()
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		log.Printf("Applying %s migrations", string(dbType))
		err = gormDB.AutoMigrate(ormTableStructs...)
		if err != nil {
			shutdownContainer()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		// If connection of an external DB that was manually started up (e.g. production, long lived dev test container), it may already be initialized, so don't try recreating tables
		log.Printf("Skipping %s migrations", string(dbType))
	}

	service := &Service{
		SqlDB:             sqlDB,
		GormDB:            gormDB,
		shutdownContainer: shutdownContainer,
	}

	return service, nil
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

func createSqlDB(ctx context.Context, dbType DBType, databaseUrl string, containerMode ContainerMode) (*sql.DB, func(), error) {
	var shutdownContainer func() = func() {} // no-op by default

	if containerMode != ContainerModeDisabled {
		log.Printf("Container mode is %s, trying to start a %s container", string(dbType), string(containerMode)) // containerMode is required or preferred
		var containerDatabaseUrl string
		var err error
		switch dbType {
		case DBTypeSQLite:
			return nil, nil, fmt.Errorf("there is no container option for sqlite")
		case DBTypePostgres:
			containerDatabaseUrl, shutdownContainer, err = startPostgresContainer(ctx, dbNameDefault, dbUsername, dbPassword)
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
		shutdownContainer()
		return nil, nil, fmt.Errorf("gave up trying to get SQL DB: %w", err)
	}

	return sqlDB, shutdownContainer, nil
}

func (s *Service) Shutdown() {
	// shutdown the underlying DB connection
	err := s.SqlDB.Close()
	if err == nil {
		log.Printf("failed to close DB: %v", err)
	}
	// if DB was started by this applciation is a container, gracefully clean up the container
	if s.shutdownContainer != nil {
		s.shutdownContainer()
	}
}
