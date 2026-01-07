// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.ServiceTemplateServerSettings) *SQLRepository {
	sqlRepository, err := NewSQLRepository(ctx, telemetryService, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize SQL provider")

	return sqlRepository
}
