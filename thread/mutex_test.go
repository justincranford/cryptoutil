package thread

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
)

func TestMutex1(t *testing.T) {
	var counter int64
	var mu sync.Mutex
	for i := 0; i < 1000; i++ {
		go func() {
			mu.Lock()
			counter += rand.Int64N(100)
			mu.Unlock()
		}()
	}
	fmt.Printf("counter: %d\n", counter)
}
