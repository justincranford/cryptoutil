// Copyright (c) 2025-2026 Justin Cranford.
//
//

package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type cancelOnDoneCallContext struct {
	context.Context
	done             chan struct{}
	cancelOnDoneCall uint32
	doneCalls        atomic.Uint32
}

func (ctx *cancelOnDoneCallContext) Done() <-chan struct{} {
	if ctx.doneCalls.Add(1) == ctx.cancelOnDoneCall {
		close(ctx.done)
	}

	return ctx.done
}

func (ctx *cancelOnDoneCallContext) Err() error {
	select {
	case <-ctx.done:
		return context.Canceled
	default:
		return nil
	}
}

// TestNewValueGenPool_Errors covers constructor error paths and the large-values safety cap.
func TestNewValueGenPool_Errors(t *testing.T) {
	t.Parallel()

	uuidGenFn := func() (googleUuid.UUID, error) { return googleUuid.NewV7() }

	tests := []struct {
		name    string
		setupFn func() (*ValueGenPool[googleUuid.UUID], error)
		wantErr string
	}{
		{
			name: "with pre-existing error",
			setupFn: func() (*ValueGenPool[googleUuid.UUID], error) {
				return NewValueGenPool[googleUuid.UUID](nil, errForTest("injected config error"))
			},
			wantErr: "injected config error",
		},
		{
			name:    "nil config",
			setupFn: func() (*ValueGenPool[googleUuid.UUID], error) { return NewValueGenPool[googleUuid.UUID](nil, nil) },
			wantErr: "config can't be nil",
		},
		{
			name: "nil generate function",
			setupFn: func() (*ValueGenPool[googleUuid.UUID], error) {
				cfg := &ValueGenPoolConfig[googleUuid.UUID]{
					ctx:                 context.Background(),
					telemetryService:    testTelemetryService,
					poolName:            "nil-fn-test",
					numWorkers:          1,
					poolSize:            1,
					maxLifetimeValues:   1,
					maxLifetimeDuration: time.Second,
					maintenanceInterval: cryptoutilSharedMagic.PoolMaintenanceInterval,
					generateFunction:    nil,
				}

				return NewValueGenPool(cfg, nil)
			},
			wantErr: "generate function can't be nil",
		},
		{
			name: "non-positive maintenance interval",
			setupFn: func() (*ValueGenPool[googleUuid.UUID], error) {
				cfg, err := NewValueGenPoolConfig(
					context.Background(), testTelemetryService,
					"invalid-maintenance-interval", 1, 1, 1, time.Second, uuidGenFn, false,
				)
				if err != nil {
					return nil, err
				}

				cfg.maintenanceInterval = 0

				return NewValueGenPool(cfg, nil)
			},
			wantErr: "maintenance interval must be positive",
		},
		{
			name: "large max lifetime values",
			setupFn: func() (*ValueGenPool[googleUuid.UUID], error) {
				return NewValueGenPool(NewValueGenPoolConfig(
					context.Background(), testTelemetryService,
					"large-values", 1, 1, ^uint64(0), time.Minute, uuidGenFn, false,
				))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			poolInstance, err := tc.setupFn()
			if tc.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, poolInstance)

			defer poolInstance.Cancel()
		})
	}
}

// TestGetAndGetMany_CanceledBeforeValue covers Get/GetMany cancel paths with verbose and non-verbose modes.
func TestGetAndGetMany_CanceledBeforeValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		verbose bool
		useGet  bool
	}{
		{name: "get non-verbose", verbose: false, useGet: true},
		{name: "get verbose", verbose: true, useGet: true},
		{name: "getmany non-verbose", verbose: false, useGet: false},
		{name: "getmany verbose", verbose: true, useGet: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
				tc.name, 1, 1, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, time.Minute, generateFn, tc.verbose,
			))
			require.NoError(t, err)
			require.NotNil(t, poolInstance)

			defer func() { close(blocks) }()
			defer poolInstance.Cancel()

			<-workerStarted
			poolInstance.Cancel()

			if tc.useGet {
				val := poolInstance.Get()
				require.Equal(t, googleUuid.UUID{}, val)
			} else {
				vals := poolInstance.GetMany(3)
				require.Empty(t, vals)
			}
		})
	}
}

