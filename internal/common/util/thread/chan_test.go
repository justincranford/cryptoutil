package thread

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sync"
	"testing"
	"time"
)

type stats struct {
	guard   sync.Mutex
	count   int64
	sum     int64
	minimum int64
	maximum int64
}

func (s *stats) record(value int64) int64 {
	s.guard.Lock()
	defer s.guard.Unlock()
	s.count++
	s.sum += value
	if value < s.minimum {
		s.minimum = value
	}
	if value > s.maximum {
		s.maximum = value
	}
	return value
}

func TestChan(t *testing.T) {
	s := &stats{minimum: int64(math.MaxInt64), maximum: int64(math.MinInt64)}
	r := &stats{minimum: int64(math.MaxInt64), maximum: int64(math.MinInt64)}
	sender := func() any {
		// Generate cryptographically secure random number 0-100 inclusive
		val, _ := rand.Int(rand.Reader, big.NewInt(101))
		return s.record(val.Int64())
	}
	receiver := func(value any) {
		r.record(value.(int64))
	}

	ctx, cancel := context.WithCancel(context.Background())
	waitAndClose := runSendersReceivers(ctx, 100, 8, 4, sender, receiver)
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	waitAndClose()

	s.guard.Lock()
	defer s.guard.Unlock()
	fmt.Printf("Senders>   Count: %d, Sum: %d, Min: %d, Max: %d, Average: %f\n", s.count, s.sum, s.minimum, s.maximum, float32(s.sum)/float32(s.count))

	r.guard.Lock()
	defer r.guard.Unlock()
	fmt.Printf("Receivers> Count: %d, Sum: %d, Min: %d, Max: %d, Average: %f\n", r.count, r.sum, r.minimum, r.maximum, float32(r.sum)/float32(r.count))
}
