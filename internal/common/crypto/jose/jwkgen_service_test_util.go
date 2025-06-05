package jose

import (
	"context"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) *JwkGenService {
	jwkGenService, err := NewJwkGenService(ctx, telemetryService)
	cryptoutilAppErr.RequireNoError(err, "failed to initialize jwkGenService")
	return jwkGenService
}
