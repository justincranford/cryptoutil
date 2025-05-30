package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

const (
	MaxLifetimeValues   = ^uint64(0)                            // Max uint64 (= 2^64-1 = 18,446,744,073,709,551,615)
	MaxLifetimeDuration = time.Duration(int64(^uint64(0) >> 1)) // Max int64  (= 2^63-1 =  9,223,372,036,854,775,807 nanoseconds = 292.47 years)
)

type ValueGenPool[T any] struct {
	poolStartTime         time.Time // used to enforce maxLifetimeDuration in N generateWorker threads and 1 monitorShutdown thread
	cfg                   *ValueGenPoolConfig[T]
	wrappedCtx            context.Context    // Cancel() calls cancelWorkersFunction which makes Done() signal available to all of the N generateWorker threads and 1 monitorShutdown thread
	cancelWorkersFunction context.CancelFunc // This is the associated cancel function for wrappedCtx; the cancel function is called by Cancel()
	permissionChannel     chan struct{}      // N generateWorker threads block wait on this channel before generating value, because value generation (e.g. RSA-4096) can be resource expensive
	valueChannel          chan T             // N generateWorker threads publish generated Values to this channel
	waitForWorkers        sync.WaitGroup     // Cancel() uses this to wait for N generateWorker threads to finish before closing valueChannel and permissionChannel
	closeOnce             sync.Once
	generateCounter       uint64
	getCounter            uint64
}

type ValueGenPoolConfig[T any] struct {
	ctx                 context.Context
	telemetryService    *cryptoutilTelemetry.TelemetryService
	poolName            string
	numWorkers          uint32
	poolSize            uint32
	maxLifetimeValues   uint64
	maxLifetimeDuration time.Duration
	generateFunction    func() (T, error)
}

