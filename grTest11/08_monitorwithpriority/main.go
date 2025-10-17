package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID       int
	Priority int
	Value    int
}

type PriorityWorker interface {
	Process(task Task) (int, error)
	GetID() string
}

type SimpleWorker struct {
	ID string
}

type TaskStats struct {
	stats map[string]int
	mu    sync.Mutex
	stop  chan struct{}
}

type resultCollector struct {
	results []int
	mu      sync.Mutex
}

func (rc *resultCollector) add(r int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.results = append(rc.results, r)
}

func (w *SimpleWorker) Process(task Task) (int, error) {
	time.Sleep(100 * time.Millisecond)
	return task.Value * 2, nil
}

func (w *SimpleWorker) GetID() string {
	return w.ID
}

func NewTaskStats() *TaskStats {
	s := &TaskStats{
		stats: make(map[string]int),
		stop:  make(chan struct{}),
	}

	go s.monitor()

	return s
}

func (s *TaskStats) monitor() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			fmt.Println("Stats: ", s.Snapshot())
		}
	}
}

func (s *TaskStats) Record(workerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats[workerID]++
}

func (s *TaskStats) Snapshot() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	results := make(map[string]int)
	for i, v := range s.stats {
		results[i] = v
	}

	return results
}

func (s *TaskStats) Stop() {
	close(s.stop)
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputTasks := []Task{
		{ID: 1, Priority: 1, Value: 2},
		{ID: 2, Priority: 2, Value: 3},
		{ID: 3, Priority: 1, Value: 4},
		{ID: 4, Priority: 2, Value: 5},
	}

	workers := []PriorityWorker{
		&SimpleWorker{ID: "worker1"},
		&SimpleWorker{ID: "worker2"},
	}

	result, err := ProcessTasks(ctx, inputTasks, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Results: %v\n", result)

}

func ProcessTasks(ctx context.Context, tasks []Task, workers []PriorityWorker) ([]int, error) {

	var (
		wg               sync.WaitGroup
		highPriorityChan = make(chan Task, len(tasks))
		lowPriorityChan  = make(chan Task, len(tasks))
		resultChan       = make(chan int, len(tasks))
		rc               = &resultCollector{}
		//results          []int
		stats = NewTaskStats()
	)

	if len(tasks) == 0 || len(workers) == 0 {
		stats.Stop()
		return nil, nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(highPriorityChan)
		defer close(lowPriorityChan)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			default:
				if task.Priority == 1 {
					highPriorityChan <- task
				} else {
					lowPriorityChan <- task
				}
			}
		}
	}()

	for _, worker := range workers {
		wg.Add(1)
		go func(w PriorityWorker) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-highPriorityChan:
					if !ok {
						select {
						case <-ctx.Done():
							return
						case task, ok := <-lowPriorityChan:
							if !ok {
								return
							}
							processTask(w, task, stats, resultChan)
						}
					} else {
						processTask(w, task, stats, resultChan)
					}
				}
			}
		}(worker)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(resultChan)
		//rc := &resultCollector{}
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-resultChan:
				if !ok {
					return
				}
				rc.add(r)
			}
		}
	}()

	wg.Wait()
	stats.Stop()

	if ctx.Err != nil {
		return rc.results, ctx.Err()
	}

	return rc.results, nil

}

func processTask(w PriorityWorker, task Task, stat *TaskStats, resultChan chan<- int) {
	result, err := w.Process(task)
	if err != nil {
		fmt.Printf("Worker %s failed to process task %d: %v\n", w.GetID(), task.ID, err)
		return
	}
	stat.Record(w.GetID())
	resultChan <- result
}
