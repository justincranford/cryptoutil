package orm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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

// var (
// dbNameDefault = "postgres" //kekservice
// 	dbUsername = fmt.Sprintf("postgresUsername%08d", rand.Intn(100_000_000))
// 	dbPassword = fmt.Sprintf("postgresPassword%08d", rand.Intn(100_000_000))
// )

func NewService(ctx context.Context, dbType DBType, dsn string, containerMode ContainerMode, applyMigrations bool) (*Service, error) {
	var gormDialector gorm.Dialector
	gormConfig := gorm.Config{}

	// create gormDialector using provided dsn, or the dsn from a required||preferred container
	gormDialector, shutdownContainer, err := newGormDialector(ctx, dbType, dsn, containerMode)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// use sqlite or postgres GORM dialector to connect to the requested database
	gormDB, err := gorm.Open(gormDialector, &gormConfig)
	if err != nil {
		shutdownContainer()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// extract the underlying SQL database connection from the GORM database connection, in case the caller needs to use it directly
	log.Printf("attempting to get SQL DB connection: %v", err)
	var sqlDB *sql.DB
	for attemptsRemaining := maxDbConnectAttempts; attemptsRemaining > 0; attemptsRemaining-- {
		sqlDB, err = gormDB.DB()
		if err == nil {
			break
		}
		log.Printf("failed to get SQL DB: %v", err)
		if attemptsRemaining > 0 {
			time.Sleep(1 * time.Second) // Wait for DB readiness
		}
	}
	if sqlDB == nil {
		log.Printf("giving up trying to get SQL")
		shutdownContainer()
		return nil, fmt.Errorf("gave up trying to get SQL DB: %w", err)
	}

	// ping the database to verify the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
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
		GormDB:            gormDB,
		SqlDB:             sqlDB,
		shutdownContainer: shutdownContainer,
	}

	return service, nil
}

func newGormDialector(ctx context.Context, dbType DBType, dsn string, containerMode ContainerMode) (gorm.Dialector, func(), error) {
	var shutdownContainer func() = func() {} // no-op by default

	if containerMode != ContainerModeDisabled {
		log.Printf("Container mode is %s, trying to start a %s container and get its DSN", string(dbType), string(containerMode))
		var containerDSN string
		var err error
		switch dbType {
		case DBTypeSQLite:
			return nil, nil, fmt.Errorf("sqlite is supported, but there is no container option for it") // TODO Support RQlite as a container option in future? It wraps SQLite with TCP support.
		case DBTypePostgres:
			containerDSN, shutdownContainer, err = startPostgresContainer(ctx, "postgres", "postgres", "postgres")
		default:
			return nil, nil, fmt.Errorf("unsupported database type: %s", dbType)
		}
		if err != nil {
			// container failed to start (e.g. Docker not installed, Docker Desktop not running, etc.)
			if containerMode == ContainerModeRequired {
				// caller requires container to be started, so return an error
				return nil, nil, fmt.Errorf("containerMode is required, and failed to start %s container: %w", string(dbType), err)
			}
			// caller prefers container to be started, but is OK to fallback on using the provider DSN instead of a DB container's DSN
			log.Printf("containerMode is preferred but container startup failed, changing containerMode to disabled and falling back to provided %s DSN: %v", string(dbType), err)
			containerMode = ContainerModeDisabled
		} else {
			log.Printf("container started successfully, using generated %s DSN: %v", string(dbType), containerDSN)
			dsn = containerDSN
		}
	}

	// execute if containerMode is disabled, or containerMode was preferred but changed to disabled because container startup failed
	var gormDialector gorm.Dialector
	if containerMode == ContainerModeDisabled {
		log.Printf("Container mode is disabled, using provided %s DSN: %s", string(dbType), dsn)
		switch dbType {
		case DBTypeSQLite:
			db, err := sql.Open("sqlite", dsn)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to open SQLite database: %w", err)
			}
			gormDialector = sqlite.Dialector{Conn: db}
		case DBTypePostgres:
			log.Printf("Initializing %s database", string(dbType))
			gormDialector = postgres.Open(dsn)
		default:
			return nil, nil, fmt.Errorf("unsupported database type: %s", dbType)
		}
	}
	return gormDialector, shutdownContainer, nil
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
