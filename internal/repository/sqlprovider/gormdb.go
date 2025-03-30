package sqlprovider

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CreateGormDB(sqlProvider *SqlProvider) (*gorm.DB, error) {
	var gormDialector gorm.Dialector
	switch sqlProvider.dbType {
	case SupportedSqlDBSQLite:
		gormDialector = sqlite.Dialector{Conn: sqlProvider.sqlDB}
	case SupportedSqlDBPostgres:
		gormDialector = postgres.New(postgres.Config{Conn: sqlProvider.sqlDB})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", sqlProvider.dbType)
	}

	gormDB, err := gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create gormDB: %w", err)
	}
	return gormDB, nil
}
