// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) *SQLRepository {
	sqlRepository, err := NewSQLRepository(ctx, telemetryService, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize SQL provider")

	return sqlRepository
}
