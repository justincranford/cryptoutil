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
	MaxLifetimeKeys     = int64(1<<63 - 1)
	MaxLifetimeDuration = time.Duration(MaxLifetimeKeys)
)

type KeyPool struct {
	poolStartTime         time.Time
	ctx                   context.Context // Close() uses this to send Done() signal to N generateWorker threads and 1 monitorShutdown thread
	cancelWorkersFunction context.CancelFunc
	telemetryService      *cryptoutilTelemetry.Service // Observability providers (i.e. logs, metrics, traces); supports publishing to STDOUT and/or OLTP+gRPC (e.g. OpenTelemetry sidecar container http://127.0.0.1:4317/)
	poolName              string
	numWorkers            int   // TODO uint32
	poolSize              int   // TODO uint32
	maxLifetimeKeys       int64 // TODO uint64
	maxLifetimeDuration   time.Duration
	permissionChannel     chan struct{} // N generateWorker threads block wait before generating Key, because Key generation (e.g. RSA-4096) can be CPU & Memory expensive
	keyChannel            chan Key      // N generateWorker threads publish generated Keys to this channel
	waitForWorkers        sync.WaitGroup
	generateFunction      func() (Key, error)
	closeOnce             sync.Once
	generateCounter       int64 // TODO uint64
	getCounter            int64 // TODO uint64
}

// NewKeyPool supports finite or indefinite pools
func NewKeyPool(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, poolName string, numWorkers int, poolSize int, maxLifetimeKeys int64, maxLifetimeDuration time.Duration, generateFunction func() (Key, error)) (*KeyPool, error) {
	poolStartTime := time.Now() // used by N generateWorker threads and 1 monitorShutdown thread to enforce maxLifetimeDuration
	if ctx == nil {
		return nil, fmt.Errorf("Context can't be nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("Telemetry service can't be nil")
	} else if len(poolName) == 0 {
		return nil, fmt.Errorf("Name can't be empty")
	} else if numWorkers < 1 {
		return nil, fmt.Errorf("Number of workers must be at least 1")
	} else if poolSize < 1 {
		return nil, fmt.Errorf("Pool size must be at least 1")
	} else if maxLifetimeKeys < 0 {
		return nil, fmt.Errorf("Max lifetime keys must be at least 1")
	} else if maxLifetimeDuration < 0 {
		return nil, fmt.Errorf("Max lifetime duration must be positive and non-zero")
	} else if numWorkers > poolSize {
		return nil, fmt.Errorf("Number of workers must be less than or equal to pool size")
	} else if int64(poolSize) > maxLifetimeKeys {
		return nil, fmt.Errorf("Pool size must be less than or equal to max lifetime keys")
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
	if pool.maxLifetimeKeys > 0 || pool.maxLifetimeDuration > 0 {
		go pool.monitorShutdown()
	}
	for i := 0; i < pool.numWorkers; i++ {
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
			reachedLimit := (pool.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.maxLifetimeDuration) || (pool.maxLifetimeKeys > 0 && atomic.LoadInt64(&pool.generateCounter) > pool.maxLifetimeKeys)
			if reachedLimit {
				pool.telemetryService.Slogger.Warn("limit reached", "pool", pool.poolName)
				pool.Close()
				return
			}
		}
	}
}

func (pool *KeyPool) generateWorker(workerNum int) {
	defer pool.waitForWorkers.Done()
	for {
		startTime := time.Now()
		pool.telemetryService.Slogger.Debug("check", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.ctx.Done(): // someone called Close()
			pool.telemetryService.Slogger.Debug("cancelled before", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.permissionChannel <- struct{}{}: // acquire permission to generate
			pool.telemetryService.Slogger.Debug("permitted", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}
		generateCounter := atomic.AddInt64(&pool.generateCounter, 1)
		if (pool.maxLifetimeKeys > 0 && generateCounter > pool.maxLifetimeKeys) || (pool.maxLifetimeDuration > 0 && time.Since(pool.poolStartTime) >= pool.maxLifetimeDuration) {
			pool.telemetryService.Slogger.Debug("release", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Warn("limit", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		}
		key, err := pool.generateFunction()
		if err != nil {
			pool.telemetryService.Slogger.Debug("release", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Error("failed", "pool", pool.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds(), "error", err)
			return
		}
		pool.telemetryService.Slogger.Debug("generated", "pool", pool.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.ctx.Done(): // someone called Close()
			pool.telemetryService.Slogger.Debug("release", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Debug("cancelled after", "pool", pool.poolName, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.keyChannel <- key:
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Debug("added", "pool", pool.poolName, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
		}
	}
}

func (pool *KeyPool) Get() Key {
	startTime := time.Now()
	pool.telemetryService.Slogger.Debug("getting", "pool", pool.poolName, "duration", time.Since(startTime).Seconds())
	key := <-pool.keyChannel
	pool.telemetryService.Slogger.Debug("received", "pool", pool.poolName, "duration", time.Since(startTime).Seconds())
	getCounter := atomic.AddInt64(&pool.getCounter, 1)
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
