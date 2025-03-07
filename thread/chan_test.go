package thread

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

func TestChan(t *testing.T) {
	dataChannel := make(chan int32)
	producerDoneChannel1 := make(chan struct{})
	producerDoneChannel2 := make(chan struct{})
	consumerDoneChannel := make(chan struct{})

	go Producer(dataChannel, producerDoneChannel1) // Producer listens for stop signal
	go Producer(dataChannel, producerDoneChannel2) // Producer listens for stop signal
	go Consumer(dataChannel, consumerDoneChannel)  // Consumer processes data

	time.Sleep(100 * time.Millisecond) // Let them run for a short time

	close(producerDoneChannel1) // Signal sender to stop
	close(producerDoneChannel2) // Signal sender to stop
	close(dataChannel)          // Close dataChannel after sender exits

	<-consumerDoneChannel // Wait for consumer to finish
}

func Producer(dataChannel chan int32, producerDoneChannel <-chan struct{}) {
	for {
		select {
		case <-producerDoneChannel:
			return // Stop producing when signaled
		case dataChannel <- rand.Int32N(1000):
			time.Sleep(time.Millisecond) // Simulate work
		}
	}
}

func Consumer(dataChannel <-chan int32, consumerDoneChannel chan struct{}) {
	var count int32
	var total int64
	for value := range dataChannel {
		count++
		total += int64(value)
	}
	fmt.Printf("Count: %d, Total: %d\n", count, total)

	close(consumerDoneChannel) // Signal completion
}
