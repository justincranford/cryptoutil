// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// RequireNewForTest creates a new SQLRepository for testing, panicking on error.
func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.ServiceTemplateServerSettings) *SQLRepository {
	sqlRepository, err := NewSQLRepository(ctx, telemetryService, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize SQL provider")

	return sqlRepository
}
