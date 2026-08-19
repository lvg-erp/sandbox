package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func safeWorker(id int, input <-chan int, output chan<- int, wg *sync.WaitGroup) {
	// wg.Go() сделает wg.Done() автоматически!

	defer func() {
		if r := recover(); r != nil {
			// Убираем debug.Stack() и выводим только понятное сообщение
			log.Printf("[Воркер %d] ⚠️ ПАНИКА: %v", id, r)
		}
	}()

	for num := range input {
		if num == 13 {
			panic(fmt.Sprintf("несчастливое число %d!", num))
		}

		time.Sleep(100 * time.Millisecond)
		result := num * num
		output <- result
	}
}

func supervisor(workersCount int, tasks <-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var isShuttingDown bool
	var restartPending sync.WaitGroup

	workerChannels := make([]chan int, workersCount)
	for i := range workerChannels {
		workerChannels[i] = make(chan int, 10)
	}

	// Функция перезапуска
	restartWorker := func(id int) {
		mu.Lock()
		defer mu.Unlock()

		if isShuttingDown {
			log.Printf("[Воркер %d] Пропускаем перезапуск (завершение)", id)
			return
		}

		log.Printf("[Воркер %d] 🔄 Перезапускаем...", id)

		newCh := make(chan int, 10)
		workerChannels[id] = newCh

		restartPending.Add(1)
		wg.Go(func() {
			defer restartPending.Done()
			safeWorker(id, newCh, out, &wg)
		})

		log.Printf("[Воркер %d] ✅ Перезапущен с новым каналом", id)
	}

	// Запускаем воркеров через wg.Go()
	for i, ch := range workerChannels {
		id := i
		input := ch
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Воркер %d] 💥 Упал с паникой: %v", id, r)
					restartWorker(id)
				}
			}()

			safeWorker(id, input, out, &wg)
		})
	}

	// Диспетчер
	wg.Go(func() {
		defer func() {
			mu.Lock()
			isShuttingDown = true
			mu.Unlock()

			restartPending.Wait()

			for _, ch := range workerChannels {
				if ch != nil {
					close(ch)
				}
			}
			log.Println("[Диспетчер] 📪 Все каналы воркеров закрыты")
		}()

		idx := 0
		for task := range tasks {
			mu.Lock()
			ch := workerChannels[idx%len(workerChannels)]
			mu.Unlock()

			if ch != nil {
				select {
				case ch <- task:
					idx++
				default:
					log.Printf("[Диспетчер] ⚠️ Канал воркера %d заполнен, задача %d пропущена", idx%len(workerChannels), task)
				}
			}
		}
	})

	// Закрываем выходной канал
	go func() {
		wg.Wait()
		close(out)
		log.Println("[Main] 🚪 Выходной канал закрыт")
	}()

	return out
}

func main() {
	fmt.Println("=== ЗАПУСК SUPERVISOR С ПЕРЕЗАПУСКОМ (wg.Go) ===\n")

	tasks := make(chan int, 100)
	results := supervisor(3, tasks)

	go func() {
		for i := 1; i <= 20; i++ {
			tasks <- i
			if i%5 == 0 {
				fmt.Printf("[Main] 📤 Отправлено %d задач\n", i)
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(tasks)
		fmt.Println("[Main] ✅ Все задачи отправлены")
	}()

	count := 0
	sum := 0

	for result := range results {
		count++
		sum += result
		fmt.Printf("[Main] 📊 Результат: %d\n", result)
	}

	fmt.Println("\n=== ИТОГОВАЯ СТАТИСТИКА ===")
	fmt.Printf("📊 Всего получено результатов: %d\n", count)
	fmt.Printf("📊 Сумма результатов: %d\n", sum)
	fmt.Println("✅ Программа завершена")
}
