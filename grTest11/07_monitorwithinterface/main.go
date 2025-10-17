package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TaskMonitor struct {
	stats map[string]int
	mu    sync.Mutex
	stop  chan struct{}
}

type Worker interface {
	Process(task int) int
	GetID() string
}

type SimpleWorker struct {
	ID string
}

func (w *SimpleWorker) Process(task int) int {
	time.Sleep(10 * time.Millisecond)
	return task * task
}

func (w *SimpleWorker) GetID() string {
	return w.ID
}

func NewMonitor() *TaskMonitor {
	taskMonitor := &TaskMonitor{
		stats: make(map[string]int),
		stop:  make(chan struct{}),
	}

	go taskMonitor.monitor()

	return taskMonitor
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tasks := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	workers := []Worker{
		&SimpleWorker{ID: "worker1"},
		&SimpleWorker{ID: "worker2"},
	}

	results, err := RunWorkers(ctx, tasks, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Results: %v\n", results)

}

func RunWorkers(ctx context.Context, tasks []int, workers []Worker) ([]int, error) {

	if len(workers) == 0 || len(tasks) == 0 {
		return nil, nil
	}

	var (
		wg          sync.WaitGroup
		taskChan    = make(chan int, len(tasks))
		resultChan  = make(chan int, len(tasks))
		results     []int
		taskMonitor = NewMonitor()
	)

	// собираем задачи
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(taskChan)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case taskChan <- task:
			}
		}
	}()

	// воркеры
	for _, worker := range workers {
		wg.Add(1)
		go func(w Worker) {
			defer wg.Done()
			for t := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := w.Process(t)
					taskMonitor.Record(w.GetID())
					resultChan <- result
				}
			}
		}(worker)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-resultChan:
				if !ok {
					return
				}
				results = append(results, r)
			}
		}
	}()

	wg.Wait()
	taskMonitor.Stop()

	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	return results, nil
}

func (ts *TaskMonitor) monitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	select {
	case <-ts.stop:
		return
	case <-ticker.C:
		fmt.Println("Stats:", ts.Snapshot())
	}
}

func (ts *TaskMonitor) Snapshot() map[string]int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	results := make(map[string]int)

	for i, v := range ts.stats {
		results[i] = v
	}

	return results
}

func (ts *TaskMonitor) Record(task string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.stats[task]++
}

func (ts *TaskMonitor) Stop() {
	close(ts.stop)
}