func TestCancellationDuringWait_LogsVerboseBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		cancelOnDoneCall uint32
		run              func(*ValueGenPool[googleUuid.UUID])
	}{
		{
			name:             "getter cancellation after initial check",
			cancelOnDoneCall: 2,
			run: func(poolInstance *ValueGenPool[googleUuid.UUID]) {
				require.Equal(t, googleUuid.UUID{}, poolInstance.Get())
			},
		},
		{
			name:             "worker cancellation during permission wait",
			cancelOnDoneCall: 1,
			run: func(poolInstance *ValueGenPool[googleUuid.UUID]) {
				poolInstance.permissionChannel = make(chan struct{}, 1)
				poolInstance.permissionChannel <- struct{}{}

				poolInstance.generateWorker(1)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := &cancelOnDoneCallContext{
				Context:          context.Background(),
				done:             make(chan struct{}),
				cancelOnDoneCall: tc.cancelOnDoneCall,
			}
			poolInstance := &ValueGenPool[googleUuid.UUID]{
				cfg: &ValueGenPoolConfig[googleUuid.UUID]{
					ctx:              context.Background(),
					telemetryService: testTelemetryService,
					poolName:         tc.name,
					verbose:          true,
				},
				stopGeneratingCtx: ctx,
			}

			tc.run(poolInstance)
		})
	}
}

// TestWorker_Limits covers value and time limit error paths in generatePublishRelease and generateWorker.
func TestWorker_Limits(t *testing.T) {
	t.Parallel()

	uuidGenFn := func() (googleUuid.UUID, error) { return googleUuid.NewV7() }

	sleepyGenFn := func() (googleUuid.UUID, error) {
		time.Sleep(2 * time.Millisecond)

		return googleUuid.NewV7()
	}

	tests := []struct {
		name                string
		maxLifetimeValues   uint64
		maxLifetimeDuration time.Duration
		generateFn          func() (googleUuid.UUID, error)
		verbose             bool
		getValue            bool
		cancelBeforeDrain   bool
	}{
		{
			name: "value limit", maxLifetimeValues: 1, maxLifetimeDuration: time.Minute,
			generateFn: uuidGenFn,
			getValue:   true, cancelBeforeDrain: true,
		},
		{
			name: "value limit verbose", maxLifetimeValues: 1, maxLifetimeDuration: time.Minute,
			generateFn: uuidGenFn, verbose: true,
			getValue: true, cancelBeforeDrain: true,
		},
		{
			name: "time limit", maxLifetimeValues: cryptoutilSharedMagic.JoseJAMaxMaterials, maxLifetimeDuration: 1 * time.Millisecond,
			generateFn:        sleepyGenFn,
			cancelBeforeDrain: true,
		},
		{
			name: "time limit verbose", maxLifetimeValues: cryptoutilSharedMagic.JoseJAMaxMaterials, maxLifetimeDuration: 1 * time.Millisecond,
			generateFn: sleepyGenFn, verbose: true,
			cancelBeforeDrain: true,
		},
		{
			name: "time limit in gpr", maxLifetimeValues: cryptoutilSharedMagic.JoseJAMaxMaterials, maxLifetimeDuration: 1 * time.Nanosecond,
			generateFn: uuidGenFn, verbose: true,
		},
		{
			name: "value limit verbose in gpr", maxLifetimeValues: 1, maxLifetimeDuration: time.Minute,
			generateFn: uuidGenFn, verbose: true,
			getValue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
				context.Background(), testTelemetryService,
				tc.name, 1, 1, tc.maxLifetimeValues, tc.maxLifetimeDuration, tc.generateFn, tc.verbose,
			))
			require.NoError(t, err)
			require.NotNil(t, poolInstance)

			defer poolInstance.Cancel()

			if tc.getValue {
				val := poolInstance.Get()
				require.NotEqual(t, googleUuid.UUID{}, val)
			}

			if tc.cancelBeforeDrain {
				poolInstance.Cancel()
			}

			for range poolInstance.generateChannel { //nolint:revive
			}
		})
	}
}

