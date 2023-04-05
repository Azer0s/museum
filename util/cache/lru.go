package cache

import "sync"

type LRU[T comparable, U comparable] struct {
	lastInserted RingBuffer[T]
	cache        map[T]U
	capacity     int
	mu           *sync.RWMutex
}

func NewLRU[T comparable, U comparable](capacity int) *LRU[T, U] {
	return &LRU[T, U]{
		lastInserted: *NewRingBuffer[T](capacity),
		cache:        make(map[T]U),
		capacity:     capacity,
		mu:           &sync.RWMutex{},
	}
}

func (lru *LRU[T, U]) Get(key T) (U, bool) {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	value, ok := lru.cache[key]
	return value, ok
}

func (lru *LRU[T, U]) Put(key T, value U) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if len(lru.cache) >= lru.capacity {
		lru.removeOldest()
	}
	lru.cache[key] = value
	lru.lastInserted.Insert(&key)
}

func (lru *LRU[T, U]) removeOldest() {
	oldestKey := lru.lastInserted.Current()
	if oldestKey == nil {
		return
	}
	delete(lru.cache, *oldestKey)
}
