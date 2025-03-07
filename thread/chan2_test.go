package thread

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

func TestChan2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numCh := make(chan int, 1000)
	var wg sync.WaitGroup

	for i := 0; i < 40; i++ {
		wg.Add(1)
		go sender(ctx, numCh, &wg)
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go receiver(ctx, numCh, &wg)
	}

	time.Sleep(10 * time.Millisecond)
	cancel()

	wg.Wait()
	fmt.Println("Graceful shutdown complete")
}

func sender(ctx context.Context, ch chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ch <- rand.IntN(100): // Random number [0,99]
		}
	}
}

func receiver(ctx context.Context, ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case num := <-ch:
			fmt.Println("Received:", num)
		}
	}
}
