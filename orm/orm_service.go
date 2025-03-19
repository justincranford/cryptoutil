package orm

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite"
)

type Service struct {
	GormDB           *gorm.DB
	SqlDB            *sql.DB
	shutdownFunction func()
}

// NewService creates a new ORM service based on provided settings
func NewService(ctx context.Context, devMode bool) (*Service, error) {
	var gormDialector gorm.Dialector
	gormConfig := gorm.Config{}
	var shutdownFunction func()
	if devMode {
		log.Println("Initializing SQLite in-memory database")
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite in-memory database: %w", err)
		}
		gormDialector = sqlite.Dialector{Conn: db}
		gormConfig = gorm.Config{}
		shutdownFunction = func() {} // no-op
	} else {
		log.Println("Initializing PostgreSQL container database")
		dsn, shutdownFunction, err := startPostgresContainer(ctx, "kekservice", "postgres", "postgres")
		if err != nil {
			shutdownFunction()
			return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
		}
		gormDialector = postgres.Open(dsn)
		gormConfig = gorm.Config{}
	}

	gormDB, err := gorm.Open(gormDialector, &gormConfig)
	if err != nil {
		shutdownFunction()
		return nil, fmt.Errorf("failed to connect to PostgreSQL using provided DSN: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		shutdownFunction()
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	err = gormDB.AutoMigrate(ormTableStructs...)
	if err != nil {
		shutdownFunction()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	service := &Service{
		GormDB:           gormDB,
		SqlDB:            sqlDB,
		shutdownFunction: shutdownFunction,
	}

	return service, nil
}

func (s *Service) Shutdown() {
	err := s.SqlDB.Close()
	if err == nil {
		log.Printf("failed to close DB: %v", err)
	}
	if s.shutdownFunction != nil {
		s.shutdownFunction()
	}
}
