package radiochatter

import "sync"

type BroadcastChannel[T any] interface {
	Subscribe() (ch <-chan T, cancel func())
}

// NewBroadcastChannel returns a channel that can be used to subscribe to
// messages sent on the provided channel.
//
// When the provided channel is closed, the channels held by any listeners will
// also be closed. Failing to close the input channel will result in a zombie
// goroutine.
func NewBroadcastChannel[T any](ch <-chan T) BroadcastChannel[T] {
	b := &broadcast[T]{
		capacity: 1,
	}

	go func() {
		defer b.clear()

		for msg := range ch {
			b.mu.Lock()
			for _, listener := range b.listeners {
				select {
				case listener <- msg:
					// The message was sent successfully
				default:
					// Looks like the listener is lagging behind and their
					// buffer is full.
				}
			}
			b.mu.Unlock()
		}
	}()

	return b
}

type broadcast[T any] struct {
	capacity  int
	mu        sync.Mutex
	listeners []chan<- T
}

func (b *broadcast[T]) Subscribe() (<-chan T, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan T, b.capacity)
	cancel := func() { b.unregister(ch) }

	b.listeners = append(b.listeners, ch)

	return ch, cancel
}

func (b *broadcast[T]) unregister(ch chan T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, listener := range b.listeners {
		if listener == ch {
			// Do a swap-remove so we can avoid shuffling all the listeners.
			b.listeners[i] = b.listeners[len(b.listeners)-1]
			b.listeners = b.listeners[:len(b.listeners)-1]
			close(listener)
			return
		}
	}
}

func (b *broadcast[T]) clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, listener := range b.listeners {
		close(listener)
	}
	b.listeners = nil
}
