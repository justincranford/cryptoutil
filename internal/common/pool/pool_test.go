package pool

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	googleUuid "github.com/google/uuid"
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
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "pool_test", false, false)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestPoolUUIDv7(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			poolConfig, err := NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeValues, tc.maxLifetimeDuration, GenerateUUIDv7Function())
			require.NoError(t, err)
			require.NotNil(t, poolConfig)
			poolInstance, err := NewValueGenPool(poolConfig)
			require.NoError(t, err)
			require.NotNil(t, poolInstance)
			defer poolInstance.Close()

			for i := uint64(0); i < tc.gets; i++ {
				generated := poolInstance.Get()
				require.NotNil(t, generated)
			}
		})
	}
}

func GenerateUUIDv7Function() func() (*googleUuid.UUID, error) {
	return func() (*googleUuid.UUID, error) { return GenerateUUIDv7() }
}

func GenerateUUIDv7() (*googleUuid.UUID, error) {
	uuidV7, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	return &uuidV7, nil
}
