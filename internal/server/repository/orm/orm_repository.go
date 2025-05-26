package orm

import (
	"context"
	"fmt"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	"cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	"gorm.io/gorm"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var ormEntities = []any{&BarrierRootKey{}, &BarrierIntermediateKey{}, &BarrierContentKey{}, &KeyPool{}, &Key{}}

type OrmRepository struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	sqlRepository    *cryptoutilSqlRepository.SqlRepository
	uuidV7KeyGenPool *pool.ValueGenPool[cryptoutilKeygen.Key]
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, applyMigrations bool) (*OrmRepository, error) {
	uuidV7KeyGenPoolConfig, err := pool.NewValueGenPoolConfig(ctx, telemetryService, "Orm UUIDv7", 2, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, cryptoutilKeygen.GenerateUUIDv7Function())
	if err != nil {
		return nil, fmt.Errorf("failed to create UUID V7 pool config: %w", err)
	}
	uuidV7KeyGenPool, err := pool.NewValueGenPool(uuidV7KeyGenPoolConfig)
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
	s.telemetryService.Slogger.Debug("stopping ORM repository")
	s.sqlRepository.Shutdown()
	s.telemetryService.Slogger.Debug("stopped ORM repository")
}
