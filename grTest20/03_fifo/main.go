package main

import (
	"fmt"
	"sync"
)

func main() {
	in := make(chan int, 100)

	out := workingFunc(in)

	go func() {
		for i := 0; i < 50; i++ {
			in <- i
			if i%10 == 0 {
				fmt.Printf("[Test] Отправлено %d чисел\n", i)
			}
		}
		close(in)
		fmt.Println("[Test] Все 50 чисел отправлены!")
	}()

	// Простой расчет, соберем результат и сумму
	count := 0
	summ := 0

	for result := range out {
		count++
		summ += result
	}
	fmt.Printf("[Test] Получено %d результатов\n", count)
	fmt.Printf("[Test] Сумма квадратов: %d\n", summ)
}

func workingFunc(in <-chan int) <-chan int {

	var (
		wg   sync.WaitGroup
		once sync.Once
	)

	out := make(chan int)
	workers := make([]chan int, 3)
	for i := range workers {
		workers[i] = make(chan int, 10)
	}

	// Запускаем воркеры на выполненение
	// Но рулит всем диспетчер
	for _, ch := range workers {
		ch := ch
		wg.Go(func() {
			for num := range ch {
				out <- num * num
			}
		})
	}

	// Диспетчер
	wg.Go(func() {
		defer func() {
			for _, ch := range workers {
				close(ch)
			}
		}()

		idx := 0
		for num := range in {
			workers[idx%len(workers)] <- num
			idx++
		}
	})

	go func() {
		wg.Wait()
		once.Do(func() { close(out) })
	}()

	return out

}
