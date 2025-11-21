// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) *JWKGenService {
	jwkGenService, err := NewJWKGenService(ctx, telemetryService, false)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize jwkGenService")

	return jwkGenService
}
