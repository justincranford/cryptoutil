// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"context"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// RequireNewForTest creates a JWKGenService for testing with panic on error.
func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) *JWKGenService {
	jwkGenService, err := NewJWKGenService(ctx, telemetryService, false)
	cryptoutilSharedApperr.RequireNoError(err, "failed to initialize jwkGenService")

	return jwkGenService
}
