package radiochatter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBroadcastChannel_SubscribeAndReceive(t *testing.T) {
	src := make(chan int)
	bcast := NewBroadcastChannel(src)

	sub1, cancel1 := bcast.Subscribe()
	defer cancel1()
	sub2, cancel2 := bcast.Subscribe()
	defer cancel2()

	go func() {
		src <- 1
		close(src)
	}()

	for i, sub := range []<-chan int{sub1, sub2} {
		select {
		case msg, ok := <-sub:
			assert.True(t, ok)
			assert.Equal(t, 1, msg)
		case <-time.After(time.Second):
			t.Errorf("Subscriber %d timed out waiting for message", i+1)
		}
	}
}

func TestBroadcastChannel_CloseSourceChannel(t *testing.T) {
	src := make(chan int)
	bcast := NewBroadcastChannel(src)

	sub, cancel := bcast.Subscribe()
	defer cancel()

	close(src)

	_, ok := <-sub
	assert.False(t, ok)
}

func TestBroadcastChannel_Unsubscribe(t *testing.T) {
	src := make(chan int)
	bcast := NewBroadcastChannel(src)

	sub, cancel := bcast.Subscribe()
	cancel()

	// This test relies on the implementation detail that unsubscribing closes the channel.
	// An alternative approach could be to send a message and ensure it's not received,
	// but that requires knowledge of the internal state or behavior.
	_, ok := <-sub
	assert.False(t, ok)
}

func TestBroadcastChannel_MultipleSubscribers(t *testing.T) {
	src := make(chan int, 10) // Buffered channel to avoid blocking
	bcast := NewBroadcastChannel(src)

	const numSubscribers = 5
	var wg sync.WaitGroup
	wg.Add(numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		sub, cancel := bcast.Subscribe()
		defer cancel()

		go func(s <-chan int) {
			defer wg.Done()
			for range s {
				// Simply consume messages to test broadcast functionality
			}
		}(sub)
	}

	for i := 0; i < 10; i++ {
		src <- i
	}
	close(src)

	wg.Wait() // Ensure all messages have been processed
}

func TestBroadcastChannel_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	src := make(chan int)
	bcast := NewBroadcastChannel(src)

	const numOps = 100
	var wg sync.WaitGroup
	wg.Add(numOps * 2) // For both subscribe and unsubscribe

	for i := 0; i < numOps; i++ {
		go func() {
			defer wg.Done()
			_, cancel := bcast.Subscribe()
			time.Sleep(10 * time.Millisecond) // Simulate some delay
			cancel()
		}()
	}

	go func() {
		for i := 0; i < numOps; i++ {
			wg.Done()
			src <- i
		}
		close(src)
	}()

	wg.Wait() // Ensure all subscribe and unsubscribe operations are complete
}
