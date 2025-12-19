package main

import (
	"fmt"
	"sync"
)

func main() {

	const stream = 4
	numbers := []int{2, 4, 6, 7, 9, 10, 12, 34, 56, 78, 15, 100, 166}
	//numbers := []int{2, 4, 6, 7}
	var wg sync.WaitGroup
	resultChan := make([]chan int, stream)

	cSize := (len(numbers) + stream - 1) / stream

	for i := 0; i < stream; i++ {
		start := cSize * i
		end := start + cSize
		if end > len(numbers) {
			end = len(numbers)
		}
		cInt := make(chan int)
		resultChan[i] = cInt
		wg.Add(1)
		// использование указателей - не работает!!!!!
		//go func(ptr *int, length int, c chan int) {
		//	defer wg.Done()
		//	sum := 0
		//	for i := 0; i < length; i++ {
		//		sum += *(ptr + i) // обращение через указатель
		//	}
		//	c <- sum
		//	close(c)
		//}(&numbers[start], end-start, cInt)
		go func(c chan int, chunk []int) {
			defer wg.Done()
			sum := 0
			for _, s := range chunk {
				sum += s
			}
			c <- sum
			close(c)
		}(cInt, numbers[start:end])
	}

	go func() {
		wg.Wait()
		//close(resultChan) - массив указателей нельзя закрывать!!!!!!!!!!!
	}()
	resultSum := 0
	for _, res := range resultChan {
		for e := range res {
			resultSum += e
		}
	}

	fmt.Println(resultSum)
}
