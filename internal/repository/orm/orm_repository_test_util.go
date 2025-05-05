package orm

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, applyMigrations bool) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, sqlRepository, applyMigrations)
	cryptoutilAppErr.RequireNoError(err, "failed to create new ORM repository")
	return ormRepository
}
