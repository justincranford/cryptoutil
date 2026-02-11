// Copyright (c) 2025 Justin Cranford
//
//

package pool

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

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
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("pool_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService

	happyPathWorkers             = []uint32{1, 3, 10}
	happyPathSize                = []uint32{1, 4, 20}
	happyPathMaxLifetimeValues   = []uint64{1, 50, cryptoutilSharedMagic.MaxPoolLifetimeValues}
	happyPathMaxLifetimeDuration = []time.Duration{cryptoutilSharedMagic.MaxPoolLifetimeDuration}
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
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
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

// TestName verifies the pool name accessor function.
func TestName(t *testing.T) {
	t.Parallel()

	poolName := "TestNamePool"
	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, poolName, 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	require.Equal(t, poolName, poolInstance.Name())
}

// TestGetMany tests the GetMany function for batch retrieval.
func TestGetMany(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		numValues int
		expectLen int
	}{
		{name: "get zero values", numValues: 0, expectLen: 0},
		{name: "get negative values", numValues: -1, expectLen: 0},
		{name: "get one value", numValues: 1, expectLen: 1},
		{name: "get five values", numValues: 5, expectLen: 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, 2, 10, 100, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
			require.NoError(t, err)
			require.NotNil(t, poolInstance)

			defer poolInstance.Cancel()

			values := poolInstance.GetMany(tc.numValues)
			if tc.expectLen == 0 {
				require.Nil(t, values)
			} else {
				require.Len(t, values, tc.expectLen)

				for _, v := range values {
					require.NotNil(t, v)
				}
			}
		})
	}
}

// TestGetManyCanceled tests GetMany when pool is canceled mid-retrieval.
func TestGetManyCanceled(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "GetManyCanceled", 1, 2, 100, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	// Get some values first to ensure pool is working
	values := poolInstance.GetMany(2)
	require.Len(t, values, 2)

	// Cancel the pool
	poolInstance.Cancel()

	// GetMany should return partial/empty results after cancel
	values = poolInstance.GetMany(5)
	require.True(t, len(values) < 5) // Should get fewer values due to cancel
}

// TestCancelNotNil tests the CancelNotNil utility function.
func TestCancelNotNil(t *testing.T) {
	t.Parallel()

	// Test with nil pool - should not panic
	CancelNotNil[any](nil)

	// Test with real pool
	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "CancelNotNil", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	CancelNotNil(poolInstance)

	// Verify pool is canceled - Get should return zero value
	value := poolInstance.Get()
	require.Nil(t, value)
}

// TestCancelAllNotNil tests the CancelAllNotNil utility function.
func TestCancelAllNotNil(t *testing.T) {
	t.Parallel()

	// Create multiple pools with the same type (*uuid.UUID)
	pool1, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "CancelAll1", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)

	pool2, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "CancelAll2", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)

	// Include nil in the slice - note: must use same type *googleUuid.UUID
	pools := []*ValueGenPool[*googleUuid.UUID]{pool1, nil, pool2}

	// Should not panic with nil in slice
	CancelAllNotNil(pools)

	// Verify both pools are canceled
	require.Nil(t, pool1.Get())
	require.Nil(t, pool2.Get())
}

// TestCancelIdempotent tests that Cancel can be called multiple times safely.
func TestCancelIdempotent(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "CancelIdempotent", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	// Cancel multiple times - should not panic
	poolInstance.Cancel()
	poolInstance.Cancel()
	poolInstance.Cancel()

	// Verify pool is canceled
	value := poolInstance.Get()
	require.Nil(t, value)
}

// TestValidateConfig_NilContext tests validation with nil context.
func TestValidateConfig_NilContext(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // SA1012: intentionally passing nil context to test validation
	_, err := NewValueGenPool(NewValueGenPoolConfig(nil, testTelemetryService, "test", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "context can't be nil")
}

// TestValidateConfig_NilTelemetry tests validation with nil telemetry service.
func TestValidateConfig_NilTelemetry(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, nil, "test", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "telemetry service can't be nil")
}

// TestValidateConfig_EmptyPoolName tests validation with empty pool name.
func TestValidateConfig_EmptyPoolName(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "name can't be empty")
}

// TestValidateConfig_ZeroWorkers tests validation with zero workers.
func TestValidateConfig_ZeroWorkers(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 0, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "number of workers can't be 0")
}

// TestValidateConfig_ZeroPoolSize tests validation with zero pool size.
func TestValidateConfig_ZeroPoolSize(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 1, 0, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool size can't be 0")
}

// TestValidateConfig_ZeroMaxValues tests validation with zero max lifetime values.
func TestValidateConfig_ZeroMaxValues(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 1, 1, 0, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "max lifetime values can't be 0")
}

// TestValidateConfig_ZeroMaxDuration tests validation with zero max lifetime duration.
func TestValidateConfig_ZeroMaxDuration(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 1, 1, 10, 0, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "max lifetime duration must be positive")
}

