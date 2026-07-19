package ttlcache

import (
	"errors"
	"sync"
	"time"
)

type entry[V any] struct {
	value     V
	found     bool
	expiresAt time.Time
}

type Cache[V any] struct {
	mu      sync.Mutex
	entries map[string]entry[V]
	now     func() time.Time
}

func New[V any](now func() time.Time) (*Cache[V], error) {
	// TODO: Please implement
	return nil, errors.New("TODO")
}

func (cache *Cache[V]) Set(key string, value V, found bool, ttl time.Duration) error {
	// TODO: Please implement
	return errors.New("TODO")
}

func (cache *Cache[V]) Get(key string) (value V, found bool, hit bool) {
	// TODO: Please implement
	return value, false, false
}

func (cache *Cache[V]) Delete(key string) {
	// TODO: Please implement
}
