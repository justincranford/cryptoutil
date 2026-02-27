// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// RequireNewForTest creates a TelemetryService for testing and panics on initialization errors.
func RequireNewForTest(ctx context.Context, settings *TelemetrySettings) *TelemetryService {
	telemetryService, err := NewTelemetryService(ctx, settings)
	cryptoutilSharedApperr.RequireNoError(err, "failed to initialize telemetry")

	return telemetryService
}
