package main

import (
	"fmt"
	"sync"
	"time"
)

const numWorkers = 3

func printNumbers(n int) {
	time.Sleep(time.Second)
	fmt.Println(n)
}

func main() {
	ch := make(chan int, 3)
	wg := sync.WaitGroup{}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for n := range ch {
				printNumbers(n)
			}
		}()
	}
	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	for i := range ch {
	//		printNumbers(i)
	//	}
	//}()
	for i := range 10 {
		ch <- i
	}

	close(ch)

	wg.Wait()

}
