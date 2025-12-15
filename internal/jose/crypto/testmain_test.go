// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"context"
	"os"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("jwkgen_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJWKGenService    *JWKGenService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		var err error

		testJWKGenService, err = NewJWKGenService(testCtx, testTelemetryService, false)
		cryptoutilAppErr.RequireNoError(err, "failed to initialize NewJWKGenService")

		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
