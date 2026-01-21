// Copyright (c) 2025 Justin Cranford
//
//

// Package pool provides high-performance object pooling for cryptographic operations.
package pool

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// ValueGenPool is a high-performance generic pool that pre-generates values using worker goroutines.
type ValueGenPool[T any] struct {
	poolStartTime               time.Time               // needed to enforce maxLifetimeDuration in N workers and 1 closeChannelsThread thread
	generateCounter             uint64                  // needed to enforce maxLifetimeValues   in N workers amd 1 closeChannelsThread thread
	getCounter                  uint64                  // needed to enforce logic, but metrics use telemetry counters
	generateCounterMetric       metric.Int64Counter     // telemetry counter for how many times Generate() was called
	getCounterMetric            metric.Int64Counter     // telemetry counter for how many times Get() was called successfully
	cfg                         *ValueGenPoolConfig[T]  // container for all configuration parameters, including telemetryService and poolName
	stopGeneratingCtx           context.Context         // Exposes Done() signal to N workers, 1 closeChannelsThread, and M getters
	stopGeneratingFunction      context.CancelFunc      // Cancel() invokes this to raise the Done() signal
	stopGeneratingOnce          sync.Once               // Cancel() uses this to guard raising the Done() signal, and log if Cancel() was already called
	permissionChannel           chan struct{}           // N workers use this channel to get and release permissions (up to pool size); generate can be expensive (e.g. RSA-4096)
	generateChannel             chan T                  // N workers use this channel to publish generated Values
	getDurationHistogram        metric.Float64Histogram // telemetry histogram metric (i.e. cumulative time & count, average, time buckets & percentiles) of wait for get
	permissionDurationHistogram metric.Float64Histogram // telemetry histogram metric (i.e. cumulative time & count, average, time buckets & percentiles) of wait for generate permission
	generateDurationHistogram   metric.Float64Histogram // telemetry histogram metric (i.e. cumulative time & count, average, time buckets & percentiles) of wait for generate completed
}

// ValueGenPoolConfig holds configuration options for creating a ValueGenPool.
type ValueGenPoolConfig[T any] struct {
	ctx                 context.Context
	telemetryService    *cryptoutilTelemetry.TelemetryService // telemetry service for metrics and logging
	poolName            string
	numWorkers          uint32
	poolSize            uint32
	maxLifetimeValues   uint64
	maxLifetimeDuration time.Duration
	generateFunction    func() (T, error)
	verbose             bool
}

// NewValueGenPool supports indefinite pools, or finite pools based on maxTime and/or maxValues.
func NewValueGenPool[T any](cfg *ValueGenPoolConfig[T], err error) (*ValueGenPool[T], error) {
	poolStartTime := time.Now().UTC()

	if err != nil { // config and err are from the call to NewValueGenPoolConfig, check the error value
		return nil, fmt.Errorf("failed to create pool config: %w", err)
	} else if err := validateConfig(cfg); err != nil { // check the config value
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Safe conversion with bounds checking
	var maxLifetimeValuesInt64 int64
	if cfg.maxLifetimeValues <= cryptoutilMagic.MaxPoolLifetimeValues {
		maxLifetimeValuesInt64 = int64(cfg.maxLifetimeValues)
	} else {
		maxLifetimeValuesInt64 = cryptoutilMagic.MaxPoolLifetimeValuesInt64
	}

	meter := cfg.telemetryService.MetricsProvider.Meter("cryptoutil.pool."+cfg.poolName, []metric.MeterOption{
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "workers", Value: attribute.IntValue(int(cfg.numWorkers))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "size", Value: attribute.IntValue(int(cfg.poolSize))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "values", Value: attribute.Int64Value(maxLifetimeValuesInt64)}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "duration", Value: attribute.Int64Value(int64(cfg.maxLifetimeDuration))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "type", Value: attribute.StringValue(fmt.Sprintf("%T", *new(T)))}), // record the type of T in the metric attributes
	}...)

	getHistogramMetric, err := meter.Float64Histogram("cryptoutil.pool.get", metric.WithUnit("ms"))
	if err != nil {
		return nil, fmt.Errorf("failed to create get metric: %w", err)
	}

	permissionHistogramMetric, err := meter.Float64Histogram("cryptoutil.pool.permission", metric.WithUnit("ms"))
	if err != nil {
		return nil, fmt.Errorf("failed to create permission metric: %w", err)
	}

	generateHistogramMetric, err := meter.Float64Histogram("cryptoutil.pool.generate", metric.WithUnit("ms"))
	if err != nil {
		return nil, fmt.Errorf("failed to create generate metric: %w", err)
	}

	getCounterMetric, err := meter.Int64Counter("cryptoutil.pool.get.count")
	if err != nil {
		return nil, fmt.Errorf("failed to create get counter metric: %w", err)
	}

	generateCounterMetric, err := meter.Int64Counter("cryptoutil.pool.generate.count")
	if err != nil {
		return nil, fmt.Errorf("failed to create generate counter metric: %w", err)
	}

	stopGeneratingCtx, stopGeneratingFunction := context.WithCancel(cfg.ctx)
	valuePool := &ValueGenPool[T]{
		poolStartTime:               poolStartTime,
		generateCounter:             0,
		getCounter:                  0,
		generateCounterMetric:       generateCounterMetric,
		getCounterMetric:            getCounterMetric,
		cfg:                         cfg,
		stopGeneratingCtx:           stopGeneratingCtx,
		stopGeneratingFunction:      stopGeneratingFunction,
		permissionChannel:           make(chan struct{}, cfg.poolSize),
		generateChannel:             make(chan T, cfg.poolSize),
		getDurationHistogram:        getHistogramMetric,
		permissionDurationHistogram: permissionHistogramMetric,
		generateDurationHistogram:   generateHistogramMetric,
	}

	var waitForWorkers sync.WaitGroup                 // closeChannelsThread uses this to wait for N worker to finish, so it is safe to close permissionChannel and valueChannel
	go valuePool.closeChannelsThread(&waitForWorkers) // close channels when it is safe; after all workers are done

	for workerNum := uint32(1); workerNum <= valuePool.cfg.numWorkers; workerNum++ {
		waitForWorkers.Add(1)

		go func() {
			defer waitForWorkers.Done()

			valuePool.generateWorker(workerNum)
		}()
	}

	return valuePool, nil
}

