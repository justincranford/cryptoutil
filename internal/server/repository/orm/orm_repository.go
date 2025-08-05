package orm

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/common/config"
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
	verboseMode      bool
	gormDB           *gorm.DB
}

func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, jwkGenService *cryptoutilJose.JwkGenService, settings *cryptoutilConfig.Settings) (*OrmRepository, error) {
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

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: sqlRepository, jwkGenService: jwkGenService, gormDB: gormDB, verboseMode: settings.VerboseMode}, nil
}

func (s *OrmRepository) Shutdown() {
	// no-op
}
