package main

import (
	"fmt"
	"sync"
)

const stream = 3

func main() {

	var (
		//wg      sync.WaitGroup
		wgProducer sync.WaitGroup
		wgConsumer sync.WaitGroup
		numbers    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		//chanResults = make(chan int, 9)
		chanInput = make(chan int, 9)
		//flag      = make(chan struct{})
		//results int64
		results int
	)

	chunkSize := (len(numbers) + stream - 1) / stream
	for i := 0; i < stream; i++ {

		start := i * chunkSize
		end := start + chunkSize
		if end > len(numbers) {
			end = len(numbers)
		}
		wgProducer.Add(1)
		go func(chunk []int) {
			defer wgProducer.Done()
			for _, ck := range chunk {
				chanInput <- ck
			}
		}(numbers[start:end])
	}

	wgConsumer.Add(1)
	go func() {
		defer wgConsumer.Done()
		for y := range chanInput {
			//atomic.AddInt64(&results, int64(y))
			results += y
		}
	}()

	go func() {
		wgProducer.Wait()
		close(chanInput)
	}()

	wgConsumer.Wait()
	//time.Sleep(100 * time.Millisecond)
	fmt.Println(results)

}
