// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"context"
	"os"
	"testing"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	testTelemetrySettings = cryptoutilSharedTelemetry.NewTestTelemetrySettings("jwkgen_service_test")
	testCtx               = context.Background()
	testTelemetryService  *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService     *JWKGenService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testTelemetrySettings)
		defer testTelemetryService.Shutdown()

		var err error

		testJWKGenService, err = NewJWKGenService(testCtx, testTelemetryService, false)
		cryptoutilSharedApperr.RequireNoError(err, "failed to initialize NewJWKGenService")

		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
