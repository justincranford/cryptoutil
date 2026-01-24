// Copyright (c) 2025 Justin Cranford
//
//

package pool

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name                string
	workers             uint32
	size                uint32
	maxLifetimeValues   uint64
	maxLifetimeDuration time.Duration
	gets                uint64
}

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("pool_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService

	happyPathWorkers             = []uint32{1, 3, 10}
	happyPathSize                = []uint32{1, 4, 20}
	happyPathMaxLifetimeValues   = []uint64{1, 50, cryptoutilMagic.MaxPoolLifetimeValues}
	happyPathMaxLifetimeDuration = []time.Duration{cryptoutilMagic.MaxPoolLifetimeDuration}
	happyPathGets                = []uint64{0, 1, 4, 50}

	happyPathTestCases = func() []*TestCase {
		testCases := make([]*TestCase, 0, len(happyPathWorkers)*len(happyPathSize)*len(happyPathMaxLifetimeValues)*len(happyPathMaxLifetimeDuration)*len(happyPathGets))
		for _, workers := range happyPathWorkers {
			for _, size := range happyPathSize {
				if workers > size {
					continue
				}
				for _, maxLifetimeValues := range happyPathMaxLifetimeValues {
					if uint64(size) > maxLifetimeValues {
						continue
					}
					for _, maxLifetimeDuration := range happyPathMaxLifetimeDuration {
						for _, gets := range happyPathGets {
							if gets > maxLifetimeValues {
								continue
							}
							name := fmt.Sprintf("workers[%d] size[%d] maxLifetimeValues[%d] maxLifetimeDuration[%v] gets[%d]", workers, size, maxLifetimeValues, maxLifetimeDuration, gets)
							testCases = append(testCases, &TestCase{name: name, workers: workers, size: size, maxLifetimeValues: maxLifetimeValues, maxLifetimeDuration: maxLifetimeDuration, gets: gets})
						}
					}
				}
			}
		}

		return testCases
	}()
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown() // this needs to run before os.Exit

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestHappyPath(t *testing.T) {
	t.Parallel()

	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeValues, tc.maxLifetimeDuration, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
			require.NoError(t, err)
			require.NotNil(t, poolInstance)

			defer poolInstance.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				generated := poolInstance.Get()
				require.NotNil(t, generated)
			}
		})
	}
}

func TestGenerateError(t *testing.T) {
	const numGets = 3

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "Fail", 1, 1, numGets, time.Second, generateErrorFunction(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	for i := uint64(0); i < numGets; i++ {
		generated := poolInstance.Get()
		require.Nil(t, generated)
	}
}

func generateErrorFunction() func() (any, error) {
	return func() (any, error) { return nil, fmt.Errorf("generate error") }
}
