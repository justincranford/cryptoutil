package telemetry

import (
	"context"
	cryptoutilAppErr "cryptoutil/internal/apperr"
)

func RequireNewService(ctx context.Context, scope string, enableOtel, enableStdout bool) *Service {
	telemetryService, err := NewService(ctx, scope, enableOtel, enableStdout)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize telemetry")
	return telemetryService
}
