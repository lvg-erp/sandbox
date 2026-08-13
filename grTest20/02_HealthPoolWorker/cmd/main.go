package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
	"workerpool/commandHandler"
	"workerpool/metrics"
	"workerpool/pkg"
	"workerpool/task"
	"workerpool/web" // ← ДОБАВИТЬ ИМПОРТ
	"workerpool/workerPool"
)

// ===== ОСНОВНАЯ ФУНКЦИЯ =====

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	pkg.PrintBannerWorker()

	// ===== КОНФИГУРАЦИЯ =====
	const (
		InitialWorkers  = 3
		TaskInterval    = 100 * time.Millisecond
		MaxTasks        = 0 // 0 - бесконечно
		MetricsInterval = 5 * time.Second
		WebPort         = "8080" // ← ДОБАВИТЬ ПОРТ
	)

	// ===== ИНИЦИАЛИЗАЦИЯ =====
	metrics := metrics.NewMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем пул с динамическим масштабированием
	pool := workerPool.NewWorkerPool(InitialWorkers)
	pool.Start()
	metrics.ActiveWorkers.Store(int64(InitialWorkers))

	log.Printf("🚀 Worker pool started with %d workers", InitialWorkers)
	log.Printf("📊 Press Ctrl+C to stop")
	log.Printf("💡 Type 'workers 5' to change worker count")
	log.Printf("💡 Type 'status' to see current state")

	// ===== ЗАПУСК КОМПОНЕНТОВ =====
	var wg sync.WaitGroup

	// Командный интерфейс
	cmdHandler := commandHandler.NewCommandHandler(pool, metrics)
	cmdHandler.Start()

	// ===== ЗАПУСК ВЕБ-СЕРВЕРА =====
	// Создаем конфигурацию веб-сервера
	webConfig := web.Config{
		Port: WebPort,
	}

	// Создаем веб-сервер
	webServer := web.NewServer(pool, metrics, webConfig)

	// Запускаем веб-сервер в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := webServer.Start(); err != nil {
			log.Printf("❌ Web server error: %v", err)
		}
	}()

	log.Printf("🌐 Web interface: http://localhost:%s", WebPort)
	log.Printf("📊 API: http://localhost:%s/api/status", WebPort)

	// Producer
	wg.Add(1)
	go producer(ctx, &wg, pool, metrics, TaskInterval, MaxTasks)

	// Metrics collector
	wg.Add(1)
	go metricsCollector(ctx, &wg, metrics, MetricsInterval)

	// Goroutine monitor
	wg.Add(1)
	go goroutineMonitor(ctx, &wg)

	// ===== ОБРАБОТКА СИГНАЛОВ =====
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("📡 Received signal: %v", sig)
	case <-ctx.Done():
		log.Println("📡 Context cancelled")
	}

	// ===== GRACEFUL SHUTDOWN =====
	log.Println("\n🛑 Starting graceful shutdown...")

	// Останавливаем веб-сервер
	log.Println("🛑 Stopping web server...")
	if err := webServer.Stop(); err != nil {
		log.Printf("Error stopping web server: %v", err)
	}

	// Останавливаем командный интерфейс
	cmdHandler.Stop()

	// Сигнал всем компонентам остановиться
	cancel()

	// Ждем завершения всех горутин
	log.Println("⏳ Waiting for goroutines to finish...")

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ All goroutines finished")
	case <-time.After(10 * time.Second):
		log.Println("⚠️ Goroutines timeout")
	}

	// Выводим финальную статистику
	log.Printf("\n📊 Final Metrics:\n%s", metrics.String())

	// Ждем завершения пула
	log.Println("⏳ Waiting for worker pool...")
	time.Sleep(1 * time.Second)

	// Финальный отчет
	log.Printf("\n🎯 Final Summary:")
	log.Printf("   Total tasks: %d", metrics.TasksSubmitted.Load())
	log.Printf("   Completed:   %d", metrics.TasksCompleted.Load())
	log.Printf("   Failed:      %d", metrics.TasksFailed.Load())
	log.Printf("   Panicked:    %d", metrics.TasksPanicked.Load())
	log.Printf("   Success rate: %.1f%%", metrics.GetSuccessRate())
	log.Printf("   Workers restarted: %d", metrics.WorkersRestarted.Load())
	log.Printf("   Final goroutines: %d", runtime.NumGoroutine())

	log.Println("\n👋 Application exited successfully")
}

// ===== PRODUCER (без изменений) =====

func producer(
	ctx context.Context,
	wg *sync.WaitGroup,
	pool *workerPool.WorkerPool,
	metrics *metrics.Metrics,
	interval time.Duration,
	maxTasks int,
) {
	defer wg.Done()
	defer log.Println("📦 Producer stopped")

	log.Println("📦 Producer started")

	var taskID int
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	burstMode := false
	burstCounter := 0

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if maxTasks > 0 && taskID >= maxTasks {
				log.Printf("🎯 Reached max tasks (%d), stopping producer", maxTasks)
				return
			}

			taskID++

			if taskID%30 == 0 {
				burstMode = true
				burstCounter = 10
				log.Printf("💥 BURST MODE: generating %d tasks quickly", burstCounter)
			}

			if burstMode && burstCounter > 0 {
				for i := 0; i < 3 && burstCounter > 0; i++ {
					tasks := createTask(taskID, &taskID)
					if err := pool.AddTask(tasks); err == nil {
						metrics.TasksSubmitted.Add(1)
						burstCounter--
					}
				}
				if burstCounter <= 0 {
					burstMode = false
				}
				continue
			}

			tasks := createTask(taskID, &taskID)

			if err := pool.AddTask(tasks); err != nil {
				log.Printf("❌ Failed to add task %d: %v", taskID, err)
				continue
			}

			metrics.TasksSubmitted.Add(1)

			if taskID%10 == 0 {
				log.Printf("📦 Producer: sent %d tasks", taskID)
			}
		}
	}
}

func createTask(id int, counter *int) workerPool.Task {
	*counter++

	taskType := id % 3
	switch taskType {
	case 0:
		return &task.SimpleTask{
			ID:   id,
			Name: fmt.Sprintf("task-%d-cnt-%d", id, *counter),
		}
	case 1:
		return &task.APITask{
			ID:   id,
			URL:  fmt.Sprintf("https://api.example.com/resource/%d?seq=%d", id, *counter),
			Body: fmt.Sprintf(`{"id":%d,"seq":%d,"data":"test"}`, id, *counter),
		}
	default:
		return &task.CPUTask{
			ID:   id,
			Size: 100 + (*counter % 100),
		}
	}
}

// ===== METRICS COLLECTOR (без изменений) =====

func metricsCollector(
	ctx context.Context,
	wg *sync.WaitGroup,
	metrics *metrics.Metrics,
	interval time.Duration,
) {
	defer wg.Done()
	defer log.Println("📊 Metrics collector stopped")

	log.Println("📊 Metrics collector started")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			log.Printf("\n📊 Current Metrics:\n%s", metrics.String())
		}
	}
}

// ===== GOROUTINE MONITOR (без изменений) =====

func goroutineMonitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Println("🔍 Goroutine monitor stopped")

	log.Println("🔍 Goroutine monitor started")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			numGoroutines := runtime.NumGoroutine()
			log.Printf("🔍 Active goroutines: %d", numGoroutines)

			if numGoroutines > 100 {
				log.Printf("⚠️ Too many goroutines: %d", numGoroutines)
			}
		}
	}
}
