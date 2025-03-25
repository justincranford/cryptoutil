package thread

import (
	"context"
	"fmt"
	"sync"
)

func runSendersReceivers(ctx context.Context, bufferSize int, senderCount int, receiverCount int, senderFunc func() any, receiverFunc func(value any)) func() {
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
		close(ch)
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