// NewValueGenPoolConfig creates a new configuration for a ValueGenPool.
func NewValueGenPoolConfig[T any](ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, poolName string, numWorkers, poolSize uint32, maxLifetimeValues uint64, maxLifetimeDuration time.Duration, generateFunction func() (T, error), verbose bool) (*ValueGenPoolConfig[T], error) {
	config := &ValueGenPoolConfig[T]{
		ctx:                 ctx,
		telemetryService:    telemetryService,
		poolName:            poolName,
		numWorkers:          numWorkers,
		poolSize:            poolSize,
		maxLifetimeValues:   maxLifetimeValues,
		maxLifetimeDuration: maxLifetimeDuration,
		generateFunction:    generateFunction,
		verbose:             verbose,
	}
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// Name returns the name of the pool.
func (pool *ValueGenPool[T]) Name() string {
	return pool.cfg.poolName
}

// Get retrieves a pre-generated value from the pool, blocking until one is available.
func (pool *ValueGenPool[T]) Get() T {
	startTime := time.Now().UTC()
	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("getting", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
	}

	select {
	case <-pool.stopGeneratingCtx.Done(): // someone called Cancel()
		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Debug("get canceled", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
		}

		var zero T

		return zero
	case value := <-pool.generateChannel: // block wait for a generated value to be published by a worker
		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Debug("received", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
		}

		getCounter := atomic.AddUint64(&pool.getCounter, 1)
		pool.getCounterMetric.Add(pool.cfg.ctx, 1)

		defer func() {
			if pool.cfg.verbose {
				pool.cfg.telemetryService.Slogger.Debug("got", "pool", pool.cfg.poolName, "get", getCounter, "duration", time.Since(startTime).Seconds())
			}

			pool.getDurationHistogram.Record(pool.cfg.ctx, float64(time.Since(startTime).Milliseconds()))
		}()

		return value
	}
}

// GetMany retrieves multiple pre-generated values from the pool.
func (pool *ValueGenPool[T]) GetMany(numValues int) []T {
	if numValues <= 0 {
		return nil
	}

	startTime := time.Now().UTC()
	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("getting many", "pool", pool.cfg.poolName, "count", numValues, "duration", time.Since(startTime).Seconds())
	}

	values := make([]T, 0, numValues)

	var zero T

	for range numValues {
		value := pool.Get()
		if reflect.DeepEqual(value, zero) {
			if pool.cfg.verbose {
				pool.cfg.telemetryService.Slogger.Debug("get many canceled", "pool", pool.cfg.poolName, "requested", numValues, "received", len(values), "duration", time.Since(startTime).Seconds())
			}

			break
		}

		values = append(values, value)
	}

	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("got many", "pool", pool.cfg.poolName, "count", len(values), "duration", time.Since(startTime).Seconds())
	}

	return values
}

