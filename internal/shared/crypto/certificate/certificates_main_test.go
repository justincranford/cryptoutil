// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"context"
	"crypto/elliptic"
	"log"
	"os"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

const (
	numWorkers = 4
	poolSize   = 20
)

var (
	testTelemetrySettings = cryptoutilSharedTelemetry.NewTestTelemetrySettings("certificates_test")
	testCtx               = context.Background()
	testTelemetryService  *cryptoutilSharedTelemetry.TelemetryService
	testKeyGenPool        *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testTelemetrySettings)
		defer testTelemetryService.Shutdown()

		var err error

		testKeyGenPool, err = cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, "certificates_test", numWorkers, poolSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
		if err != nil {
			log.Fatalf("failed to create key pool: %v", err)
		}

		defer testKeyGenPool.Cancel()

		rc = m.Run()
	}()
	os.Exit(rc)
}
