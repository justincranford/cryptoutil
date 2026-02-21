// Copyright (c) 2025 Justin Cranford
//
//

package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestNewValueGenPool_WithError covers the passed-in error path (line 61).
func TestNewValueGenPool_WithError(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool[googleUuid.UUID](nil, errForTest("injected config error"))
	require.Error(t, err)
	require.ErrorContains(t, err, "injected config error")
}

// TestNewValueGenPool_NilGenerateFunction covers the validateConfig error path (lines 69, 423).
func TestNewValueGenPool_NilGenerateFunction(t *testing.T) {
	t.Parallel()

	// Directly instantiate config (package-level test, private fields accessible).
	cfg := &ValueGenPoolConfig[googleUuid.UUID]{
		ctx:                 context.Background(),
		telemetryService:    testTelemetryService,
		poolName:            "nil-fn-test",
		numWorkers:          1,
		poolSize:            1,
		maxLifetimeValues:   1,
		maxLifetimeDuration: time.Second,
		generateFunction:    nil, // Triggers validateConfig nil generateFunction check.
	}

	_, err := NewValueGenPool(cfg, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "generate function can't be nil")
}

// TestNewValueGenPool_NilConfig covers validateConfig nil-config return (line 425).
func TestNewValueGenPool_NilConfig(t *testing.T) {
	t.Parallel()

	_, err := NewValueGenPool[googleUuid.UUID](nil, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "config can't be nil")
}

// TestNewValueGenPool_LargeMaxLifetimeValues covers the large-values safety cap else branch (line 71).
func TestNewValueGenPool_LargeMaxLifetimeValues(t *testing.T) {
	t.Parallel()

	// ^uint64(0) exceeds MaxPoolLifetimeValues (max int64), triggering the int64 safety cap.
	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"large-values", 1, 1, ^uint64(0), time.Minute,
		cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()
}

