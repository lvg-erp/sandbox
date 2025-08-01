package main

import (
	"fmt"
	"math/big"
	"sync"
)

const workers = 3

func main() {
	arrayInt := []int{18, 83, 41, 79, 95, 57, 57, 9, 27, 42, 19, 47, 24, 86, 75, 1, 1, 78, 97, 98, 44, 82, 14, 68, 53, 33, 95, 49, 20, 91, 85, 3, 1, 54, 51, 15, 11, 95, 39, 84, 95, 49, 16, 77, 41, 57, 15, 53, 80, 34}
	//arrayInt := []int64{1, 2, 3}

	var wg sync.WaitGroup
	chIn := make(chan string, workers)
	chOut := make(chan string)
	chunkSize := (len(arrayInt) + workers - 1) / workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if end > len(arrayInt) {
			end = len(arrayInt)
		}
		go func(chunk []int) {
			defer wg.Done()
			for _, e := range chunk {
				result := fmt.Sprintf("Факториал числа \"%d\": %d", e, factorialInt(e))
				chIn <- result
			}
		}(arrayInt[start:end])
	}

	go func() {
		for f := range chIn {
			chOut <- f
		}
		close(chOut)
	}()

	go func() {
		wg.Wait()
		close(chIn)
	}()

	for o := range chOut {
		fmt.Println(o)
	}

}

// надо использовать именно big.Int
func factorialInt(n int) *big.Int {
	result := big.NewInt(1)
	for i := 2; i <= n; i++ {
		result.Mul(result, big.NewInt(int64(i)))
	}
	return result
}
