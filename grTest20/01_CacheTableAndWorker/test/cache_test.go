package test

import (
	"cachetable/api"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func BenchmarkCacheSetParallel_(b *testing.B) {
	c := api.NewCache(10 * time.Second)
	defer c.Close()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)
			c.Set(key, "value", 1*time.Minute)
			i++
		}
	})
}

func TestCacheLazyExpiration(t *testing.T) {
	c := api.NewCache(0) // Отключаем фоновую очистку
	defer c.Close()

	c.Set("mykey", "data", 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond) // Ждем пока протухнет

	val, ok := c.Get("mykey")
	if ok || val != nil {
		t.Error("Lazy expiration failed: key should be deleted on Get")
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Allocated memory: %d KB\n", m.Alloc/1024)
}
