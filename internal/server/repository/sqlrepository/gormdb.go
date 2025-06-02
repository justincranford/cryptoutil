package sqlrepository

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

const (
	gormLoggerSlowThreshold             = 200 * time.Millisecond
	gormLoggerLogLevel                  = logger.Info
	gormLoggerIgnoreRecordNotFoundError = false
	gormLoggerColorful                  = true
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

	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             gormLoggerSlowThreshold,
		LogLevel:                  gormLoggerLogLevel,
		IgnoreRecordNotFoundError: gormLoggerIgnoreRecordNotFoundError,
		Colorful:                  gormLoggerColorful,
	})
	gormDB, err := gorm.Open(gormDialector, &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to create gormDB: %w", err)
	}
	return gormDB, nil
}
