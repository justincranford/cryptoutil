package thread

import (
	"strconv"
	"sync"
	"testing"
)

func TestThreads(t *testing.T) {
	tests := []struct {
		count int
	}{
		{1},
		{100},
	}

	for _, test := range tests {
		t.Run(strconv.Itoa(test.count), func(t *testing.T) {
			RunThreads(t, test.count)
		})
	}
}

func RunThreads(t *testing.T, count int) {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go Worker(t, i, &wg)
	}
	wg.Wait()
}

func Worker(t *testing.T, worker int, wg *sync.WaitGroup) {
	t.Log("Worker done", worker)
	wg.Done()
}