func TestGeneratePublishRelease_ValueLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	poolConfig, err := NewValueGenPoolConfig(
		ctx, testTelemetryService,
		"value-limit-direct", 1, 1, 1, time.Minute,
		func() (googleUuid.UUID, error) { return googleUuid.NewV7() },
		false,
	)
	require.NoError(t, err)

	meter := testTelemetryService.MetricsProvider.Meter("cryptoutil.pool.value-limit-direct")
	generateCounterMetric, err := meter.Int64Counter("cryptoutil.pool.generate.count")
	require.NoError(t, err)

	poolInstance := &ValueGenPool[googleUuid.UUID]{
		poolStartTime:         time.Now().UTC(),
		generateCounter:       1,
		generateCounterMetric: generateCounterMetric,
		cfg:                   poolConfig,
		permissionChannel:     make(chan struct{}, 1),
	}

	poolInstance.permissionChannel <- struct{}{}

	err = poolInstance.generatePublishRelease(1, time.Now().UTC())
	require.ErrorContains(t, err, "reached generate limit")
}

func TestCancel_FullGenerateBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		poolSize uint32
	}{
		{name: "minimum test buffer", poolSize: cryptoutilSharedMagic.DefaultPoolConfigTest.MaxSize},
		{name: "maximum test buffer", poolSize: cryptoutilSharedMagic.TestPoolMaxSize},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			secondGenerationStarted := make(chan struct{})
			releaseSecondGeneration := make(chan struct{})

			var (
				generationSignalOnce sync.Once
				releaseOnce          sync.Once
				generateCount        atomic.Uint64
			)

			generateFn := func() (googleUuid.UUID, error) {
				if generateCount.Add(1) > uint64(tc.poolSize) {
					generationSignalOnce.Do(func() { close(secondGenerationStarted) })
					<-releaseSecondGeneration
				}

				return googleUuid.NewV7()
			}

			poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
				context.Background(), testTelemetryService,
				tc.name, cryptoutilSharedMagic.DefaultPoolConfigTest.NumWorkers, tc.poolSize,
				cryptoutilSharedMagic.JoseJADefaultMaxMaterials, time.Minute, generateFn, false,
			))
			require.NoError(t, err)

			require.NotNil(t, poolInstance)
			defer func() {
				poolInstance.Cancel()
				releaseOnce.Do(func() { close(releaseSecondGeneration) })
			}()

			select {
			case <-secondGenerationStarted:
			case <-time.After(cryptoutilSharedMagic.ClientShutdownRequestTimeout):
				t.Fatal("worker did not start generation after filling the buffer")
			}

			require.Len(t, poolInstance.generateChannel, int(tc.poolSize))

			poolInstance.Cancel()
			releaseOnce.Do(func() { close(releaseSecondGeneration) })

			require.Eventually(t, func() bool {
				return len(poolInstance.permissionChannel) == 0
			}, cryptoutilSharedMagic.ClientShutdownRequestTimeout, cryptoutilSharedMagic.WaitBeforeShutdownDuration,
				"worker must release its generation permission after cancellation")
		})
	}
}

func TestCloseChannelsThread_TickerPath(t *testing.T) {
	t.Parallel()

	generateFn := func() (googleUuid.UUID, error) {
		time.Sleep(2 * time.Millisecond)

		return googleUuid.NewV7()
	}

	cfg, err := NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"ticker-path", 1, 1, cryptoutilSharedMagic.JoseJAMaxMaterials, time.Millisecond, generateFn, false,
	)
	require.NoError(t, err)

	cfg.maintenanceInterval = time.Millisecond

	poolInstance, err := NewValueGenPool(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestCloseChannelsThread_InfinitePath covers the infinite-pool path in closeChannelsThread.
func TestCloseChannelsThread_InfinitePath(t *testing.T) {
	t.Parallel()

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"infinite-pool", 1, 1, 0, 0,
		cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	poolInstance.Cancel()

	for range poolInstance.generateChannel { //nolint:revive
	}
}

// TestWorker_PanicRecoveryInGeneratePublishRelease covers the panic recovery defer.
func TestWorker_PanicRecoveryInGeneratePublishRelease(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	generateFn := func() (googleUuid.UUID, error) {
		if callCount.Add(1) == 1 {
			panic("test panic in generate function")
		}

		return googleUuid.NewV7()
	}

	poolInstance, err := NewValueGenPool(NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"panic-recovery-gpr", 1, 1, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, time.Minute, generateFn, false,
	))
	require.NoError(t, err)
	require.NotNil(t, poolInstance)

	defer poolInstance.Cancel()

	val := poolInstance.Get()
	require.NotEqual(t, googleUuid.UUID{}, val)
}

// errForTest returns a simple error for test injection.
func errForTest(msg string) error {
	return &testError{msg: msg}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
