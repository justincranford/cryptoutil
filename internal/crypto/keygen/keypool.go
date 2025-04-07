package keygen

import (
	"context"
	"fmt"
	"sync"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

const (
	MaxKeys = 1<<63 - 1
	MaxTime = time.Duration(MaxKeys)
)

type KeyPool struct {
	telemetryService  *cryptoutilTelemetry.Service
	startTime         time.Time
	name              string
	ctx               context.Context
	numWorkers        int
	size              int
	maxKeys           int
	maxTime           time.Duration
	permissionChannel chan any
	keyChannel        chan Key
	waitGroup         sync.WaitGroup
	generateFunction  func() (Key, error)
	cancelFunction    context.CancelFunc
	guardCounters     sync.Mutex
	generateCounter   int
	getCounter        int
}

func NewKeyPool(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, name string, numWorkers int, size int, maxKeys int, maxTime time.Duration, generateFunction func() (Key, error)) (*KeyPool, error) {
	if numWorkers > size {
		return nil, fmt.Errorf("More workers than pool size is not allowed")
	} else if size > maxKeys {
		return nil, fmt.Errorf("Bigger pool size than lifetime max keys is not allowed")
	}

	wrappedCtx, cancelFunction := context.WithCancel(ctx)
	pool := &KeyPool{
		telemetryService:  telemetryService,
		startTime:         time.Now(),
		name:              name,
		ctx:               wrappedCtx,
		numWorkers:        numWorkers,
		size:              size,
		maxKeys:           maxKeys,
		maxTime:           maxTime,
		permissionChannel: make(chan any, size),
		keyChannel:        make(chan Key, size),
		generateFunction:  generateFunction,
		cancelFunction:    cancelFunction,
	}
	if pool.maxKeys > 0 || pool.maxTime > 0 {
		go pool.shutdownWorker()
	}
	for i := 0; i < pool.numWorkers; i++ {
		pool.waitGroup.Add(1)
		go pool.generateWorker(i + 1)
	}
	return pool, nil
}

func (pool *KeyPool) shutdownWorker() {
	for {
		startTime := time.Now()
		select {
		case <-pool.ctx.Done():
			pool.telemetryService.Slogger.Debug("cancelled", "pool", pool.name, "duration", time.Since(startTime).Seconds())
			return
		default:
			reachedLimit, _ := pool.checkPoolLimits(false)
			if reachedLimit {
				pool.telemetryService.Slogger.Warn("limit", "pool", pool.name, "duration", time.Since(startTime).Seconds())
				return
			}
			time.Sleep(time.Second)
		}
	}
}

func (pool *KeyPool) generateWorker(workerNum int) {
	defer pool.waitGroup.Done()
	for {
		startTime := time.Now()
		select {
		case <-pool.ctx.Done():
			pool.telemetryService.Slogger.Debug("cancelled before", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.permissionChannel <- 1: // acquire permission to generate
		}
		reachedLimit, generateCounter := pool.checkPoolLimits(true)
		if reachedLimit {
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Warn("limit", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		}
		key, err := pool.generateFunction()
		if err != nil {
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Error("failed", "pool", pool.name, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds(), "error", err)
			return
		}
		pool.telemetryService.Slogger.Debug("generated", "pool", pool.name, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.ctx.Done():
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Debug("cancelled after", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.keyChannel <- key:
			pool.telemetryService.Slogger.Debug("added", "pool", pool.name, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
		}
	}
}

func (pool *KeyPool) Get() Key {
	startTime := time.Now()
	pool.telemetryService.Slogger.Debug("getting", "pool", pool.name, "duration", time.Since(startTime).Seconds())
	key := <-pool.keyChannel
	pool.telemetryService.Slogger.Debug("received", "pool", pool.name, "duration", time.Since(startTime).Seconds())
	pool.guardCounters.Lock()
	pool.getCounter++
	getCounter := pool.getCounter
	pool.guardCounters.Unlock()
	defer func() {
		pool.telemetryService.Slogger.Debug("got", "pool", pool.name, "get", getCounter, "duration", time.Since(startTime).Seconds())
	}()
	return key
}

func (pool *KeyPool) checkPoolLimits(incrementGenerateCounter bool) (bool, int) {
	pool.guardCounters.Lock()
	defer pool.guardCounters.Unlock()
	if incrementGenerateCounter {
		pool.generateCounter = pool.generateCounter + 1
	}
	isDone := (pool.maxKeys > 0 && pool.generateCounter > pool.maxKeys) || (pool.maxTime > 0 && time.Since(pool.startTime) >= pool.maxTime)
	return isDone, pool.generateCounter
}

func (pool *KeyPool) Close() {
	startTime := time.Now()
	pool.guardCounters.Lock()
	defer pool.guardCounters.Unlock()

	if pool.cancelFunction == nil {
		defer func() {
			pool.telemetryService.Slogger.Warn("already closed", "pool", pool.name, "duration", time.Since(startTime).Seconds())
		}()
	} else {
		defer func() {
			pool.telemetryService.Slogger.Debug("close ok", "pool", pool.name, "duration", time.Since(startTime).Seconds())
		}()
		pool.cancelFunction()
		pool.cancelFunction = nil
	}

	pool.waitGroup.Wait()

	if pool.keyChannel != nil {
		close(pool.keyChannel)
		pool.keyChannel = nil
	}
}
