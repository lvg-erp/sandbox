package main

import (
	"fmt"
	"sync"
)

const stream = 3

func main() {

	result := make(chan int, stream)
	var (
		wg sync.WaitGroup
	)

	numbers := []int{1, 2, 3, 4, 5, 6, 7}
	chunkSize := (len(numbers) + stream - 1) / stream

	for i := 0; i < stream; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if end > len(numbers) {
			end = len(numbers)
		}

		go func(chunk []int) {
			defer wg.Done()
			localSum := 0
			for _, e := range chunk {
				localSum += e
			}

			result <- localSum

		}(numbers[start:end])
	}

	go func() {
		wg.Wait()
		close(result)

	}()
	var endSum int
	for o := range result {
		endSum += o
	}

	fmt.Println(endSum)

}
