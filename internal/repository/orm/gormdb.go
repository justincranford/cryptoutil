package orm

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CreateGormDB(dbType DBType, sqlDB *sql.DB) (*gorm.DB, error) {
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
