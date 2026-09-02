package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== ЗАПУСК ПРОГРАММЫ ===\n")

	tasks := make(chan int, 100)
	results := supervisor(3, tasks)

	go func() {
		for i := 0; i < 50; i++ { // ← увеличил до 50 задач
			if i%5 == 0 {
				fmt.Printf("[Main] Отправлено %d задач\n", i)
			}
			tasks <- i
			time.Sleep(50 * time.Millisecond)
		}
		close(tasks)
		fmt.Println("[Main] Все задачи отправлены")
	}()

	count := 0
	sum := 0

	for res := range results {
		count++
		sum += res
		fmt.Printf("[Main] Результат: %d\n", res)
	}

	fmt.Println("\n=== ИТОГОВАЯ СТАТИСТИКА ===")
	fmt.Printf("Всего получено результатов: %d\n", count)
	fmt.Printf("Сумма результатов: %d\n", sum)
	fmt.Println("✅ Программа завершена")
}

func safeWorker(id int, in <-chan int, out chan<- int, restartCount *int, taskCounter *int) {
	defer func() {
		if r := recover(); r != nil {
			*restartCount++
			log.Printf("[Воркер %d] 💥 Паника #%d: %v", id, *restartCount, r)
		}
	}()

	for number := range in {
		*taskCounter++

		// Падаем на КАЖДОЙ задаче для теста!
		// ИЛИ используй: if *taskCounter%2 == 0
		if *taskCounter%2 == 0 {
			panic(fmt.Sprintf("воркер %d падает на задаче #%d!", id, *taskCounter))
		}

		time.Sleep(100 * time.Millisecond)
		result := number * number
		out <- result
	}
}

func supervisor(initialWorkers int, tasks <-chan int) <-chan int {
	out := make(chan int)

	var (
		wg, restart     sync.WaitGroup
		mu              sync.Mutex
		isShuttingDown  bool
		restartCounters = make([]int, initialWorkers)
		taskCounters    = make([]int, initialWorkers) // ← счетчик задач для каждого воркера
	)

	workChannels := make([]chan int, initialWorkers)
	for i := range workChannels {
		workChannels[i] = make(chan int, 10)
	}

	restartWorker := func(id int) {
		mu.Lock()
		defer mu.Unlock()

		if isShuttingDown {
			log.Printf("[Воркер %d] ⏹️ Пропускаем перезапуск (завершение)", id)
			return
		}

		if restartCounters[id] >= 3 {
			log.Printf("[Воркер %d] 🗑️ УДАЛЕН из пула (3 падения)", id)
			if workChannels[id] != nil {
				close(workChannels[id])
				workChannels[id] = nil
			}
			return
		}

		log.Printf("[Воркер %d] 🔄 Перезапуск #%d", id, restartCounters[id]+1)

		newCh := make(chan int, 10)
		workChannels[id] = newCh

		restart.Add(1)
		wg.Go(func() {
			defer restart.Done()
			safeWorker(id, newCh, out, &restartCounters[id], &taskCounters[id])
		})

		log.Printf("[Воркер %d] ✅ Перезапущен", id)
	}

	// Запускаем воркеров
	for i := 0; i < initialWorkers; i++ {
		id := i
		ch := workChannels[i]

		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Воркер %d] 💥 Завершился паникой: %v", id, r)

					if restartCounters[id] >= 3 {
						log.Printf("[Воркер %d] 🛑 УДАЛЕН (>=3 падений)", id)
						mu.Lock()
						if workChannels[id] != nil {
							close(workChannels[id])
							workChannels[id] = nil
						}
						mu.Unlock()
						return
					}

					restartWorker(id)
				}
			}()

			safeWorker(id, ch, out, &restartCounters[id], &taskCounters[id])
		})
	}

	// Диспетчер
	wg.Go(func() {
		defer func() {
			mu.Lock()
			isShuttingDown = true
			mu.Unlock()

			restart.Wait()

			for _, ch := range workChannels {
				if ch != nil {
					close(ch)
				}
			}
			log.Println("[Диспетчер] Все каналы закрыты")
		}()

		idx := 0
		for task := range tasks {
			mu.Lock()

			// Ищем живой канал
			found := false
			for i := 0; i < len(workChannels); i++ {
				workerIdx := (idx + i) % len(workChannels)
				if workChannels[workerIdx] != nil {
					ch := workChannels[workerIdx]
					mu.Unlock()

					select {
					case ch <- task:
						idx = workerIdx + 1
						found = true
					default:
						log.Printf("[Диспетчер] ⚠️ Воркер %d занят", workerIdx)
						idx = workerIdx + 1
						found = true
					}
					break
				}
			}

			if !found {
				mu.Unlock()
				log.Printf("[Диспетчер] ❌ Нет активных воркеров! Задача %d потеряна", task)
			}
		}
	})

	go func() {
		wg.Wait()
		close(out)
		log.Println("[Main] Выходной канал закрыт")
	}()

	return out
}
