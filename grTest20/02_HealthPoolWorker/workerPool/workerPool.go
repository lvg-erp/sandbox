package workerPool

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"workerpool/pkg"

	"golang.org/x/sync/errgroup"
)

var workerCounter uint64

type Task interface {
	Process() error
}

type PanicInfo struct {
	Time  time.Time
	Err   interface{}
	Stack string
}

type WorkerPool struct {
	task       chan Task
	workerWg   sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	errorGroup *errgroup.Group

	// управление воркерами
	workerCount   int32          // текущее количество воркеров (атомарное)
	minWorkers    int32          // минимальное количество
	maxWorkers    int32          // максимальное количество
	targetWorkers int32          // целевое количество (для масштабирования)
	restartChan   chan struct{}  // канал для перезапуска воркера
	panicChan     chan PanicInfo // канал для паники
	scaleMu       sync.Mutex     // мьютекс для масштабирования
	scaleCh       chan struct{}  // канал для сигнала масштабирования

	// Метрики для автомасштабирования
	queueLength  int64 // текущая длина очереди
	totalTasks   int64 // всего задач
	pendingTasks int64 // ожидающие задачи

	shutdownChan   chan os.Signal
	isShuttingDown atomic.Bool
	shutdownOnce   sync.Once
}

func NewWorkerPool(workerCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	minWorkers := int32(workerCount / 2)
	if minWorkers < 1 {
		minWorkers = 1
	}
	maxWorkers := int32(workerCount * 4)
	if maxWorkers < int32(workerCount) {
		maxWorkers = int32(workerCount * 2)
	}

	return &WorkerPool{
		workerCount:    int32(workerCount),
		minWorkers:     minWorkers,
		maxWorkers:     maxWorkers,
		targetWorkers:  int32(workerCount),
		task:           make(chan Task, 1000), // увеличил буфер для тестов
		ctx:            ctx,
		cancel:         cancel,
		errorGroup:     eg,
		restartChan:    make(chan struct{}, maxWorkers),
		panicChan:      make(chan PanicInfo, int(maxWorkers)),
		shutdownChan:   make(chan os.Signal, 1),
		isShuttingDown: atomic.Bool{},
		shutdownOnce:   sync.Once{},
		scaleCh:        make(chan struct{}, 1),
	}
}

// ===== МЕТОДЫ УПРАВЛЕНИЯ КОЛИЧЕСТВОМ ВОРКЕРОВ =====

// SetWorkers - устанавливает целевое количество воркеров
func (wp *WorkerPool) SetWorkers(count int) error {
	if wp.isShuttingDown.Load() {
		return pkg.ErrShuttingDown
	}

	target := int32(count)
	if target < wp.minWorkers {
		target = wp.minWorkers
		log.Printf("⚠️ Adjusted workers to minimum: %d", wp.minWorkers)
	}
	if target > wp.maxWorkers {
		target = wp.maxWorkers
		log.Printf("⚠️ Adjusted workers to maximum: %d", wp.maxWorkers)
	}

	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()

	current := atomic.LoadInt32(&wp.workerCount)
	if target == current {
		return nil
	}

	atomic.StoreInt32(&wp.targetWorkers, target)

	// Отправляем сигнал на масштабирование
	select {
	case wp.scaleCh <- struct{}{}:
	default:
	}

	log.Printf("📊 Scaling workers: %d → %d", current, target)
	return nil
}

// GetWorkerCount - возвращает текущее количество воркеров
func (wp *WorkerPool) GetWorkerCount() int {
	return int(atomic.LoadInt32(&wp.workerCount))
}

// GetTargetWorkerCount - возвращает целевое количество воркеров
func (wp *WorkerPool) GetTargetWorkerCount() int {
	return int(atomic.LoadInt32(&wp.targetWorkers))
}

// GetMinWorkers - возвращает минимальное количество воркеров
func (wp *WorkerPool) GetMinWorkers() int {
	return int(wp.minWorkers)
}

// GetMaxWorkers - возвращает максимальное количество воркеров
func (wp *WorkerPool) GetMaxWorkers() int {
	return int(wp.maxWorkers)
}

// SetMinWorkers - устанавливает минимальное количество воркеров
func (wp *WorkerPool) SetMinWorkers(count int) {
	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()
	wp.minWorkers = int32(count)
}

// SetMaxWorkers - устанавливает максимальное количество воркеров
func (wp *WorkerPool) SetMaxWorkers(count int) {
	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()
	wp.maxWorkers = int32(count)
}

// ===== ВНУТРЕННИЕ МЕТОДЫ =====

