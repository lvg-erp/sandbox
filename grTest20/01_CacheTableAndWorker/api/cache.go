package api

import (
	"context"
	"time"
)

// Cache — потокобезопасный in-memory кэш с поддержкой TTL и шардированием.
type Cache struct {
	shards [shardCount]shard

	// Механизм для graceful shutdown фонового воркера
	cancel context.CancelFunc
	done   chan struct{} // Закрывается, когда горутина очистки действительно завершилась
}

// NewCache создает новый экземпляр кэша.
// cleanupInterval — интервал запуска фонового удаления протухших ключей.
// Если передать 0, фоновая очистка будет отключена (будет работать только Lazy Expiration при Get).
func NewCache(cleanupInterval time.Duration) *Cache {
	c := &Cache{}

	// Инициализируем мапы внутри каждого шарда
	for i := 0; i < shardCount; i++ {
		c.shards[i].items = make(map[string]item)
	}

	// Запускаем фоновую очистку, если интервал > 0
	if cleanupInterval > 0 {
		var ctx context.Context
		ctx, c.cancel = context.WithCancel(context.Background())
		c.done = make(chan struct{})
		go c.cleanupLoop(ctx, cleanupInterval)
	}

	return c
}

// Set добавляет или обновляет значение по ключу с указанным TTL.
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	idx := c.getShardIndex(key)
	s := &c.shards[idx]

	var expireAt int64
	if ttl > 0 {
		expireAt = time.Now().Add(ttl).Unix()
	}

	s.mu.Lock()
	s.items[key] = item{
		value:    value,
		expireAt: expireAt,
	}
	s.mu.Unlock()
}

// Get возвращает значение по ключу.
// Если ключ отсутствует или его TTL истек, возвращается nil, false.
// При обнаружении протухшего ключа происходит его удаление (Lazy Expiration).
func (c *Cache) Get(key string) (any, bool) {
	idx := c.getShardIndex(key)
	s := &c.shards[idx]

	// 1. Быстрый путь: читаем под RLock (разрешает параллельное чтение)
	s.mu.RLock()
	it, ok := s.items[key]
	s.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// 2. Проверяем, протух ли ключ
	now := time.Now().Unix()
	if it.expireAt > 0 && now > it.expireAt {
		// Ключ протух. Нам нужно его удалить, для этого нужен Write Lock.
		s.mu.Lock()
		// 3. DOUBLE-CHECK: пока мы ждали Write Lock, другая горутина
		// могла уже удалить этот ключ или перезаписать его с новым TTL.
		if it, ok := s.items[key]; ok && it.expireAt > 0 && now > it.expireAt {
			delete(s.items, key)
		}
		s.mu.Unlock()
		return nil, false
	}

	return it.value, true
}

// Delete удаляет ключ из кэша вручную.
func (c *Cache) Delete(key string) {
	idx := c.getShardIndex(key)
	s := &c.shards[idx]

	s.mu.Lock()
	delete(s.items, key)
	s.mu.Unlock()
}

// cleanupLoop фоновая горутина, которая периодически проходит по всем шардам
// и удаляет протухшие ключи (Active Expiration), чтобы избежать утечек памяти
// от ключей, к которым больше не обращаются.
func (c *Cache) cleanupLoop(ctx context.Context, interval time.Duration) {
	defer close(c.done) // Сигнализируем в Close(), что мы вышли

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().Unix()
			// Проходим по всем шардам последовательно
			for i := 0; i < shardCount; i++ {
				s := &c.shards[i]
				s.mu.Lock()
				for k, it := range s.items {
					if it.expireAt > 0 && now > it.expireAt {
						delete(s.items, k)
					}
				}
				s.mu.Unlock()
			}
		case <-ctx.Done():
			// Получили сигнал остановки из Close()
			return
		}
	}
}

// Close корректно останавливает фоновую горутину очистки.
// Дожидается её полного завершения перед выходом.
func (c *Cache) Close() {
	if c.cancel != nil {
		c.cancel() // Отправляем сигнал остановки в cleanupLoop
		<-c.done   // Ждем, пока горутина выйдет (закроет канал)
	}
}
