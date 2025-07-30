package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const workers = 4

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	var sum int
	var wg sync.WaitGroup
	var mu sync.Mutex
	chStorage := make(chan int, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			interval := r.Intn(workers) + 1
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()
			//time.Sleep(500 * time.Millisecond)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					select {
					case <-ctx.Done():
						return
					case chStorage <- rand.Intn(100) + 1:
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(chStorage)
	}()

	done := make(chan struct{})
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for num := range chStorage {
			ctxProc, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			doneProcessing := make(chan struct{})
			go func() {
				time.Sleep(time.Duration(r.Intn(600)) * time.Millisecond)
				select {
				case doneProcessing <- struct{}{}:
				default:
				}
			}()
			select {
			case <-ctxProc.Done():
				cancel()
				continue
			case <-doneProcessing:
				mu.Lock()
				sum += num
				mu.Unlock()
			}
			cancel()
		}
		mu.Lock()
		fmt.Printf("Final sum: %d\n", sum)
		fmt.Println("Program finished")
		mu.Unlock()
		close(done)
	}()

	tickerSum := time.NewTicker(5 * time.Second)
	defer tickerSum.Stop()

	for {
		select {
		case <-done:
			tickerSum.Stop()
			return
		case <-tickerSum.C:
			mu.Lock()
			fmt.Printf("Sum subtotal at %s: %d\n", time.Now().Format(time.RFC3339), sum)
			mu.Unlock()
		}
	}

}
