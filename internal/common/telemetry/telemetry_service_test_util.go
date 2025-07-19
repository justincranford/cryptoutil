package telemetry

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
)

func RequireNewForTest(ctx context.Context, enableOtel bool, enableStdout bool, scope string) *TelemetryService {
	telemetryService, err := NewTelemetryService(ctx, enableOtel, enableStdout, scope)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize telemetry")
	return telemetryService
}