// NewValueGenPool supports finite or indefinite pools
func NewValueGenPool[T any](config *ValueGenPoolConfig[T], err error) (*ValueGenPool[T], error) {
	poolStartTime := time.Now()
	if err != nil {
		return nil, fmt.Errorf("failed to create pool config: %w", err)
	}
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	wrappedCtx, cancelFunction := context.WithCancel(config.ctx)
	valuePool := &ValueGenPool[T]{
		poolStartTime:         poolStartTime,
		cfg:                   config,
		wrappedCtx:            wrappedCtx,
		cancelWorkersFunction: cancelFunction,
		permissionChannel:     make(chan struct{}, config.poolSize),
		valueChannel:          make(chan T, config.poolSize),
	}
	go valuePool.monitorShutdown()
	for workerNum := uint32(1); workerNum <= valuePool.cfg.numWorkers; workerNum++ {
		valuePool.waitForWorkers.Add(1)
		go func() {
			defer valuePool.waitForWorkers.Done()
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

func (pool *ValueGenPool[T]) Name() string {
	return pool.cfg.poolName
}

func (pool *ValueGenPool[T]) monitorShutdown() {
	ticker := time.NewTicker(500 * time.Millisecond) // time keeps on ticking ticking ticking... into the future
	defer ticker.Stop()
	for {
		select {
		case <-pool.wrappedCtx.Done(): // someone else called Cancel()
			pool.cfg.telemetryService.Slogger.Debug("cancelled", "pool", pool.cfg.poolName)
			return
		case <-ticker.C:
			reachedLimit := (pool.cfg.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.cfg.maxLifetimeDuration) || (pool.cfg.maxLifetimeValues > 0 && atomic.LoadUint64(&pool.generateCounter) > pool.cfg.maxLifetimeValues)
			if reachedLimit {
				pool.cfg.telemetryService.Slogger.Warn("limit reached", "pool", pool.cfg.poolName)
				pool.shutdown()
				return
			}
		}
	}
}

func (pool *ValueGenPool[T]) generateWorker(workerNum uint32) {
	startTime := time.Now()
	defer func() {
		if r := recover(); r != nil {
			pool.cfg.telemetryService.Slogger.Error("Worker panic recovered", "pool", pool.cfg.poolName, "worker", workerNum, "panic", r)
		}
		pool.cfg.telemetryService.Slogger.Debug("Worker done", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}()
	pool.cfg.telemetryService.Slogger.Debug("Worker started", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	for {
		pool.cfg.telemetryService.Slogger.Debug("check", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.wrappedCtx.Done(): // someone called Cancel()
			pool.cfg.telemetryService.Slogger.Debug("Worker canceled", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.permissionChannel <- struct{}{}: // acquire permission to generate
			err := pool.generateValueAndReleasePermission(workerNum, startTime) // Use method with defer to guarantee permission release even if there is an error or panic
			if err != nil {
				pool.cfg.telemetryService.Slogger.Debug("worker stopped", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)
				return
			}
		}
	}
}

func (pool *ValueGenPool[T]) generateValueAndReleasePermission(workerNum uint32, startTime time.Time) error {
	defer func() {
		if r := recover(); r != nil {
			pool.cfg.telemetryService.Slogger.Error("Recovered from panic", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "panic", r)
		}
		<-pool.permissionChannel
		pool.cfg.telemetryService.Slogger.Debug("Released permission", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}()
	pool.cfg.telemetryService.Slogger.Debug("Permission granted", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())

	generateCounter := atomic.AddUint64(&pool.generateCounter, 1)
	if (pool.cfg.maxLifetimeValues > 0 && generateCounter > pool.cfg.maxLifetimeValues) || (pool.cfg.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.cfg.maxLifetimeDuration) {
		pool.cfg.telemetryService.Slogger.Warn("Won't Generate", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		return fmt.Errorf("pool %s reached max lifetime values %d or max lifetime duration %s", pool.cfg.poolName, pool.cfg.maxLifetimeValues, pool.cfg.maxLifetimeDuration)
	}

	value, err := pool.cfg.generateFunction()
	if err != nil {
		pool.cfg.telemetryService.Slogger.Error("Value generation failed", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)
		pool.cancelWorkersFunction() // signal to all workers to stop
		return fmt.Errorf("pool %s worker %d failed to generate value: %w", pool.cfg.poolName, workerNum, err)
	}
	pool.cfg.telemetryService.Slogger.Debug("Generated", "pool", pool.cfg.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())

	select {
	case <-pool.wrappedCtx.Done(): // someone called Cancel()
		pool.cfg.telemetryService.Slogger.Debug("Context canceled during publish", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	case pool.valueChannel <- value:
		pool.cfg.telemetryService.Slogger.Debug("Value added to channel", "pool", pool.cfg.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}
	return nil
}

func (pool *ValueGenPool[T]) Get() T {
	startTime := time.Now()
	pool.cfg.telemetryService.Slogger.Debug("getting", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
	select {
	case <-pool.wrappedCtx.Done(): // someone called Cancel()
		pool.cfg.telemetryService.Slogger.Debug("closed", "pool", pool.cfg.poolName, "worker", time.Since(startTime).Seconds())
		var zero T
		return zero
	case value := <-pool.valueChannel:
		pool.cfg.telemetryService.Slogger.Debug("received", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
		getCounter := atomic.AddUint64(&pool.getCounter, 1)
		defer func() {
			pool.cfg.telemetryService.Slogger.Debug("got", "pool", pool.cfg.poolName, "get", getCounter, "duration", time.Since(startTime).Seconds())
		}()
		return value
	}
}

func (pool *ValueGenPool[T]) Cancel() {
	pool.cancelWorkersFunction()
}

func (pool *ValueGenPool[T]) shutdown() {
	pool.closeOnce.Do(func() {
		startTime := time.Now()

		if pool.cancelWorkersFunction == nil {
			defer func() {
				pool.cfg.telemetryService.Slogger.Warn("already shutdown", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
			}()
		} else {
			defer func() {
				pool.cfg.telemetryService.Slogger.Info("shutdown ok", "pool", pool.cfg.poolName, "duration", time.Since(startTime).Seconds())
			}()
			pool.cancelWorkersFunction() // send Done() signal to N generateWorker threads and 1 monitorShutdown thread (if they are still listening to the shared context.WithCancel)
			pool.cancelWorkersFunction = nil
		}

		pool.waitForWorkers.Wait()

		if pool.valueChannel != nil {
			close(pool.valueChannel)
			pool.valueChannel = nil
		}
		if pool.permissionChannel != nil {
			close(pool.permissionChannel)
			pool.permissionChannel = nil
		}
	})
}
