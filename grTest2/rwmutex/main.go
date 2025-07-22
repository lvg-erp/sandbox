package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	wr = 2
	r  = 5
)

func main() {

	in := make(map[string]int)

	var wg sync.WaitGroup
	var mu sync.RWMutex
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < wr; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				mu.Lock()
				key, err := randomString(6)
				if err != nil {
					log.Fatal("Error generate key")
				}
				_ = writingToMap(in, key)

				mu.Unlock()
			}
		}()
	}

	for y := 0; y < r; y++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					mu.RLock()
					for key, value := range in {
						fmt.Printf("Reading %d: key=%s, value=%d\n", idx, key, value)
					}
					mu.RUnlock()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(y)
	}

	wg.Wait()
	mu.RLock()
	fmt.Println("Final map:", in)
	mu.RUnlock()

}

func writingToMap(in map[string]int, key string) map[string]int {
	in[key]++
	return in
}

func randomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b) // Генерируем случайные байты
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	for i, v := range b {
		b[i] = charset[v%byte(len(charset))]
	}
	return string(b), nil
}
