package main

import (
	"fmt"
	"sync"
)

func main() {

	nums := []int{1, 2, 3}

	r := sqrt(nums)
	fmt.Println(r)

}

func sqrt(nums []int) []int {
	results := make([]int, 0, len(nums))
	ch := make(chan struct{ index, value int })
	var wg sync.WaitGroup
	for i, num := range nums {
		wg.Add(1)
		go func(idx, n int) {
			defer wg.Done()
			sqr := num * num
			data := struct {
				index, value int
			}{
				index: i,
				value: sqr,
			}
			ch <- data
		}(i, num)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for d := range ch {

		results = append(results, d.value)
	}

	return results
}
