package sqlprovider

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CreateGormDB(dbType SupportedSqlDB, sqlDB *sql.DB) (*gorm.DB, error) {
	var gormDialector gorm.Dialector
	switch dbType {
	case SupportedSqlDBSQLite:
		gormDialector = sqlite.Dialector{Conn: sqlDB}
	case SupportedSqlDBPostgres:
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
