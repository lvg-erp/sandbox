package main

import (
	"fmt"
	"sync"
)

const (
	working = 4 // количество максимально работающих горутин
)

func main() {

	inAr := createArInt()
	var wg sync.WaitGroup
	var mu sync.Mutex
	ch := make(chan int, 99)
	var countOK int
	semaphore := make(chan struct{}, working)
	for _, i := range inAr {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(el int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			toCh := multiply2(i)
			ch <- toCh

			mu.Lock()
			countOK++
			mu.Unlock()
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for c := range ch {
		fmt.Println(c)
	}

}

func createArInt() []int {
	var ar []int

	for i := 0; i < 100; i++ {
		ar = append(ar, i+1)
	}

	return ar
}

func multiply2(in int) int {
	return in * 2
}
