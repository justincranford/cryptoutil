package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// useful constants for indefinite pools; use smaller values for finite pools
const (
	MaxLifetimeValues   = ^uint64(0)                            // Max uint64 (= 2^64-1 = 18,446,744,073,709,551,615)
	MaxLifetimeDuration = time.Duration(int64(^uint64(0) >> 1)) // Max int64  (= 2^63-1 =  9,223,372,036,854,775,807 nanoseconds = 292.47 years)
)

type ValueGenPool[T any] struct {
	poolStartTime               time.Time              // needed to enforce maxLifetimeDuration in N workers and 1 closeChannelsThread thread
	generateCounter             uint64                 // needed to enforce maxLifetimeValues in N workers amd 1 closeChannelsThread thread
	getCounter                  uint64                 // metrics for how many times Get() was called successfully
	cfg                         *ValueGenPoolConfig[T] // container for all configuration parameters, including telemetryService and poolName
	cancellableCtx              context.Context        // Exposes Done() signal to N workers, 1 closeChannelsThread, and M getters
	cancelFunction              context.CancelFunc     // Cancel() invokes this to raise the Done() signal
	permissionChannel           chan struct{}          // N workers use this channel to get and release permissions (up to pool size); generate can be expensive (e.g. RSA-4096)
	valueChannel                chan T                 // N workers use this channel to publish generated Values
	cancelOnce                  sync.Once              // Cancel() uses this to guard raising the Done() signal, and log if Cancel() was already called
	meter                       metric.Meter
	getDurationHistogram        metric.Float64Histogram
	permissionDurationHistogram metric.Float64Histogram
	generateDurationHistogram   metric.Float64Histogram
}

type ValueGenPoolConfig[T any] struct {
	ctx                 context.Context
	telemetryService    *cryptoutilTelemetry.TelemetryService // TODO change generateCounter and getCounter to telemetryService.MetricsProvider.Counter()
	poolName            string
	numWorkers          uint32
	poolSize            uint32
	maxLifetimeValues   uint64
	maxLifetimeDuration time.Duration
	generateFunction    func() (T, error)
}

