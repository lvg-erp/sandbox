package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// передача данных из одного канала в два и более каналов

func generateChan() chan int {
	ch := make(chan int)

	go func() {
		for i := range 5 {
			ch <- i
		}
		close(ch)
	}()

	return ch
}

func tee(ctx context.Context, in chan int, numChans int) []chan int {
	chans := make([]chan int, numChans)

	for i := range numChans {
		chans[i] = make(chan int)
	}

	go func() {
		for i := range numChans {
			defer close(chans[i])
		}

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				wg := &sync.WaitGroup{}
				for i := range numChans {
					wg.Add(1)
					go func() {
						defer wg.Done()
						select {
						case <-ctx.Done():
							return
						case chans[i] <- val:
						}
					}()
				}
				wg.Wait()
			}
		}
	}()

	return chans
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chans := tee(ctx, generateChan(), 3)

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		for v := range chans[0] {
			fmt.Println("writing metrics....", v)
		}
	}()

	go func() {
		defer wg.Done()
		for v := range chans[1] {
			time.Sleep(time.Second)
			fmt.Println("sending rest....", v)
		}
	}()

	go func() {
		defer wg.Done()
		for v := range chans[2] {
			fmt.Println("eventing other services....", v)
		}
	}()
	wg.Wait()
}
