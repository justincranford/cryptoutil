package certificate

import (
	"context"
	"crypto/elliptic"
	"log"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

const (
	numWorkers = 4
	poolSize   = 20
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("certificates_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testKeyGenPool       *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		var err error

		testKeyGenPool, err = cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(testCtx, testTelemetryService, "certificates_test", numWorkers, poolSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
		if err != nil {
			log.Fatalf("failed to create key pool: %v", err)
		}

		defer testKeyGenPool.Cancel()

		rc = m.Run()
	}()
	os.Exit(rc)
}
