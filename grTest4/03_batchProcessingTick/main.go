package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	ch := make(chan int)
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				wg.Add(1)
				go func() {
					defer wg.Done()
					ch <- rand.IntN(100) * 3
					fmt.Printf("Time %v: generated %d\n", t, ch)
				}()
			}
		}
	}()

	var results []int
	wg.Add(1)
	go func() {
		defer wg.Done()
		for num := range ch {
			results = append(results, num)
		}
	}()

	time.Sleep(10 * time.Second)

	go func() {
		close(done)
		wg.Wait()
		close(ch)
	}()

	fmt.Println("Итоговые числа:", results)

}
