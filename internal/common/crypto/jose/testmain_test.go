package jose

import (
	"context"
	"os"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
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
		testJWKGenService, err = NewJWKGenService(testCtx, testTelemetryService)
		cryptoutilAppErr.RequireNoError(err, "failed to initialize NewJWKGenService")
		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