// Cancel stops all pool workers and releases resources.
func (pool *ValueGenPool[T]) Cancel() {
	startTime := time.Now().UTC()
	didCancelThisTime := false

	pool.stopGeneratingOnce.Do(func() {
		defer func() {
			if pool.cfg.verbose {
				pool.cfg.telemetryService.Slogger.Debug("canceled ok", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
			}
		}()

		pool.stopGeneratingFunction() // raise Done() signal to N workers, 1 closeChannelsThread, and M getters
		pool.stopGeneratingFunction = nil
		didCancelThisTime = true
	})

	if !didCancelThisTime {
		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Warn("already canceled", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
		}
	}
}

func (pool *ValueGenPool[T]) generateWorker(workerNum uint32) {
	startTime := time.Now().UTC()

	defer func() {
		if panicValue := recover(); panicValue != nil {
			pool.cfg.telemetryService.Slogger.Error("worker panic recovered", "pool", pool.cfg.poolName, "worker", workerNum, "panic", panicValue, "stack", string(debug.Stack()))
		}

		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Debug("worker done", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}
	}()

	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("worker started", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}

	for {
		startPermissionTime := time.Now().UTC()

		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Debug("wait for permission", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}

		select {
		case <-pool.stopGeneratingCtx.Done(): // someone called Cancel()
			if pool.cfg.verbose {
				pool.cfg.telemetryService.Slogger.Debug("worker canceled before generate", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			}

			return
		case pool.permissionChannel <- struct{}{}: // acquire permission to generate
			pool.generateDurationHistogram.Record(pool.cfg.ctx, float64(time.Since(startPermissionTime).Milliseconds()))

			err := pool.generatePublishRelease(workerNum, startTime) // attempt to generate inside a function, where permission is always released, even if there is an error or panic
			if err != nil {                                          // if there was an error, log it and stop the worker
				pool.cfg.telemetryService.Slogger.Error("worker error", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)

				return
			}
		}
	}
}

// IMPORTANT: don't call Cancel() in this function, because it waits for all permissions to be released, and this function doesn't release its permission until it returns.
func (pool *ValueGenPool[T]) generatePublishRelease(workerNum uint32, startTime time.Time) error {
	defer func() { // always release permission, even if there was an error or panic
		if panicValue := recover(); panicValue != nil {
			pool.cfg.telemetryService.Slogger.Error("recovered from panic", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "panic", panicValue, "stack", string(debug.Stack()))
		}

		<-pool.permissionChannel // release permission to generate, so other workers can generate, or Cancel() can close the channel

		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Debug("released permission", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}
	}()

	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("permission granted", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}

	// permission granted, worker is promising to generate a value if time not exceeded, of if this counter value doesn't exceed the generate limit
	generateCounter := atomic.AddUint64(&pool.generateCounter, 1)
	pool.generateCounterMetric.Add(pool.cfg.ctx, 1)

	if pool.cfg.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.cfg.maxLifetimeDuration {
		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Warn("time limit reached", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}

		pool.cfg.telemetryService.Slogger.Info("worker done", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "info", fmt.Sprintf("pool %s reached max lifetime %s", pool.cfg.poolName, pool.cfg.maxLifetimeDuration))

		return fmt.Errorf("pool %s reached max lifetime %s", pool.cfg.poolName, pool.cfg.maxLifetimeDuration)
	}

	if pool.cfg.maxLifetimeValues > 0 && generateCounter > pool.cfg.maxLifetimeValues {
		if pool.cfg.verbose {
			pool.cfg.telemetryService.Slogger.Warn("generate limit reached", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}

		pool.cfg.telemetryService.Slogger.Info("worker done", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "info", fmt.Sprintf("pool %s reached generate limit %d", pool.cfg.poolName, pool.cfg.maxLifetimeValues))

		return fmt.Errorf("pool %s reached generate limit %d", pool.cfg.poolName, pool.cfg.maxLifetimeValues)
	}

	// at this point, the worker is committed to generating a value, so only stop if there is an error
	generateStartTime := time.Now().UTC()
	value, err := pool.cfg.generateFunction()
	generateDuration := float64(time.Since(generateStartTime).Milliseconds())

	defer func() {
		pool.generateDurationHistogram.Record(pool.cfg.ctx, generateDuration)
	}()

	if err != nil {
		pool.cfg.telemetryService.Slogger.Error("generation failed", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)
		pool.Cancel() // signal all workers to stop (e.g. does generateFunction() have a bug?)

		return fmt.Errorf("pool %s worker %d failed to generate value: %w", pool.cfg.poolName, workerNum, err)
	}

	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("generated", "pool", pool.cfg.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
	}

	pool.generateChannel <- value

	if pool.cfg.verbose {
		pool.cfg.telemetryService.Slogger.Debug("published", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}

	return nil
}

func (pool *ValueGenPool[T]) closeChannelsThread(waitForWorkers *sync.WaitGroup) {
	if pool.cfg.maxLifetimeDuration == 0 && pool.cfg.maxLifetimeValues == 0 {
		// this is an infinite pool; no need to periodically wake up to check limits, because there are no limits
		<-pool.stopGeneratingCtx.Done() // block waiting indefinitely until someone calls Cancel()
		pool.cfg.telemetryService.Slogger.Debug("canceled", "pool", pool.cfg.poolName)
		pool.closePermissionAndGenerateChannels(waitForWorkers)

		return
	}

	// this is a finite pool; periodically wake up and check if one of the pool limits has been reached (e.g. time), especially if all workers and getters are idle
	ticker := time.NewTicker(cryptoutilMagic.PoolMaintenanceInterval) // time keeps on ticking ticking ticking... into the future

	defer func() { ticker.Stop() }() //nolint:errcheck

	for {
		select {
		case <-pool.stopGeneratingCtx.Done(): // someone called Cancel()
			pool.cfg.telemetryService.Slogger.Debug("canceled", "pool", pool.cfg.poolName)
			pool.closePermissionAndGenerateChannels(waitForWorkers)

			return
		case <-ticker.C: // wake up and check the limits
			timeLimitReached := (pool.cfg.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.cfg.maxLifetimeDuration)
			if timeLimitReached || (pool.cfg.maxLifetimeValues > 0 && atomic.LoadUint64(&pool.generateCounter) > pool.cfg.maxLifetimeValues) {
				if timeLimitReached {
					pool.cfg.telemetryService.Slogger.Warn("time limit reached", "pool", pool.cfg.poolName)
				} else {
					pool.cfg.telemetryService.Slogger.Warn("generate limit reached", "pool", pool.cfg.poolName)
				}

				pool.Cancel() // signal to all workers to stop generating
				pool.closePermissionAndGenerateChannels(waitForWorkers)

				return
			}
		}
	}
}

func (pool *ValueGenPool[T]) closePermissionAndGenerateChannels(waitForWorkers *sync.WaitGroup) {
	pool.cfg.telemetryService.Slogger.Debug("waiting for workers", "pool", pool.cfg.poolName)
	waitForWorkers.Wait() // wait for all workers to stop before closing permissionChannel and valueChannel
	pool.cfg.telemetryService.Slogger.Debug("closing channels", "pool", pool.cfg.poolName)
	close(pool.generateChannel)   //nolint:errcheck
	close(pool.permissionChannel) //nolint:errcheck
}

func validateConfig[T any](config *ValueGenPoolConfig[T]) error {
	if config == nil {
		return fmt.Errorf("config can't be nil")
	} else if config.ctx == nil {
		return fmt.Errorf("context can't be nil")
	} else if config.telemetryService == nil {
		return fmt.Errorf("telemetry service can't be nil")
	} else if len(config.poolName) == 0 {
		return fmt.Errorf("name can't be empty")
	} else if config.numWorkers == 0 {
		return fmt.Errorf("number of workers can't be 0")
	} else if config.poolSize == 0 {
		return fmt.Errorf("pool size can't be 0")
	} else if config.maxLifetimeValues == 0 {
		return fmt.Errorf("max lifetime values can't be 0")
	} else if config.maxLifetimeDuration <= 0 {
		return fmt.Errorf("max lifetime duration must be positive and non-zero")
	} else if config.numWorkers > config.poolSize {
		return fmt.Errorf("number of workers can't be greater than pool size")
	} else if uint64(config.poolSize) > config.maxLifetimeValues {
		return fmt.Errorf("pool size can't be greater than max lifetime values")
	} else if config.generateFunction == nil {
		return fmt.Errorf("generate function can't be nil")
	}

	return nil
}
