package pool

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"

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
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService

	happyPathWorkers             = []uint32{1, 3, 10}
	happyPathSize                = []uint32{1, 4, 20}
	happyPathMaxLifetimeValues   = []uint64{1, 50, MaxLifetimeValues}
	happyPathMaxLifetimeDuration = []time.Duration{MaxLifetimeDuration}
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
							name := fmt.Sprintf("workers[%d] size[%d] maxLifetimeValues[%d] maxLifetimeDuration[%v] gets[%d]", workers, size, maxLifetimeValues, time.Duration(maxLifetimeDuration), gets)
							testCases = append(testCases, &TestCase{name: name, workers: workers, size: size, maxLifetimeValues: maxLifetimeValues, maxLifetimeDuration: time.Duration(maxLifetimeDuration), gets: gets})
						}
					}
				}
			}
		}
		return testCases
	}()
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "pool_test", false, true)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestHappyPath(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeValues, tc.maxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function()))
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
	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "Fail", 1, 1, numGets, time.Second, generateErrorFunction()))
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
