package main

import (
	"fmt"
	"sync"
)

const worckers = 100

type Counter struct {
	mu  sync.Mutex
	val int
}

func (c *Counter) increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.val++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.val
}

func NewCounter() *Counter {
	return &Counter{}
}

func main() {

	nCount := NewCounter()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := 0; y < 1000; y++ {
				nCount.increment()
			}
		}()
	}

	wg.Wait()

	fmt.Printf("Total count is %v\n: ", nCount.val)
}
