package cache

import "testing"

func TestLru(t *testing.T) {
	lru := NewLRU[int, string](2)
	lru.Put(1, "one")
	lru.Put(2, "two")
	lru.Put(3, "three")
	_, ok := lru.Get(1)
	if ok {
		t.Errorf("Expected key 1 to be removed from cache")
	}
	_, ok = lru.Get(2)
	if !ok {
		t.Errorf("Expected key 2 to be in cache")
	}
	_, ok = lru.Get(3)
	if !ok {
		t.Errorf("Expected key 3 to be in cache")
	}
}

func TestLruWithMoreData(t *testing.T) {
	lru := NewLRU[int, string](100)
	for i := 0; i < 100; i++ {
		lru.Put(i, "value")
	}

	for i := 0; i < 100; i++ {
		_, ok := lru.Get(i)
		if !ok {
			t.Errorf("Expected key %d to be in cache", i)
		}
	}

	for i := 100; i < 200; i++ {
		lru.Put(i, "value")
	}

	for i := 0; i < 100; i++ {
		_, ok := lru.Get(i)
		if ok {
			t.Errorf("Expected key %d to be removed from cache", i)
		}
	}

	for i := 100; i < 200; i++ {
		_, ok := lru.Get(i)
		if !ok {
			t.Errorf("Expected key %d to be in cache", i)
		}
	}
}

func TestSizeOneLru(t *testing.T) {
	lru := NewLRU[int, string](1)
	lru.Put(1, "one")
	lru.Put(2, "two")
	_, ok := lru.Get(1)
	if ok {
		t.Errorf("Expected key 1 to be removed from cache")
	}
	_, ok = lru.Get(2)
	if !ok {
		t.Errorf("Expected key 2 to be in cache")
	}
}
