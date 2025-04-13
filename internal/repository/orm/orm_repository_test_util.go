package orm

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, applyMigrations bool) *OrmRepository {
	ormRepository, err := NewOrmRepository(ctx, telemetryService, sqlRepository, applyMigrations)
	cryptoutilAppErr.RequireNoError(err, "failed to create new ORM repository")
	return ormRepository
}
