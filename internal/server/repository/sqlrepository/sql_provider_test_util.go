package sqlrepository

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) *SQLRepository {
	sqlRepository, err := NewSQLRepository(ctx, telemetryService, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize SQL provider")
	return sqlRepository
}
