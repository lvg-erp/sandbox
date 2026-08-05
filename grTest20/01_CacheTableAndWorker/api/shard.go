package api

import (
	"hash/fnv"
	"sync"
)

// Количество сегментов. Обязательно степень двойки (2, 4, 8, 16, 32...)
const shardCount = 16

// shardMask используется для быстрого вычисления индекса.
// 16 в двоичном виде: 00010000. 16-1 = 15: 00001111.
// Побитовое И (AND) с такой маской эквивалентен делению по модулю (% 16),
// но работает на уровне процессора за 1 такт.
const shardMask = shardCount - 1

// item представляет собой запись в кэше.
// Использование int64 (Unix timestamp) вместо time.Time экономит 24 байта
// на каждую аллокацию и снижает нагрузку на Garbage Collector.
type item struct {
	value    any
	expireAt int64 // 0 означает, что ключ живет вечно
}

// shard — это изолированный сегмент хранилища со своей картой и мьютексом.
// Блокировка (Lock) одного шарда не блокирует работу остальных 15.
type shard struct {
	mu    sync.RWMutex
	items map[string]item
}

// getShardIndex вычисляет хэш ключа (FNV-1a) и применяет битовую маску,
// чтобы распределить ключи равномерно от 0 до 15.
func (c *Cache) getShardIndex(key string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(key))
	if err != nil {
		return 0
	}
	return h.Sum32() & shardMask
}
