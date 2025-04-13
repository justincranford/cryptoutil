package telemetry

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/apperr"
)

func RequireNewForTest(ctx context.Context, scope string, enableOtel, enableStdout bool) *TelemetryService {
	telemetryService, err := NewTelemetryService(ctx, scope, enableOtel, enableStdout)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize telemetry")
	return telemetryService
}
