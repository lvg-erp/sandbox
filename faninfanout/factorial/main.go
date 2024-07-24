package main

import (
	"factorial/alg"
	"fmt"
	"sync"
)

func main() {
	// массив чисел для поиска факториала
	nums := []int{5, 7, 10, 11, 12, 13, 14, 15, 16}
	var numsFactorial []int
	//канал для результатов
	results := make(chan int, len(nums))

	var wg sync.WaitGroup
	// fan-out задача распределяется между несколькими горутинами
	for _, m := range nums {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			//вычисляем факториал и отправляем результат в results
			results <- alg.Factorial(num)
		}(m)
	}
	//после завершения горутин закрываем канал и обнуляем счетчик WaitGroup
	go func() {
		wg.Wait()
		close(results)
	}()
	// fan-in перебор результатов из канала results

	for result := range results {
		numsFactorial = append(numsFactorial, result)
	}
	//sort.Slice(numsFactorial, func(i, j int) bool {
	//	return numsFactorial[i] < numsFactorial[j]
	//})

	fmt.Println(numsFactorial)

}
