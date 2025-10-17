package main

import (
	"fmt"
	"time"
)

func main() {
	jobs := make(chan int, 10)
	results := make(chan int, 10)
	rate := 500 * time.Millisecond

	go func() {
		rateLimitWorker(jobs, results, rate)
	}()

	go func() {
		for i := 1; i <= 5; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	for r := range results {
		fmt.Println(r)
	}

}

func rateLimitWorker(jobs <-chan int, results chan<- int, rate time.Duration) {

	ticker := time.NewTicker(rate)
	defer ticker.Stop()
	defer close(results)

	for j := range jobs {
		<-ticker.C
		results <- j * j
	}

}
