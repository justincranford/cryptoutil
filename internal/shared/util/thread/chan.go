// Copyright (c) 2025 Justin Cranford
//
//

// Package thread provides concurrent programming utilities including channel patterns.
package thread

import (
	"context"
	"fmt"
	"sync"
)

func runSendersReceivers(ctx context.Context, bufferSize, senderCount, receiverCount int, senderFunc func() any, receiverFunc func(value any)) func() {
	ch := make(chan any, bufferSize)

	var wg sync.WaitGroup
	for i := 0; i < senderCount; i++ {
		wg.Add(1)

		go sender(ctx, ch, &wg, senderFunc)
	}

	var receiverWg sync.WaitGroup
	for i := 0; i < receiverCount; i++ {
		receiverWg.Add(1)

		go receiver(ctx, ch, &receiverWg, receiverFunc)
	}

	return func() {
		fmt.Println("waiting")
		wg.Wait()
		fmt.Println("closing")
		close(ch) //nolint:errcheck
		fmt.Println("close complete")
	}
}

func sender(ctx context.Context, ch chan<- any, wg *sync.WaitGroup, senderFunc func() any) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case ch <- senderFunc():
		}
	}
}

func receiver(ctx context.Context, ch <-chan any, wg *sync.WaitGroup, receiverFunc func(any)) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case value := <-ch:
			receiverFunc(value)
		}
	}
}
