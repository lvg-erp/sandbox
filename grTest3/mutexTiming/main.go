package main

import (
	"context"
	"fmt"
	"golang.org/x/exp/rand"
	"log"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	hashTable := make(map[string]int)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Println("Timeout context")
					return
				default:
					key, err := randomString(6)
					if err != nil {
						log.Printf("Error generating key: %v", err)
						continue
					}
					mu.Lock()
					_ = writingToMap(hashTable, key)
					mu.Unlock()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	for y := 0; y < 5; y++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Println("Timeout context")
					return
				default:
					mu.RLock()
					for r, v := range hashTable {
						fmt.Printf("Reading map key: %v, value: %v\n", r, v)
					}
					mu.RUnlock()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()
	mu.RLock()
	fmt.Printf("Final map: %v\n", hashTable)
	mu.RUnlock()

}

func randomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	// инициализация генератора случайных чисел
	rand.Seed(uint64(time.Now().UnixNano()))
	_, err := rand.Read(b) // Генерируем случайные байты
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	for i, v := range b {
		b[i] = charset[v%byte(len(charset))]
	}
	return string(b), nil
}

func writingToMap(in map[string]int, key string) map[string]int {
	in[key] = rand.Intn(30)
	return in
}
