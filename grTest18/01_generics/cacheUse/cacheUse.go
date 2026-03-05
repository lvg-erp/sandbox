package cacheUse

import (
	"context"
	"sync"
)

type result[V any] struct {
	val V
	err error
}

type Cache[K ~string, V any] struct {
	mu       sync.RWMutex
	data     map[K]V
	inFlight map[K]chan *result[V]
	loader   func(context.Context, []K) (map[K]V, error)
}

func NewCache[K ~string, V any](loader func(ctx context.Context, keys []K) (map[K]V, error)) *Cache[K, V] {
	return &Cache[K, V]{
		data:     make(map[K]V),
		inFlight: make(map[K]chan *result[V]),
		loader:   loader,
	}
}

func (c *Cache[K, V]) GetMany(ctx context.Context, keys []K) (map[K]V, error) {
	c.mu.RLock()
	missing := make([]K, 0, len(keys))
	results := make(map[K]V, len(keys))
	waiting := make(map[K]chan *result[V])
	for _, key := range keys {
		if val, ok := c.data[key]; ok {
			results[key] = val
			continue
		}
		if ch, ok := c.inFlight[key]; ok {
			waiting[key] = ch
			continue
		}
		missing = append(missing, key)
	}
	c.mu.RUnlock()

	if len(missing) > 0 {
		c.mu.Lock()
		// Обновляем inFlight для новых ключей
		for _, key := range missing {
			if _, ok := c.inFlight[key]; !ok {
				ch := make(chan *result[V], 1)
				c.inFlight[key] = ch
				waiting[key] = ch
			} else {
				// на случай, если кто-то вставил канал между нами
				ch := c.inFlight[key]
				waiting[key] = ch
			}
		}
		c.mu.Unlock() // отключили блокировку перед запуском загрузки

		// Запускаем асинхронную загрузку
		go func() {
			loaded, err := c.loader(ctx, missing)
			c.mu.Lock()
			for _, key := range missing {
				ch := waiting[key]
				if val, ok := loaded[key]; ok && err == nil {
					ch <- &result[V]{val: val}
				} else {
					ch <- &result[V]{err: err}
				}
				close(ch)
				delete(c.inFlight, key)
				if err == nil {
					c.data[key] = loaded[key]
				}
			}
			c.mu.Unlock()
		}()
	}

	// Ожидаем для каждого ключа
	for key, ch := range waiting {
		select {
		case res := <-ch:
			if res != nil && res.err != nil {
				return nil, res.err
			}
			results[key] = res.val
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return results, nil
}

func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	res, err := c.GetMany(ctx, []K{key})
	if err != nil {
		var zero V
		return zero, err
	}

	return res[key], nil
}
