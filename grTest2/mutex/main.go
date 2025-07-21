package main

import (
	"fmt"
	"sync"
)

func main() {

	var counter int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter += 1000
			mu.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println(counter)
}
