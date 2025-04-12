package orm

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, sqlProvider *cryptoutilSqlProvider.SqlProvider, applyMigrations bool) *RepositoryProvider {
	repositoryProvider, err := NewRepositoryOrm(ctx, telemetryService, sqlProvider, applyMigrations)
	cryptoutilAppErr.RequireNoError(err, "failed to create new repository provider")
	return repositoryProvider
}
