// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"context"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("jwkgen_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *JWKGenService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		var err error

		testJWKGenService, err = NewJWKGenService(testCtx, testTelemetryService, false)
		cryptoutilSharedApperr.RequireNoError(err, "failed to initialize NewJWKGenService")

		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}
