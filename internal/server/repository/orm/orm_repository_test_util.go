package orm

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, jwkGenService *cryptoutilJose.JwkGenService, settings *cryptoutilConfig.Settings) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, sqlRepository, jwkGenService, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to create new ORM repository")
	return ormRepository
}
