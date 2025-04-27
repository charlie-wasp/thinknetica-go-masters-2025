package cache

import "sync"

type Cache[T comparable] struct {
	mu      sync.Mutex
	storage map[T]any
}

func New[T comparable]() *Cache[T] {
	return &Cache[T]{
		storage: make(map[T]any),
	}
}

func (c *Cache[T]) Get(key T) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	value := c.storage[key]
	return value
}

func (c *Cache[T]) Set(key T, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage[key] = value
}
