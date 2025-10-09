package main

import (
	"fmt"
	"math/big"
	"sync"
)

const workers = 2

func main() {

	var (
		wg sync.WaitGroup
	)

	chIn := make(chan int, 10)
	chOut := make(chan *big.Int, 10)

	for e := 1; e <= 10; e++ {
		chIn <- e
	}

	close(chIn)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for toCh := range chIn {
				eCh := factorialInt(toCh)
				chOut <- eCh
			}

		}()
	}

	go func() {
		wg.Wait()
		close(chOut)
	}()

	for fi := range chOut {
		fmt.Println(fi)
	}

}

func factorialInt(n int) *big.Int {
	result := big.NewInt(1)
	for i := 2; i <= n; i++ {
		result.Mul(result, big.NewInt(int64(i)))
	}
	return result
}

// один из вариантов решения
//const workers = 2
//
//func main() {
//	var wg sync.WaitGroup
//	jobs := make(chan int, 10)
//	results := make(chan *big.Int, 10)
//	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//
//	// Запуск воркеров
//	for i := 0; i < workers; i++ {
//		wg.Add(1)
//		go worker(jobs, results, &wg)
//	}
//
//	// Отправка задач в канал jobs
//	go func() {
//		for _, n := range numbers {
//			jobs <- n
//		}
//		close(jobs)
//	}()
//
//	// Закрытие канала results после завершения воркеров
//	go func() {
//		wg.Wait()
//		close(results)
//	}()
//
//	// Сбор результатов
//	for result := range results {
//		fmt.Println(result)
//	}
//}
//
//func worker(jobs <-chan int, results chan<- *big.Int, wg *sync.WaitGroup) {
//	defer wg.Done()
//	for n := range jobs {
//		result := factorialInt(n)
//		results <- result
//	}
//}
//
//func factorialInt(n int) *big.Int {
//	result := big.NewInt(1)
//	for i := 2; i <= n; i++ {
//		result.Mul(result, big.NewInt(int64(i)))
//	}
//	return result
//}
