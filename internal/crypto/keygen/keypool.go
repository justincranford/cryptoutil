package keygen

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

const (
	MaxLifetimeKeys     = uint64(18446744073709551615)
	MaxLifetimeDuration = time.Duration(1<<63 - 1)
)

type KeyPool struct {
	poolStartTime         time.Time
	ctx                   context.Context // Close() uses this to send Done() signal to N generateWorker threads and 1 monitorShutdown thread
	cancelWorkersFunction context.CancelFunc
	telemetryService      *cryptoutilTelemetry.Service // Observability providers (i.e. logs, metrics, traces); supports publishing to STDOUT and/or OLTP+gRPC (e.g. OpenTelemetry sidecar container http://127.0.0.1:4317/)
	poolName              string
	numWorkers            uint32
	poolSize              uint32
	maxLifetimeKeys       uint64
	maxLifetimeDuration   time.Duration
	permissionChannel     chan struct{}  // N generateWorker threads block wait before generating Key, because Key generation (e.g. RSA-4096) can be CPU & Memory expensive
	keyChannel            chan Key       // N generateWorker threads publish generated Keys to this channel
	waitForWorkers        sync.WaitGroup // Close() uses this to wait for N generateWorker threads to finish before closing keyChannel and permissionChannel
	generateFunction      func() (Key, error)
	closeOnce             sync.Once
	generateCounter       uint64
	getCounter            uint64
}

