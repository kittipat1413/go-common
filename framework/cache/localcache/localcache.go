package localcache

import (
	"context"
	"sync"
	"time"

	cache "github.com/kittipat1413/go-common/framework/cache"
)

type item[T any] struct {
	data    T
	expires time.Time
}

type localcache[T any] struct {
	mutex sync.RWMutex
	items map[string]item[T]
}

func New[T any]() cache.Cache[T] {
	return &localcache[T]{
		items: make(map[string]item[T]),
	}
}

func (c *localcache[T]) Get(ctx context.Context, key string, initializer cache.Initializer[T]) (T, error) {
	if data, ok := c.get(key); ok {
		return data, nil
	} else {
		if initializer == nil {
			return data, cache.ErrCacheMiss
		}
		return c.initialize(key, initializer)
	}
}

func (c *localcache[T]) Set(ctx context.Context, key string, value T, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = item[T]{
		data:    value,
		expires: time.Now().Add(duration),
	}
}

func (c *localcache[T]) Invalidate(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

func (c *localcache[T]) InvalidateAll(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]item[T])
	return nil
}

func (c *localcache[T]) get(key string) (result T, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if itm, found := c.items[key]; found && time.Now().Before(itm.expires) {
		result, ok = itm.data, true
	} else {
		ok = false
	}
	return
}

func (c *localcache[T]) initialize(key string, initializer cache.Initializer[T]) (result T, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Double-check if the item was initialized by another goroutine
	if itm, found := c.items[key]; found && time.Now().Before(itm.expires) {
		return itm.data, nil
	}

	var age time.Duration
	if result, age, err = initializer(); err != nil {
		return
	}

	c.items[key] = item[T]{
		data:    result,
		expires: time.Now().Add(age),
	}
	return
}
