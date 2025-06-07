package cache

import "sync"

type Cache[K comparable, V any] struct {
	mu      sync.Mutex
	storage map[K]V
}

func New[K comparable, V any]() *Cache[T] {
	return &Cache[T]{
		storage: make(map[T]V),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.storage[key]
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage[key] = value
}
