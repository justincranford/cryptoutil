package keypairgen

import (
	"context"
	"fmt"
	"sync"
)

// KeyPair interface to represent any type of key pair
type KeyPair interface{}

// KeyPairPool represents a pool of key pairs
type KeyPairPool struct {
	pool    chan KeyPair
	workers int
	genFunc func() (KeyPair, error)
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewKeyPairPool creates a new key pair pool
func NewKeyPairPool(size, workers int, genFunc func() (KeyPair, error)) *KeyPairPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &KeyPairPool{
		pool:    make(chan KeyPair, size),
		workers: workers,
		genFunc: genFunc,
		ctx:     ctx,
		cancel:  cancel,
	}
	pool.start()
	return pool
}

// start initializes worker goroutines to fill the pool
func (kp *KeyPairPool) start() {
	for i := 0; i < kp.workers; i++ {
		kp.wg.Add(1)
		go kp.worker()
	}
}

// worker generates key pairs and fills the pool
func (kp *KeyPairPool) worker() {
	defer kp.wg.Done()
	for {
		select {
		case <-kp.ctx.Done(): // Stop worker if context is canceled
			return
		default:
			key, err := kp.genFunc()
			if err != nil {
				fmt.Println("Error generating key:", err)
				return
			}
			select {
			case <-kp.ctx.Done():
				return
			case kp.pool <- key:
			}
		}
	}
}

// Get retrieves the next available key pair (blocking if none are available)
func (kp *KeyPairPool) Get() KeyPair {
	return <-kp.pool
}

// Close shuts down the pool and waits for workers to finish
func (kp *KeyPairPool) Close() {
	kp.cancel()    // Signal workers to stop
	kp.wg.Wait()   // Wait for all workers to finish
	close(kp.pool) // Close the pool channel
}
