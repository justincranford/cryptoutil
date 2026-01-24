// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"context"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
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
		cryptoutilSharedApperr.RequireNoError(err, "failed to initialize NewJWKGenService")

		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
