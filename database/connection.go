package database

import (
	"database/sql"
	"fmt"
	"log"

	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite"
)

const (
	driverNameSqlite  = "sqlite"
	databaseUrlSqlite = ":memory:"
)

type Service struct {
	db     *sql.DB
	gormDB *gorm.DB
}

func NewService() (*Service, error) {
	return openDatabase(driverNameSqlite, databaseUrlSqlite)
}

func openDatabase(driverName string, databaseUrl string) (*Service, error) {
	if len(driverName) < 1 {
		return nil, fmt.Errorf("database driver must be non-blank")
	} else if len(databaseUrl) < 1 {
		return nil, fmt.Errorf("database name must be non-blank")
	}

	log.Printf("opening database")
	db, err := sql.Open(driverName, databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	log.Printf("pinging database")
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("opening ORM")
	gormDB, err := gorm.Open(gormsqlite.Dialector{Conn: db}, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open ORM: %w", err)
	}

	return &Service{db: db, gormDB: gormDB}, nil
}

func (s *Service) DB() *sql.DB {
	return s.db
}

func (s *Service) GormDB() *gorm.DB {
	return s.gormDB
}

func (s *Service) Shutdown() {
	if s.db == nil {
		log.Printf("database is already closed")
		return
	}
	log.Printf("closing database")
	if err := s.db.Close(); err != nil {
		log.Printf("close database error: %v", err)
	} else {
		log.Printf("closed database")
	}
	s.db = nil
}
