// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
)

func RequireNewForTest(ctx context.Context, settings *cryptoutilConfig.ServerSettings) *TelemetryService {
	telemetryService, err := NewTelemetryService(ctx, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize telemetry")

	return telemetryService
}
