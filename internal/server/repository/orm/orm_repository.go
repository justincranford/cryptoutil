package orm

import (
	"context"
	"fmt"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	"gorm.io/gorm"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var ormEntities = []any{&BarrierRootKey{}, &BarrierIntermediateKey{}, &BarrierContentKey{}, &ElasticKey{}, &MaterialKey{}}

type OrmRepository struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	sqlRepository    *cryptoutilSqlRepository.SqlRepository
	jwkGenService    *cryptoutilJose.JwkGenService
	gormDB           *gorm.DB
	applyMigrations  bool
}

func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, jwkGenService *cryptoutilJose.JwkGenService, applyMigrations bool) (*OrmRepository, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if sqlRepository == nil {
		return nil, fmt.Errorf("sqlRepository must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	gormDB, err := cryptoutilSqlRepository.CreateGormDB(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	if applyMigrations {
		telemetryService.Slogger.Debug("applying migrations")

		// Get the raw SQL DB from GORM
		sqlDB, err := gormDB.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get SQL DB from GORM: %w", err)
		}

		// Apply SQL migrations using the embedded migration files
		err = cryptoutilSqlRepository.ApplyEmbeddedSqlMigrations(telemetryService, sqlDB)
		if err != nil {
			return nil, fmt.Errorf("failed to apply SQL migrations: %w", err)
		}

		telemetryService.Slogger.Debug("migrations completed successfully")
	} else {
		telemetryService.Slogger.Debug("skipping migrations")
	}
	err = cryptoutilSqlRepository.LogSchema(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to log schemas: %w", err)
	}

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: sqlRepository, jwkGenService: jwkGenService, gormDB: gormDB, applyMigrations: applyMigrations}, nil
}

func (s *OrmRepository) Shutdown() {
	// no-op
}