// scaleWorkers - выполняет масштабирование
func (wp *WorkerPool) scaleWorkers() {
	wp.scaleMu.Lock()
	defer wp.scaleMu.Unlock()

	if wp.isShuttingDown.Load() {
		return
	}

	current := atomic.LoadInt32(&wp.workerCount)
	target := atomic.LoadInt32(&wp.targetWorkers)

	if current == target {
		return
	}

	if current < target {
		// Увеличиваем количество воркеров
		diff := target - current
		log.Printf("📈 Scaling up: adding %d workers (%d → %d)", diff, current, target)

		for i := int32(0); i < diff; i++ {
			wp.startWorker()
			atomic.AddInt32(&wp.workerCount, 1)
		}
	} else {
		// Уменьшаем количество воркеров
		diff := current - target
		log.Printf("📉 Scaling down: removing %d workers (%d → %d)", diff, current, target)

		// Отправляем сигнал на остановку лишних воркеров
		for i := int32(0); i < diff; i++ {
			wp.stopWorker()
		}
	}
}

// stopWorker - останавливает один воркер (отправляет сигнал через контекст)
func (wp *WorkerPool) stopWorker() {
	// Используем специальный канал для остановки воркеров
	// или просто уменьшаем счетчик и воркеры сами завершатся при проверке
	// Но лучше использовать механизм graceful остановки

	// Вариант: создаем временный контекст для остановки воркера
	// Но проще: воркеры сами проверяют количество и завершаются
	// Реализуем через канал stopWorkerCh
}

// ===== ОБНОВЛЕННЫЙ МОНИТОР =====

func (wp *WorkerPool) monitor() error {
	// Таймер для периодической проверки нагрузки
	scaleTicker := time.NewTicker(2 * time.Second)
	defer scaleTicker.Stop()

	// Таймер для автоматического масштабирования
	autoScaleTicker := time.NewTicker(5 * time.Second)
	defer autoScaleTicker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			log.Println("📊 Monitor stopping")
			return nil

		case <-wp.restartChan:
			if wp.isShuttingDown.Load() {
				continue
			}
			log.Println("🔄 Restarting worker...")
			wp.startWorker()
			atomic.AddInt32(&wp.workerCount, 1)

		case panicInfo := <-wp.panicChan:
			log.Printf("📊 Worker panic details: %+v", panicInfo)

		case <-wp.scaleCh:
			// Ручное масштабирование
			wp.scaleWorkers()

		case <-autoScaleTicker.C:
			// Автоматическое масштабирование по нагрузке
			wp.autoScale()

		case <-scaleTicker.C:
			// Периодическая проверка состояния
			wp.updateMetrics()
		}
	}
}

// ===== АВТОМАТИЧЕСКОЕ МАСШТАБИРОВАНИЕ =====

// autoScale - автоматическое масштабирование на основе нагрузки
func (wp *WorkerPool) autoScale() {
	if wp.isShuttingDown.Load() {
		return
	}

	// Получаем метрики
	queueLen := int64(len(wp.task))
	atomic.StoreInt64(&wp.queueLength, queueLen)

	current := atomic.LoadInt32(&wp.workerCount)

	// Вычисляем коэффициент нагрузки
	// Если очередь > 2x от количества воркеров - нужно увеличить
	// Если очередь < 0.5x от количества воркеров - можно уменьшить
	ratio := float64(queueLen) / float64(current)

	// Логируем состояние
	if queueLen > 0 {
		log.Printf("📊 Auto-scale check: queue=%d, workers=%d, ratio=%.2f",
			queueLen, current, ratio)
	}

	var target int32
	needScale := false

	// Правила масштабирования
	if ratio > 2.0 && current < wp.maxWorkers {
		// Очередь растет - добавляем воркеров
		target = current + int32(queueLen/10) + 1
		if target > wp.maxWorkers {
			target = wp.maxWorkers
		}
		needScale = true
		log.Printf("📈 Auto-scale up: queue too large (ratio=%.2f)", ratio)

	} else if ratio < 0.3 && current > wp.minWorkers && queueLen > 0 {
		// Очередь почти пуста - убираем воркеров
		target = current - 1
		if target < wp.minWorkers {
			target = wp.minWorkers
		}
		needScale = true
		log.Printf("📉 Auto-scale down: queue small (ratio=%.2f)", ratio)

	} else if queueLen == 0 && current > wp.minWorkers {
		// Совсем нет задач - постепенно уменьшаем
		target = current - 1
		if target < wp.minWorkers {
			target = wp.minWorkers
		}
		needScale = true
		log.Printf("📉 Auto-scale down: no tasks")
	}

	if needScale && target != current {
		wp.scaleMu.Lock()
		atomic.StoreInt32(&wp.targetWorkers, target)
		wp.scaleMu.Unlock()

		select {
		case wp.scaleCh <- struct{}{}:
		default:
		}
	}
}

