package test

import (
	"cachetable/api"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkCacheSetParallel проверяет скорость записи под конкурентной нагрузкой.
// Мы генерируем уникальные ключи, чтобы разные горутины попадали в разные шарды.
func BenchmarkCacheSetParallel(b *testing.B) {
	// Создаем кэш с фоновой очисткой раз в секунду
	c := api.NewCache(1 * time.Second)
	defer c.Close()

	// Атомарный счетчик для генерации уникальных ключей
	var counter uint64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Увеличиваем счетчик атомарно
			id := atomic.AddUint64(&counter, 1)
			key := fmt.Sprintf("key-%d", id)

			// Пишем значение с TTL 1 минута
			c.Set(key, "some_payload_data", 1*time.Minute)
		}
	})
}

// BenchmarkCacheGetParallel проверяет скорость чтения.
// Сначала заполняем кэш, потом параллельно читаем.
func BenchmarkCacheGetParallel(b *testing.B) {
	c := api.NewCache(0) // Отключаем фоновую очистку, чтобы она не мешала бенчмарку
	defer c.Close()

	// Предзаполняем 10 000 ключей
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, "value", 5*time.Minute)
	}

	b.ResetTimer() // Останавливаем таймер, чтобы подготовка не считалась

	b.RunParallel(func(pb *testing.PB) {
		// Читаем ключи по кругу
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%10000)
			c.Get(key)
			i++
		}
	})
}

// BenchmarkCacheGetLazyExpiration проверяет скорость чтения протухших ключей.
// Это самый тяжелый сценарий, так как требует перехода от RLock к Lock и удаления.
func BenchmarkCacheGetLazyExpiration(b *testing.B) {
	c := api.NewCache(0)
	defer c.Close()

	// Создаем 1000 ключей, которые УЖЕ протухли (TTL в прошлом)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("expired-key-%d", i)
		// Используем отрицательный TTL, чтобы ключ сразу стал протухшим
		c.Set(key, "dead_data", -1*time.Second)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("expired-key-%d", i%1000)
			c.Get(key) // Должен вернуть false и удалить ключ (Lazy)
			i++
		}
	})
}
