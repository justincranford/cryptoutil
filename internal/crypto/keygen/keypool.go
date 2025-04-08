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
	MaxKeys = int64(1<<63 - 1)
	MaxTime = time.Duration(MaxKeys)
)

type KeyPool struct {
	telemetryService  *cryptoutilTelemetry.Service
	startTime         time.Time
	name              string
	ctx               context.Context
	numWorkers        int
	size              int
	maxKeys           int64
	maxTime           time.Duration
	permissionChannel chan struct{}
	keyChannel        chan Key
	waitForWorkers    sync.WaitGroup
	generateFunction  func() (Key, error)
	closeOnce         sync.Once
	cancelFunction    context.CancelFunc
	generateCounter   int64
	getCounter        int64
}

func NewKeyPool(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, name string, numWorkers int, size int, maxKeys int64, maxTime time.Duration, generateFunction func() (Key, error)) (*KeyPool, error) {
	if numWorkers > size {
		return nil, fmt.Errorf("More workers than pool size is not allowed")
	} else if int64(size) > maxKeys {
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
		permissionChannel: make(chan struct{}, size),
		keyChannel:        make(chan Key, size),
		generateFunction:  generateFunction,
		cancelFunction:    cancelFunction,
	}
	if pool.maxKeys > 0 || pool.maxTime > 0 {
		go pool.monitorShutdown()
	}
	for i := 0; i < pool.numWorkers; i++ {
		pool.waitForWorkers.Add(1)
		go pool.generateWorker(i + 1)
	}
	return pool, nil
}

func (pool *KeyPool) monitorShutdown() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-pool.ctx.Done():
			pool.telemetryService.Slogger.Debug("cancelled", "pool", pool.name)
			return
		case <-ticker.C:
			reachedLimit := (pool.maxTime > 0 && time.Since(pool.startTime) >= pool.maxTime) || (pool.maxKeys > 0 && atomic.LoadInt64(&pool.generateCounter) > pool.maxKeys)
			if reachedLimit {
				pool.telemetryService.Slogger.Warn("limit reached", "pool", pool.name)
				return
			}
		}
	}
}

func (pool *KeyPool) generateWorker(workerNum int) {
	defer pool.waitForWorkers.Done()
	for {
		startTime := time.Now()
		pool.telemetryService.Slogger.Debug("check", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.ctx.Done():
			pool.telemetryService.Slogger.Debug("cancelled before", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.permissionChannel <- struct{}{}: // acquire permission to generate
			pool.telemetryService.Slogger.Debug("permitted", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
		}
		generateCounter := atomic.AddInt64(&pool.generateCounter, 1)
		if (pool.maxKeys > 0 && generateCounter > pool.maxKeys) || (pool.maxTime > 0 && time.Since(pool.startTime) >= pool.maxTime) {
			pool.telemetryService.Slogger.Debug("release", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Warn("limit", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		}
		key, err := pool.generateFunction()
		if err != nil {
			pool.telemetryService.Slogger.Debug("release", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Error("failed", "pool", pool.name, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds(), "error", err)
			return
		}
		pool.telemetryService.Slogger.Debug("generated", "pool", pool.name, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
		select {
		case <-pool.ctx.Done():
			pool.telemetryService.Slogger.Debug("release", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Debug("cancelled after", "pool", pool.name, "worker", workerNum, "duration", time.Since(startTime).Seconds())
			return
		case pool.keyChannel <- key:
			<-pool.permissionChannel // release permission to generate
			pool.telemetryService.Slogger.Debug("added", "pool", pool.name, "worker", workerNum, "generate", generateCounter, "duration", time.Since(startTime).Seconds())
		}
	}
}

func (pool *KeyPool) Get() Key {
	startTime := time.Now()
	pool.telemetryService.Slogger.Debug("getting", "pool", pool.name, "duration", time.Since(startTime).Seconds())
	key := <-pool.keyChannel
	pool.telemetryService.Slogger.Debug("received", "pool", pool.name, "duration", time.Since(startTime).Seconds())
	getCounter := atomic.AddInt64(&pool.getCounter, 1)
	defer func() {
		pool.telemetryService.Slogger.Debug("got", "pool", pool.name, "get", getCounter, "duration", time.Since(startTime).Seconds())
	}()
	return key
}

func (pool *KeyPool) Close() {
	pool.closeOnce.Do(func() {
		startTime := time.Now()

		if pool.cancelFunction == nil {
			defer func() {
				pool.telemetryService.Slogger.Warn("already closed", "pool", pool.name, "duration", time.Since(startTime).Seconds())
			}()
		} else {
			defer func() {
				pool.telemetryService.Slogger.Info("close ok", "pool", pool.name, "duration", time.Since(startTime).Seconds())
			}()
			pool.cancelFunction()
			pool.cancelFunction = nil
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
