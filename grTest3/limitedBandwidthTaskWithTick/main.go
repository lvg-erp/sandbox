package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

const worker = 10

func main() {

	var wg sync.WaitGroup
	var mu sync.RWMutex
	var counter int
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool, worker)
	ch := make(chan int, 10)
	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 1; i <= worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					ch <- rand.Intn(100) * 3
					mu.Lock()
					counter++
					mu.Unlock()
				}
			}
		}()
	}
	time.Sleep(10 * time.Second)
	ticker.Stop()

	for y := 1; y <= worker; y++ {
		done <- true
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for c := range ch {
		fmt.Println(c)
	}
	fmt.Printf("Counters success %v\n: ", counter)
}
