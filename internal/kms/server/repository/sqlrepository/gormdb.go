// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	gormLoggerSlowThreshold             = cryptoutilMagic.DBLoggerSlowThreshold
	gormLoggerLogLevel                  = logger.Warn // Changed from logger.Info to reduce verbosity.
	gormLoggerIgnoreRecordNotFoundError = false
	gormLoggerColorful                  = true
)

func CreateGormDB(sqlRepository *SQLRepository) (*gorm.DB, error) {
	var gormDialector gorm.Dialector

	switch sqlRepository.dbType {
	case DBTypeSQLite:
		gormDialector = sqlite.Dialector{Conn: sqlRepository.sqlDB}
	case DBTypePostgres:
		postgresConfig := postgres.Config{
			Conn: sqlRepository.sqlDB,
		}
		gormDialector = postgres.New(postgresConfig)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", sqlRepository.dbType)
	}

	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             gormLoggerSlowThreshold,
		LogLevel:                  gormLoggerLogLevel,
		IgnoreRecordNotFoundError: gormLoggerIgnoreRecordNotFoundError,
		Colorful:                  gormLoggerColorful,
		ParameterizedQueries:      true,
	})
	gormConfig := gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
	}

	gormDB, err := gorm.Open(gormDialector, &gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gormDB: %w", err)
	}

	// Enable gorm debug mode if verbose mode is enabled
	if sqlRepository.verboseMode {
		gormDB = gormDB.Debug()
	}

	return gormDB, nil
}
