package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	count int
	mu    sync.Mutex
}

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{}
}

func main() {

	var wg sync.WaitGroup
	counter := NewSafeCounter()

	numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	//numberOut := make([]int, len(numbers))

	for _, _ = range numbers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
			//numberOut = append(numberOut, increment(number))
		}()

	}

	wg.Wait()

	fmt.Println("current slice")
	fmt.Println(counter.Get())

}

func (sc *SafeCounter) Increment() {
	sc.mu.Lock()
	sc.count++
	sc.mu.Unlock()
}

func (sc *SafeCounter) Get() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.count
}
