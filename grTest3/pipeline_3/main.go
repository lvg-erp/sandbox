package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	ch1 := make(chan int, 50)
	ch2 := make(chan int, 19)
	ch3 := make(chan int, 19)
	var result []int
	// генерировать числа 1 - 50
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch1)
		for i := 0; i < 49; i++ {
			toCh := rand.Intn(100)
			ch1 <- toCh
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch2)
		for ch := range ch1 {
			if ch%3 == 0 {
				ch2 <- ch
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch3)
		for ch := range ch2 {
			ch3 <- ch + ch
		}
	}()

	wg.Wait()

	for ch := range ch3 {
		result = append(result, ch)
	}

	fmt.Println(result)

}