// TestGet_CanceledBeforeValue covers Get() cancel path (line 173) and
// the worker cancel-before-next-generate path (line 263, triggered during cleanup).
func TestGet_CanceledBeforeValue(t *testing.T) {
	t.Parallel()

	blocks := make(chan struct{})
	workerStarted := make(chan struct{})

	var startOnce sync.Once

	generateFn := func() (googleUuid.UUID, error) {
		startOnce.Do(func() { close(workerStarted) })
		<-blocks // Block until released.

		return googleUuid.Nil, nil
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"get-canceled", 1, 1, 10, time.Minute, generateFn, false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer func() { close(blocks) }() // Unblock worker on cleanup.
	defer poolInstance.Cancel()      // Idempotent cancel on cleanup.

	<-workerStarted           // Wait for worker to enter generateFunction.
	poolInstance.Cancel()     // Cancel pool before any value is published.
	val := poolInstance.Get() // Should hit Done() path at line 173.
	require.Equal(t, googleUuid.UUID{}, val)
}

// TestGetMany_CanceledBeforeValue covers GetMany() zero-value break path (line 218).
func TestGetMany_CanceledBeforeValue(t *testing.T) {
	t.Parallel()

	blocks := make(chan struct{})
	workerStarted := make(chan struct{})

	var startOnce sync.Once

	generateFn := func() (googleUuid.UUID, error) {
		startOnce.Do(func() { close(workerStarted) })
		<-blocks

		return googleUuid.Nil, nil
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"getmany-canceled", 1, 1, 10, time.Minute, generateFn, false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer func() { close(blocks) }()
	defer poolInstance.Cancel()

	<-workerStarted
	poolInstance.Cancel()

	vals := poolInstance.GetMany(3) // Should hit break path at line 218.
	require.Empty(t, vals)
}

// TestWorker_ValueLimitReached covers the generate-value-limit error paths
// in generatePublishRelease (line 336) and generateWorker (line 306).
func TestWorker_ValueLimitReached(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"value-limit", 1, 1, 1, time.Minute,
		cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Consume the single allowed value; worker will attempt a second generation
	// which hits generateCounter > maxLifetimeValues (line 336), then logs the
	// worker error (line 306) and exits.
	val := poolInstance.Get()
	require.NotEqual(t, googleUuid.UUID{}, val)

	// Cancel and drain channel — channel closes only after all worker goroutines
	// finish, ensuring the error-path code is captured in the coverage profile.
	poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestWorker_TimeLimitReached covers the pool-time-limit error paths
// in generatePublishRelease (line 326) and generateWorker (line 306).
func TestWorker_TimeLimitReached(t *testing.T) {
	t.Parallel()

	// Use a generate function that takes 2ms so the pool's 1ms lifetime expires.
	generateFn := func() (googleUuid.UUID, error) {
		time.Sleep(2 * time.Millisecond)

		return googleUuid.NewV7()
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"time-limit", 1, 1, 100, 1*time.Millisecond, generateFn, false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Cancel and drain channel — channel closes only after worker goroutines finish,
	// ensuring error-path code is captured in coverage.
	poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestCloseChannelsThread_TickerPath covers the ticker-based close path in
// closeChannelsThread (line 375). NOT parallel — modifies package-level var.
func TestCloseChannelsThread_TickerPath(t *testing.T) {
	origInterval := poolMaintenanceInterval
	poolMaintenanceInterval = 1 * time.Millisecond

	defer func() { poolMaintenanceInterval = origInterval }()

	// Use a generate function that sleeps longer than the pool lifetime so
	// the time limit is guaranteed to be detected by the ticker.
	generateFn := func() (googleUuid.UUID, error) {
		time.Sleep(2 * time.Millisecond)

		return googleUuid.NewV7()
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"ticker-path", 1, 1, 100, 1*time.Millisecond, generateFn, false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Drain channel — closes when ticker triggers time-limit close path.
	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestCloseChannelsThread_InfinitePath covers the infinite-pool path in closeChannelsThread (line 377).
// With both maxLifetimeValues=0 and maxLifetimeDuration=0, the pool is infinite and blocks on Done().
func TestCloseChannelsThread_InfinitePath(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"infinite-pool", 1, 1, 0, 0,
		cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	// Cancel triggers the Done() case inside closeChannelsThread's infinite-pool branch.
	poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestGet_CanceledVerbose covers the verbose debug body in Get()'s Done case (line 175).
func TestGet_CanceledVerbose(t *testing.T) {
	t.Parallel()

	blocks := make(chan struct{})
	workerStarted := make(chan struct{})

	var startOnce sync.Once

	generateFn := func() (googleUuid.UUID, error) {
		startOnce.Do(func() { close(workerStarted) })
		<-blocks

		return googleUuid.Nil, nil
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"get-canceled-verbose", 1, 1, 10, time.Minute, generateFn, true,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer func() { close(blocks) }()
	defer poolInstance.Cancel()

	<-workerStarted
	poolInstance.Cancel()

	val := poolInstance.Get()
	require.Equal(t, googleUuid.UUID{}, val)
}

// TestGetMany_CanceledVerbose covers the verbose body in GetMany()'s zero-value break (line 220).
func TestGetMany_CanceledVerbose(t *testing.T) {
	t.Parallel()

	blocks := make(chan struct{})
	workerStarted := make(chan struct{})

	var startOnce sync.Once

	generateFn := func() (googleUuid.UUID, error) {
		startOnce.Do(func() { close(workerStarted) })
		<-blocks

		return googleUuid.Nil, nil
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"getmany-canceled-verbose", 1, 1, 10, time.Minute, generateFn, true,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer func() { close(blocks) }()
	defer poolInstance.Cancel()

	<-workerStarted
	poolInstance.Cancel()

	vals := poolInstance.GetMany(3)
	require.Empty(t, vals)
}

// TestWorker_ValueLimitVerbose covers verbose debug bodies for value limit (lines 265, 308, 338).
// It drains the generateChannel to wait for worker goroutines to finish their defers.
func TestWorker_ValueLimitVerbose(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"value-limit-verbose", 1, 1, 1, time.Minute,
		cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), true,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	val := poolInstance.Get()
	require.NotEqual(t, googleUuid.UUID{}, val)

	// Drain channel until closed; closing happens only after all worker defers complete,
	// ensuring the verbose debug bodies in goroutine defers are captured in coverage.
	poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestWorker_TimeLimitVerbose covers verbose debug bodies for time limit (lines 265, 308, 328).
// It drains the generateChannel to wait for worker goroutines to finish their defers.
func TestWorker_TimeLimitVerbose(t *testing.T) {
	t.Parallel()

	generateFn := func() (googleUuid.UUID, error) {
		time.Sleep(2 * time.Millisecond)

		return googleUuid.NewV7()
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"time-limit-verbose", 1, 1, 100, 1*time.Millisecond, generateFn, true,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	// Drain channel until closed; ensures worker defers complete before test exits.
	poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// errForTest returns a simple error for test injection.
func errForTest(msg string) error {
	return &testError{msg: msg}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
