package telemetry

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilConfig "cryptoutil/internal/common/config"
)

func RequireNewForTest(ctx context.Context, settings *cryptoutilConfig.Settings) *TelemetryService {
	telemetryService, err := NewTelemetryService(ctx, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize telemetry")
	return telemetryService
}