// NewValueGenPool supports indefinite pools, or finite pools based on maxTime and/or maxValues
func NewValueGenPool[T any](cfg *ValueGenPoolConfig[T], err error) (*ValueGenPool[T], error) {
	poolStartTime := time.Now()
	if err != nil { // config and err are from the call to NewValueGenPoolConfig, check the error value
		return nil, fmt.Errorf("failed to create pool config: %w", err)
	} else if err := validateConfig(cfg); err != nil { // check the config value
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	meter := cfg.telemetryService.MetricsProvider.Meter("cryptoutil.pool."+cfg.poolName, []metric.MeterOption{
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "workers", Value: attribute.IntValue(int(cfg.numWorkers))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "size", Value: attribute.IntValue(int(cfg.poolSize))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "values", Value: attribute.Int64Value(int64(cfg.maxLifetimeValues))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "duration", Value: attribute.Int64Value(int64(cfg.maxLifetimeDuration))}),
		metric.WithInstrumentationAttributes(attribute.KeyValue{Key: "type", Value: attribute.StringValue(fmt.Sprintf("%T", *new(T)))}), // record the type of T in the metric attributes
	}...)
	getDurationHistogram, err := meter.Float64Histogram("cryptoutil.pool.get", metric.WithUnit("s"))
	permissionDurationHistogram, err := meter.Float64Histogram("cryptoutil.pool.permission", metric.WithUnit("s"))
	generateDurationHistogram, err := meter.Float64Histogram("cryptoutil.pool.generate", metric.WithUnit("s"))

	cancellableCtx, cancelFunction := context.WithCancel(cfg.ctx)
	valuePool := &ValueGenPool[T]{
		poolStartTime:               poolStartTime,
		cfg:                         cfg,
		cancellableCtx:              cancellableCtx,
		cancelFunction:              cancelFunction,
		permissionChannel:           make(chan struct{}, cfg.poolSize),
		valueChannel:                make(chan T, cfg.poolSize),
		meter:                       meter,
		getDurationHistogram:        getDurationHistogram,
		permissionDurationHistogram: permissionDurationHistogram,
		generateDurationHistogram:   generateDurationHistogram,
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

func NewValueGenPoolConfig[T any](ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, poolName string, numWorkers uint32, poolSize uint32, maxLifetimeValues uint64, maxLifetimeDuration time.Duration, generateFunction func() (T, error)) (*ValueGenPoolConfig[T], error) {
	config := &ValueGenPoolConfig[T]{
		ctx:                 ctx,
		telemetryService:    telemetryService,
		poolName:            poolName,
		numWorkers:          numWorkers,
		poolSize:            poolSize,
		maxLifetimeValues:   maxLifetimeValues,
		maxLifetimeDuration: maxLifetimeDuration,
		generateFunction:    generateFunction,
	}
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return config, nil
}

func (pool *ValueGenPool[T]) Name() string {
	return pool.cfg.poolName
}

func (pool *ValueGenPool[T]) Get() T {
	startTime := time.Now()
	pool.cfg.telemetryService.Slogger.Debug("getting", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
	select {
	case <-pool.cancellableCtx.Done(): // someone called Cancel()
		pool.cfg.telemetryService.Slogger.Debug("cancelled", "pool", pool.cfg.poolName, "worker", time.Since(startTime).Seconds())
		var zero T
		return zero
	case value := <-pool.valueChannel: // block wait for a generated value to be published by a worker
		pool.cfg.telemetryService.Slogger.Debug("received", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
		getCounter := atomic.AddUint64(&pool.getCounter, 1)
		defer func() {
			pool.cfg.telemetryService.Slogger.Debug("got", "pool", pool.cfg.poolName, "get", getCounter, "duration", time.Since(startTime).Seconds())
			pool.getDurationHistogram.Record(pool.cfg.ctx, time.Since(startTime).Seconds())
		}()
		return value
	}
}

func (pool *ValueGenPool[T]) Cancel() {
	startTime := time.Now()
	didCancel := false
	pool.cancelOnce.Do(func() {
		defer func() {
			pool.cfg.telemetryService.Slogger.Debug("cancelled ok", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
		}()
		pool.cancelFunction() // raise Done() signal to N workers, 1 closeChannelsThread, and M getters
		pool.cancelFunction = nil
		didCancel = true
	})
	if !didCancel {
		pool.cfg.telemetryService.Slogger.Warn("already cancelled", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
	}
}

func (pool *ValueGenPool[T]) generateWorker(workerNum uint32) {
	startTime := time.Now()
	defer func() {
		if r := recover(); r != nil {
			pool.cfg.telemetryService.Slogger.Error("worker panic recovered", "pool", pool.cfg.poolName, "worker", workerNum, "panic", r)
		}
		pool.cfg.telemetryService.Slogger.Debug("worker done", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}()

	pool.cfg.telemetryService.Slogger.Debug("worker started", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	for {
		startPermissionTime := time.Now()
		pool.cfg.telemetryService.Slogger.Debug("wait for permission", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.cancellableCtx.Done(): // someone called Cancel()
			pool.cfg.telemetryService.Slogger.Debug("worker canceled before generate", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.permissionChannel <- struct{}{}: // acquire permission to generate
			pool.generateDurationHistogram.Record(pool.cfg.ctx, time.Since(startPermissionTime).Seconds())
			err := pool.generatePublishRelease(workerNum, startTime) // worker has permission; attempt to generate, but always release permission, even if there is an error or panic
			if err != nil {
				pool.cfg.telemetryService.Slogger.Debug("worker stopped", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)
				return
			}
		}
	}
}

func (pool *ValueGenPool[T]) generatePublishRelease(workerNum uint32, startTime time.Time) error {
	defer func() { // always release permission, even if there was an error or panic
		if r := recover(); r != nil {
			pool.cfg.telemetryService.Slogger.Error("recovered from panic", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "panic", r)
		}
		<-pool.permissionChannel
		pool.cfg.telemetryService.Slogger.Debug("released permission", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}()
	pool.cfg.telemetryService.Slogger.Debug("permission granted", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())

	generateCounter := atomic.AddUint64(&pool.generateCounter, 1)

	timeLimitReached := (pool.cfg.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.cfg.maxLifetimeDuration)
	if timeLimitReached || (pool.cfg.maxLifetimeValues > 0 && atomic.LoadUint64(&pool.generateCounter) > pool.cfg.maxLifetimeValues) {
		if timeLimitReached {
			pool.cfg.telemetryService.Slogger.Warn("time limit reached", "pool", pool.cfg.poolName)
		} else {
			pool.cfg.telemetryService.Slogger.Warn("generate limit reached", "pool", pool.cfg.poolName)
		}
		pool.cfg.telemetryService.Slogger.Warn("limit reached", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		// pool.Cancel() // IMPORTANT: don't call Cancel(), otherwise getters won't finish getting generated values from this finite pool
		return fmt.Errorf("pool %s reached max lifetime values %d or max lifetime duration %s", pool.cfg.poolName, pool.cfg.maxLifetimeValues, pool.cfg.maxLifetimeDuration)
	}

	generateStartTime := time.Now() // reset startTime to measure the duration of the generation
	value, err := pool.cfg.generateFunction()
	generateDuration := time.Since(generateStartTime).Seconds()
	defer func() {
		pool.generateDurationHistogram.Record(pool.cfg.ctx, generateDuration)
	}()
	if err != nil {
		pool.cfg.telemetryService.Slogger.Error("generation failed", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)
		pool.Cancel() // signal all workers to stop (e.g. does generateFunction() have a bug?)
		return fmt.Errorf("pool %s worker %d failed to generate value: %w", pool.cfg.poolName, workerNum, err)
	}
	pool.cfg.telemetryService.Slogger.Debug("Generated", "pool", pool.cfg.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())

	select {
	case <-pool.cancellableCtx.Done(): // someone called Cancel(), skip the blocking wait to publish the generated value (i.e. throw it away)
		pool.cfg.telemetryService.Slogger.Debug("canceled before publish", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	case pool.valueChannel <- value:
		pool.cfg.telemetryService.Slogger.Debug("published", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}
	return nil
}

func (pool *ValueGenPool[T]) closeChannelsThread(waitForWorkers *sync.WaitGroup) {
	if pool.cfg.maxLifetimeDuration == 0 && pool.cfg.maxLifetimeValues == 0 {
		// this is an infinite pool; no need to periodically wake up to check limits, because there are no limits
		select {
		case <-pool.cancellableCtx.Done(): // block waiting indefinitely until someone calls Cancel()
			pool.cfg.telemetryService.Slogger.Debug("cancelled", "pool", pool.cfg.poolName)
			pool.closeChannels(waitForWorkers)
			return
		}
	}

	// this is a finite pool; periodically wake up and check if one of the pool limits has been reached (e.g. time), especially if all workers and getters are idle
	ticker := time.NewTicker(500 * time.Millisecond) // time keeps on ticking ticking ticking... into the future
	defer ticker.Stop()
	for {
		select {
		case <-pool.cancellableCtx.Done(): // someone called Cancel()
			pool.cfg.telemetryService.Slogger.Debug("cancelled", "pool", pool.cfg.poolName)
			pool.closeChannels(waitForWorkers)
			return
		case <-ticker.C: // wake up and check the limits
			timeLimitReached := (pool.cfg.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.cfg.maxLifetimeDuration)
			if timeLimitReached || (pool.cfg.maxLifetimeValues > 0 && atomic.LoadUint64(&pool.generateCounter) > pool.cfg.maxLifetimeValues) {
				if timeLimitReached {
					pool.cfg.telemetryService.Slogger.Warn("time limit reached", "pool", pool.cfg.poolName)
				} else {
					pool.cfg.telemetryService.Slogger.Warn("generate limit reached", "pool", pool.cfg.poolName)
				}
				pool.Cancel() // signal to all workers to stop
				pool.closeChannels(waitForWorkers)
				return
			}
		}
	}
}

func (pool *ValueGenPool[T]) closeChannels(waitForWorkers *sync.WaitGroup) {
	pool.cfg.telemetryService.Slogger.Debug("waiting for workers", "pool", pool.cfg.poolName)
	waitForWorkers.Wait() // wait for all workers to stop before closing permissionChannel and valueChannel
	pool.cfg.telemetryService.Slogger.Debug("closing channels", "pool", pool.cfg.poolName)
	close(pool.valueChannel)
	close(pool.permissionChannel)
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
