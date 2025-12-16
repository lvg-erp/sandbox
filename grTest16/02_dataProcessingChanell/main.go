package main

import (
	"fmt"
	"sync"
)

func main() {

	numbers := []int{2, 3, 4, 5, 6, 7}
	var wg sync.WaitGroup
	resultSqr := make(chan string, len(numbers))

	for _, n := range numbers {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			sqrt := sqrtNumbers(&num)
			resultSqr <- fmt.Sprintf("Квадратный корень из %d равен %d\n", num, sqrt)
		}(n)
	}

	go func() {
		wg.Wait()
		close(resultSqr)
	}()

	for res := range resultSqr {
		fmt.Printf(res)
	}

}

func sqrtNumbers(in *int) int {
	return (*in) * (*in)
}