// TestValidateConfig_NegativeMaxDuration tests validation with negative max lifetime duration.
func TestValidateConfig_NegativeMaxDuration(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 1, 1, 10, -time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "max lifetime duration must be positive")
}

// TestValidateConfig_WorkersGreaterThanPoolSize tests validation when workers > pool size.
func TestValidateConfig_WorkersGreaterThanPoolSize(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 5, 2, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "number of workers can't be greater than pool size")
}

// TestValidateConfig_PoolSizeGreaterThanMaxValues tests validation when pool size > max values.
func TestValidateConfig_PoolSizeGreaterThanMaxValues(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 1, 20, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool size can't be greater than max lifetime values")
}

// TestValidateConfig_NilGenerateFunction tests validation with nil generate function.
func TestValidateConfig_NilGenerateFunction(t *testing.T) {
	t.Parallel()

	// Need to provide a typed nil function since Go can't infer T from nil
	var nilGenFunc func() (*googleUuid.UUID, error)

	_, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "test", 1, 1, 10, time.Second, nilGenFunc, false))
	require.Error(t, err)
	require.Contains(t, err.Error(), "generate function can't be nil")
}

// TestNewValueGenPoolConfigError tests error propagation from NewValueGenPoolConfig.
func TestNewValueGenPoolConfigError(t *testing.T) {
	t.Parallel()

	// Create a config that returns an error (nil context)
	//nolint:staticcheck // SA1012: intentionally passing nil context to test error propagation
	cfg, err := NewValueGenPoolConfig(nil, testTelemetryService, "test", 1, 1, 10, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false)
	require.Error(t, err)
	require.Nil(t, cfg)

	// Pass the error to NewValueGenPool
	pool, err := NewValueGenPool(cfg, err)
	require.Error(t, err)
	require.Nil(t, pool)
	require.Contains(t, err.Error(), "failed to create pool config")
}

// TestVerboseMode tests pool operations with verbose logging enabled.
func TestVerboseMode(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "VerbosePool", 2, 5, 20, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), true))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Perform operations to trigger verbose logging
	value := poolInstance.Get()
	require.NotNil(t, value)

	values := poolInstance.GetMany(3)
	require.Len(t, values, 3)

	poolInstance.Cancel()
}

// TestConcurrentGetOperations tests concurrent Get operations.
func TestConcurrentGetOperations(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "ConcurrentGet", 4, 10, 100, time.Second, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	var wg sync.WaitGroup

	var successCount atomic.Int64

	numGoroutines := 10
	getsPerGoroutine := 5

	for range numGoroutines {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range getsPerGoroutine {
				value := poolInstance.Get()
				if value != nil {
					successCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	require.Equal(t, int64(numGoroutines*getsPerGoroutine), successCount.Load())
}

// TestMaxLifetimeValuesLimit tests that pool respects maxLifetimeValues limit.
func TestMaxLifetimeValuesLimit(t *testing.T) {
	t.Parallel()

	maxValues := uint64(5)
	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "MaxValuesLimit", 2, 3, maxValues, time.Second*30, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Get all available values
	var gotCount int

	for range int(maxValues) + 5 { // Try to get more than maxValues
		value := poolInstance.Get()
		if value != nil {
			gotCount++
		}
	}

	// Should get at most maxValues (could get fewer due to timing)
	require.LessOrEqual(t, gotCount, int(maxValues)+int(3)) // +poolSize for buffered values
}

// TestMaxLifetimeDurationLimit tests that pool respects maxLifetimeDuration limit.
// The pool's closeChannelsThread checks limits every PoolMaintenanceInterval (500ms),
// so we need to wait long enough for the check to discover and enforce the time limit.
// After the limit is enforced, existing buffered values can still be consumed, but
// no new values will be generated and eventually Get() returns zero.
func TestMaxLifetimeDurationLimit(t *testing.T) {
	t.Parallel()

	// Use a short duration (100ms) that will expire before the maintenance check
	// Pool size 2 means up to 2 buffered values
	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(testCtx, testTelemetryService, "MaxDurationLimit", 1, 2, 1000, 100*time.Millisecond, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Wait for:
	// 1. The duration limit to expire (100ms)
	// 2. The maintenance check to discover and enforce the limit (up to 500ms)
	// 3. Workers to stop and channels to close
	// Total: ~700ms should be plenty
	time.Sleep(700 * time.Millisecond)

	// After pool is cancelled due to time limit, drain any buffered values
	// and then Get should return zero value
	for range 10 { // More than pool size to ensure we exhaust buffer
		value := poolInstance.Get()
		if value == nil {
			// Got zero value, which is expected after pool is cancelled and buffer is empty
			return // Test passes
		}
	}

	// If we got here, we retrieved 10 values which should be impossible
	// given pool size of 2 and time limit that should have stopped generation
	t.Fatal("Expected pool to return nil after time limit, but kept returning values")
}
