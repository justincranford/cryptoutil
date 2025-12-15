// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilConfig "cryptoutil/internal/shared/config"
)

func RequireNewForTest(ctx context.Context, settings *cryptoutilConfig.Settings) *TelemetryService {
	telemetryService, err := NewTelemetryService(ctx, settings)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize telemetry")

	return telemetryService
}
