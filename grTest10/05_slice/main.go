package main

import (
	"fmt"
	"sync"
)

const workers = 3

func main() {

	var (
		wg sync.WaitGroup
	)
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	chDivideEight := make(chan int)
	var arDivideEight []int
	chunkSize := (len(numbers) + workers - 1) / workers

	for i := 0; i < workers; i++ {

		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if end > len(numbers) {
			end = len(numbers)
		}

		go func(chank []int) {
			defer wg.Done()
			for _, c := range chank {
				if c%4 == 0 {
					chDivideEight <- c
				}
			}

		}(numbers[start:end])
	}

	go func() {
		wg.Wait()
		close(chDivideEight)
	}()

	//вывод результатов
	for e := range chDivideEight {
		arDivideEight = append(arDivideEight, e)
	}

	fmt.Println(arDivideEight)

}
