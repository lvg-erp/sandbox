package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

type Cache struct {
	str     map[string]interface{}
	expires map[string]time.Time
	mu      sync.Mutex
	stop    chan struct{}
}

func NewCache() *Cache {
	c := &Cache{
		str:     make(map[string]interface{}),
		expires: make(map[string]time.Time),
		stop:    make(chan struct{}),
	}
	go c.cleanup()
	return c
}

func main() {
	var keyCache []string
	cacheTest := NewCache()
	for i := 0; i < 10; i++ {
		k := generateString()
		keyCache = append(keyCache, k)
		cacheTest.Set(k, i, 5*time.Second)
	}

	fmt.Println(cacheTest)

	for _, key := range keyCache {
		if val, ok := cacheTest.Get(key); ok {
			fmt.Printf("Before TTL, key %s: %v\n", key, val)
		} else {
			fmt.Printf("Before TTL, key %s: not found or expired\n", key)
		}
	}
	time.Sleep(6 * time.Second)
	for _, key := range keyCache {
		if val, ok := cacheTest.Get(key); ok {
			fmt.Printf("After TTL, key %s: %v\n", key, val)
		} else {
			fmt.Printf("After TTL, key %s: not found or expired\n", key)
		}
	}
	cacheTest.Stop()
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.str[key] = value
	c.expires[key] = time.Now().Add(ttl)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.str[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(c.expires[key]) {
		delete(c.str, key)
		delete(c.expires, key)
		return nil, false
	}
	return val, true
}
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.str, key)
	delete(c.expires, key)
}

func generateString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 4
	rand.Seed(uint64(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)

}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, exp := range c.expires {
				if time.Now().After(exp) {
					delete(c.str, key)
					delete(c.expires, key)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache) Stop() {
	close(c.stop)
}
