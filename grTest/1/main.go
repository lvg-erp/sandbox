package main

import (
	"fmt"
	"sync"
)

func sqrt(x int) int {
	return x * x
}

func squareNumCh(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range ch {
		fmt.Println(sqrt(i))
	}
}

func main() {
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//ch := make(chan int, 4)
	//for i := 1; i <= 4; i++ {
	//	//s := sqrt(i)
	//	//fmt.Println(s)
	//	workCh(ctx, i, ch)
	//}
	wg := &sync.WaitGroup{}
	ch := make(chan int, 4)
	workers := 2
	wg.Add(workers)
	//go func() { ------- НЕВЕРНО
	//	squareNumCh(ch, wg)
	//}()

	for i := 0; i < workers; i++ {
		go squareNumCh(ch, wg)
	}

	for i := 1; i <= 4; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()

}
