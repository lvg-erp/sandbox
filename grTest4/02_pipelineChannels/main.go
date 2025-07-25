package main

import (
	"errors"
	"fmt"
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

func generateUniqueRandomNumbers(n, min, max int, seed int64) ([]int, error) {
	if n > max-min+1 {
		return nil, errors.New("requested number of unique values exceeds range")
	}

	numbers := make([]int, max-min+1)
	for i := range numbers {
		numbers[i] = min + i
	}

	rng := rand.New(rand.NewSource(uint64(seed)))
	for i := len(numbers) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}

	return numbers[:n], nil
}

func main() {

	n := 10
	minN := 1
	maxN := 100
	seed := time.Now().UnixNano()

	var wg sync.WaitGroup
	chN := make(chan int, n)
	chEven := make(chan int)
	chMtx := make(chan int)

	numbers, err := generateUniqueRandomNumbers(n, minN, maxN, seed)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, number := range numbers {
			chN <- number
		}
		close(chN)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := range chN {
			if n%2 == 0 {
				chEven <- n
			}
		}
		close(chEven)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range chEven {
			chMtx <- e * 2
		}
		close(chMtx)
	}()

	go func() {
		wg.Wait()
	}()

	for k := range chMtx {
		fmt.Println(k)
	}

}

//func generateRandomNumbers(n, min, max int) []int {
//
//	result := make([]int, n)
//
//	for i := 0; i < n; i++ {
//		result[i] = rand.IntN(max-min+1) + min
//	}
//
//	return result
//}

//func generateRandomNumbers(n, min, max int) chan int {
//	// Создаем слайс для хранения чисел
//	result := make(chan int, n)
//
//	// Генерируем n случайных чисел
//	for i := 0; i < n; i++ {
//		result <- rand.IntN(max-min+1) + min
//	}
//	close(result)
//	return result
//}
