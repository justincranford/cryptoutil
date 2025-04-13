package sqlprovider

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func CreateGormDB(sqlRepository *SqlRepository) (*gorm.DB, error) {
	var gormDialector gorm.Dialector
	switch sqlRepository.dbType {
	case DBTypeSQLite:
		gormDialector = sqlite.Dialector{Conn: sqlRepository.sqlDB}
	case DBTypePostgres:
		gormDialector = postgres.New(postgres.Config{Conn: sqlRepository.sqlDB})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", sqlRepository.dbType)
	}

	gormDB, err := gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create gormDB: %w", err)
	}
	return gormDB, nil
}
