package cache

import "sync"

type RingBuffer[T comparable] struct {
	current int
	buffer  []*T
	mu      *sync.RWMutex
}

func NewRingBuffer[T comparable](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		current: 0,
		buffer:  make([]*T, size),
		mu:      &sync.RWMutex{},
	}
}

func (rb *RingBuffer[T]) Next() *T {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.current++
	if rb.current >= len(rb.buffer) {
		rb.current = 0
	}
	return rb.buffer[rb.current]
}

func (rb *RingBuffer[T]) Insert(value *T) {
	rb.mu.Lock()
	rb.buffer[rb.current] = value
	rb.mu.Unlock()
	rb.Next()
}

func (rb *RingBuffer[T]) InsertAll(values []*T) {
	for _, value := range values {
		rb.Insert(value)
	}
}

func (rb *RingBuffer[T]) All() []*T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	fromCurrent := rb.buffer[rb.current:]
	fromStart := rb.buffer[:rb.current]
	return append(fromCurrent, fromStart...)
}

func (rb *RingBuffer[T]) Current() *T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	return rb.buffer[rb.current]
}