// NewKeyPool supports finite or indefinite pools
func NewKeyPool(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, poolName string, numWorkers uint32, poolSize uint32, maxLifetimeKeys uint64, maxLifetimeDuration time.Duration, generateFunction func() (Key, error)) (*KeyPool, error) {
	poolStartTime := time.Now() // used by N generateWorker threads and 1 monitorShutdown thread to enforce maxLifetimeDuration
	if ctx == nil {
		return nil, fmt.Errorf("Context can't be nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("Telemetry service can't be nil")
	} else if len(poolName) == 0 {
		return nil, fmt.Errorf("Name can't be empty")
	} else if numWorkers == 0 {
		return nil, fmt.Errorf("Number of workers can't be 0")
	} else if poolSize == 0 {
		return nil, fmt.Errorf("Pool size can't be 0")
	} else if maxLifetimeKeys == 0 {
		return nil, fmt.Errorf("Max lifetime keys can't be 0")
	} else if maxLifetimeDuration <= 0 {
		return nil, fmt.Errorf("Max lifetime duration must be positive and non-zero")
	} else if numWorkers > poolSize {
		return nil, fmt.Errorf("Number of workers can't be greater than pool size")
	} else if uint64(poolSize) > maxLifetimeKeys {
		return nil, fmt.Errorf("Pool size can't be greater than max lifetime keys")
	}

	wrappedCtx, cancelFunction := context.WithCancel(ctx)
	pool := &KeyPool{
		poolStartTime:         poolStartTime,
		ctx:                   wrappedCtx,
		telemetryService:      telemetryService,
		poolName:              poolName,
		numWorkers:            numWorkers,
		poolSize:              poolSize,
		maxLifetimeKeys:       maxLifetimeKeys,
		maxLifetimeDuration:   maxLifetimeDuration,
		permissionChannel:     make(chan struct{}, poolSize),
		keyChannel:            make(chan Key, poolSize),
		generateFunction:      generateFunction,
		cancelWorkersFunction: cancelFunction,
	}
	go pool.monitorShutdown()
	for i := uint32(0); i < pool.numWorkers; i++ {
		pool.waitForWorkers.Add(1)
		go pool.generateWorker(i + 1)
	}
	return pool, nil
}

func (pool *KeyPool) monitorShutdown() {
	ticker := time.NewTicker(500 * time.Millisecond) // time keeps on ticking ticking ticking... into the future
	defer ticker.Stop()
	for {
		select {
		case <-pool.ctx.Done(): // someone else called Close()
			pool.telemetryService.Slogger.Debug("cancelled", "pool", pool.poolName)
			return
		case <-ticker.C:
			reachedLimit := (pool.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.maxLifetimeDuration) || (pool.maxLifetimeKeys > 0 && atomic.LoadUint64(&pool.generateCounter) > pool.maxLifetimeKeys)
			if reachedLimit {
				pool.telemetryService.Slogger.Warn("limit reached", "pool", pool.poolName)
				pool.Close()
				return
			}
		}
	}
}

func (pool *KeyPool) generateWorker(workerNum uint32) {
	startTime := time.Now()
	defer func() {
		if r := recover(); r != nil {
			pool.telemetryService.Slogger.Error("Worker panic recovered", "pool", pool.poolName, "worker", workerNum, "panic", r)
		}
		pool.waitForWorkers.Done()
		pool.telemetryService.Slogger.Debug("Worker done", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}()
	pool.telemetryService.Slogger.Debug("Worker started", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	for {
		pool.telemetryService.Slogger.Debug("check", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.ctx.Done(): // someone called Close()
			pool.telemetryService.Slogger.Debug("Worker canceled", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.permissionChannel <- struct{}{}: // acquire permission to generate
			pool.generateKeyAndReleasePermission(workerNum, startTime) // Use method with defer to guarantee permission release even if there is an error or panic
		}
	}
}

func (pool *KeyPool) generateKeyAndReleasePermission(workerNum uint32, startTime time.Time) {
	defer func() {
		if r := recover(); r != nil {
			pool.telemetryService.Slogger.Error("Recovered from panic", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "panic", r)
		}
		<-pool.permissionChannel
		pool.telemetryService.Slogger.Debug("Released permission", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}()
	pool.telemetryService.Slogger.Debug("Permission granted", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())

	generateCounter := atomic.AddUint64(&pool.generateCounter, 1)
	if (pool.maxLifetimeKeys > 0 && generateCounter > pool.maxLifetimeKeys) || (pool.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.maxLifetimeDuration) {
		pool.telemetryService.Slogger.Warn("Limit reached", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		return
	}

	key, err := pool.generateFunction()
	if err != nil {
		pool.telemetryService.Slogger.Error("Key generation failed", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds(), "error", err)
		return
	}
	pool.telemetryService.Slogger.Debug("Generated", "pool", pool.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())

	select {
	case <-pool.ctx.Done(): // someone called Close()
		pool.telemetryService.Slogger.Debug("Context canceled during publish", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	case pool.keyChannel <- key:
		pool.telemetryService.Slogger.Debug("Key added to channel", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
	}
}

func (pool *KeyPool) Get() Key {
	startTime := time.Now()
	pool.telemetryService.Slogger.Debug("getting", "pool", pool.poolName, "duration", time.Since(startTime).Seconds())
	key := <-pool.keyChannel
	pool.telemetryService.Slogger.Debug("received", "pool", pool.poolName, "duration", time.Since(startTime).Seconds())
	getCounter := atomic.AddUint64(&pool.getCounter, 1)
	defer func() {
		pool.telemetryService.Slogger.Debug("got", "pool", pool.poolName, "get", getCounter, "duration", time.Since(startTime).Seconds())
	}()
	return key
}

func (pool *KeyPool) Close() {
	pool.closeOnce.Do(func() {
		startTime := time.Now()

		if pool.cancelWorkersFunction == nil {
			defer func() {
				pool.telemetryService.Slogger.Warn("already closed", "pool", pool.poolName, "duration", time.Since(startTime).Seconds())
			}()
		} else {
			defer func() {
				pool.telemetryService.Slogger.Info("close ok", "pool", pool.poolName, "duration", time.Since(startTime).Seconds())
			}()
			pool.cancelWorkersFunction() // send Done() signal to N generateWorker threads and 1 monitorShutdown thread (if they are still listening to the shared context.WithCancel)
			pool.cancelWorkersFunction = nil
		}

		pool.waitForWorkers.Wait()

		if pool.keyChannel != nil {
			close(pool.keyChannel)
			pool.keyChannel = nil
		}
		if pool.permissionChannel != nil {
			close(pool.permissionChannel)
			pool.permissionChannel = nil
		}
	})
}
