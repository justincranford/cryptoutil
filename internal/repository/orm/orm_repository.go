package orm

import (
	"context"
	"fmt"

	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilUtil "cryptoutil/internal/util"

	"gorm.io/gorm"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var (
	ormEntities                   = []any{&RootKey{}, &IntermediateKey{}, &ContentKey{}, &KeyPool{}, &Key{}}
	ErrKeyPoolIDMustBeNonZeroUUID = fmt.Errorf("invalid Key Pool ID: %w", cryptoutilUtil.ErrNonZeroUUID)
	ErrKeyIDMustBeNonZeroUUID     = fmt.Errorf("invalid Key ID: %w", cryptoutilUtil.ErrNonZeroUUID)
)

type OrmRepository struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	sqlRepository    *cryptoutilSqlRepository.SqlRepository
	uuidV7KeyGenPool *cryptoutilKeygen.KeyGenPool
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, applyMigrations bool) (*OrmRepository, error) {
	uuidV7KeyGenPoolConfig, err := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Orm UUIDv7", 2, 3, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateUUIDv7Function())
	if err != nil {
		return nil, fmt.Errorf("failed to create UUID V7 pool config: %w", err)
	}
	uuidV7KeyGenPool, err := cryptoutilKeygen.NewGenKeyPool(uuidV7KeyGenPoolConfig)
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

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: sqlRepository, uuidV7KeyGenPool: uuidV7KeyGenPool, gormDB: gormDB, applyMigrations: applyMigrations}, nil
}

func (s *OrmRepository) Shutdown() {
	s.telemetryService.Slogger.Debug("stopping ORM repository")
	s.sqlRepository.Shutdown()
	s.telemetryService.Slogger.Debug("stopped ORM repository")
}
