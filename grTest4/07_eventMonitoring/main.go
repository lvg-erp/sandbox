package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

const workers = 3

func main() {

	wg := sync.WaitGroup{}
	ch := make(chan int, workers)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for i := 0; i < workers; i++ {
		interval := time.Duration(rand.IntN(3)) * time.Second
		if interval == 0 {
			interval = time.Second
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					select {
					case ch <- idx:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for e := range ch {
		fmt.Println(e)
	}

}
