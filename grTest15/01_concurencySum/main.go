package main

import (
	"fmt"
	"sync"
)

func main() {

	const stream = 3

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var wgP sync.WaitGroup
	var mu sync.Mutex

	var results int
	cSize := (len(numbers) + stream - 1) / stream

	for i := 0; i < stream; i++ {
		start := cSize * i
		end := cSize + start
		if end > len(numbers) {
			end = len(numbers)
		}
		wgP.Add(1)
		go func(chunk []int) {
			defer wgP.Done()
			locSum := 0
			for _, y := range chunk {
				locSum += y
			}
			mu.Lock()
			results += locSum
			mu.Unlock()
		}(numbers[start:end])

	}

	wgP.Wait()

	fmt.Println(results)

}
