package orm

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JwkGenService, sqlRepository *cryptoutilSqlRepository.SqlRepository, applyMigrations bool) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, jwkGenService, sqlRepository, applyMigrations)
	cryptoutilAppErr.RequireNoError(err, "failed to create new ORM repository")
	return ormRepository
}
