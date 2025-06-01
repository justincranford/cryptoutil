package orm

import (
	"context"
	"fmt"

	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	"gorm.io/gorm"

	googleUuid "github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var ormEntities = []any{&BarrierRootKey{}, &BarrierIntermediateKey{}, &BarrierContentKey{}, &KeyPool{}, &Key{}}

type OrmRepository struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	sqlRepository    *cryptoutilSqlRepository.SqlRepository
	uuidV7KeyGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, applyMigrations bool) (*OrmRepository, error) {
	uuidV7KeyGenPool, err := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Orm UUIDv7", 2, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function()))
	if err != nil {
		return nil, fmt.Errorf("failed to create UUID V7 pool: %w", err)
	}

	gormDB, err := cryptoutilSqlRepository.CreateGormDB(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		telemetryService.Slogger.Debug("applying migrations")
		err = gormDB.AutoMigrate(ormEntities...)
		if err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		telemetryService.Slogger.Debug("skipping migrations")
	}
	err = cryptoutilSqlRepository.LogSchema(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to log schemas: %w", err)
	}

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: sqlRepository, uuidV7KeyGenPool: uuidV7KeyGenPool, gormDB: gormDB, applyMigrations: applyMigrations}, nil
}

func (s *OrmRepository) Shutdown() {
	// no-op
}
