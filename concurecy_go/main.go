package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {

	//создать конвейер
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	//Запустить 3 рабочии горутины
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(3)

	for w := 1; w <= 3; w++ {
		go func(id int) {
			defer wg.Done()
			processJobs(ctx, id, jobs, results)
		}(w)
	}

	//загружаем jobs и закрываем
	go func() {
		for j := 1; j <= 9; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	// запускаем сборщик
	go func() {
		wg.Wait()
		close(results)
	}()

	//Обработать результаты
	totalResult := 0
	for r := range results {
		totalResult += r
		fmt.Printf("Получил результат: %d\n", r)
	}
	fmt.Printf("Окончательный результат: %d\n", totalResult)

}

func processJobs(ctx context.Context, id int, jobs <-chan int, results chan<- int) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return // канал закрыт
			}
			fmt.Printf("Worker %d processing job %d\n", id, job)
			time.Sleep(100 * time.Millisecond) // имитация работы
			select {
			case results <- job * 2:
				// Результат отправлен
			case <-ctx.Done():
				return
			}
		}
	}
}