// updateMetrics - обновляет метрики
func (wp *WorkerPool) updateMetrics() {
	queueLen := int64(len(wp.task))
	atomic.StoreInt64(&wp.queueLength, queueLen)
}

// ===== ОБНОВЛЕННЫЙ START =====

func (wp *WorkerPool) Start() {
	signal.Notify(wp.shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	wp.errorGroup.Go(wp.monitor)

	// Запускаем начальное количество воркеров
	for i := int32(0); i < wp.workerCount; i++ {
		wp.startWorker()
	}

	// Запускаем обработчик сигналов
	go wp.handleSignals()

	log.Printf("🚀 Worker pool started with %d workers (min=%d, max=%d)",
		wp.workerCount, wp.minWorkers, wp.maxWorkers)
}

// ===== ОБНОВЛЕННЫЙ STARTWORKER =====

func (wp *WorkerPool) startWorker() {
	wp.workerWg.Add(1)
	wp.errorGroup.Go(func() error {
		defer wp.workerWg.Done()
		defer wp.recoverWorker()

		workerID := generateID()
		log.Printf("🔄 Worker %d started (total: %d)", workerID, wp.GetWorkerCount())

		for {
			select {
			case <-wp.ctx.Done():
				log.Printf("🛑 Worker %d stopping (context)", workerID)
				return nil

			case task, ok := <-wp.task:
				if !ok {
					log.Printf("🛑 Worker %d stopping (channel closed)", workerID)
					return nil
				}
				wp.executeTask(task, int(workerID))
			}
		}
	})
}

// ===== ОСТАЛЬНЫЕ МЕТОДЫ =====

func (wp *WorkerPool) Shutdown() {
	wp.shutdownOnce.Do(wp.doShutdown)
}

func (wp *WorkerPool) doShutdown() {
	wp.isShuttingDown.Store(true)
	log.Printf("🛑 Initiating shutdown...")

	close(wp.task)
	log.Println("✅ Tasks channel closed")

	wp.cancel()
	log.Println("✅ Context cancelled")

	done := make(chan struct{})
	go func() {
		wp.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ All workers finished.")
	case <-time.After(30 * time.Second):
		log.Println("⚠️ Shutdown timeout - forcing exit")
	}

	if err := wp.errorGroup.Wait(); err != nil {
		log.Printf("❌ Error group with error: %v", err)
	}

	log.Println("✅ Shutdown complete")
}

func (wp *WorkerPool) AddTask(task Task) error {
	if wp.isShuttingDown.Load() {
		return pkg.ErrShuttingDown
	}

	select {
	case wp.task <- task:
		atomic.AddInt64(&wp.totalTasks, 1)
		return nil
	case <-wp.ctx.Done():
		return pkg.ErrShuttingDown
	default:
		return pkg.ErrQueueFull
	}
}

func (wp *WorkerPool) recoverWorker() {
	if r := recover(); r != nil {
		stack := debug.Stack()
		log.Printf("💥 Worker panicked: %v\nStack:\n%s", r, stack)

		select {
		case wp.restartChan <- struct{}{}:
		default:
		}

		wp.panicChan <- PanicInfo{
			Time:  time.Now(),
			Err:   r,
			Stack: string(stack),
		}
	}
}

func (wp *WorkerPool) executeTask(task Task, workerID int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("💥 Task panicked worker %d: %v", workerID, r)
		}
	}()

	if err := task.Process(); err != nil {
		log.Printf("❌ Task failed worker %d: %v", workerID, err)
	}
}

func (wp *WorkerPool) handleSignals() {
	<-wp.shutdownChan
	log.Println("📡 Received shutdown signal")
	wp.Shutdown()
}

func (wp *WorkerPool) QueueSize() int {
	return len(wp.task)
}

func generateID() int64 {
	epoch := int64(1609459200000)
	timestamp := time.Now().UnixNano()/1e6 - epoch
	counter := atomic.AddUint64(&workerCounter, 1)
	return (timestamp << 23) | int64(counter&((1<<23)-1))
}

// Tasks - возвращает канал задач (только для чтения)
func (wp *WorkerPool) Tasks() <-chan Task {
	return wp.task
}

// QueueCapacity - возвращает емкость очереди
func (wp *WorkerPool) QueueCapacity() int {
	return cap(wp.task)
}

func (wp *WorkerPool) IsShuttingDown() bool {
	return wp.isShuttingDown.Load()
}
