package main

import (
	"fmt"
	"pingpong/utils"
	"sync"
	"time"
)

func main() {
	done := make(chan struct{})
	var (
		wg      sync.WaitGroup
		counter int64
	)
	ping := make(chan string)
	pong := make(chan string)

	// запускаем три воркера

	for i := 0; i < 3; i++ {
		wg.Go(func() {
			utils.UpperWorker(ping, pong, done, &counter)
		})
	}

	for i := 0; i < 2; i++ {
		wg.Go(func() {
			utils.LowerWorker(pong, ping, done, &counter)
		})
	}

	wg.Go(func() {
		select {
		case ping <- "Hello":
		case <-done:
		}
	})

	<-time.After(2 * time.Second)
	fmt.Println("Завершение")
	close(done)
	wg.Wait()
}
