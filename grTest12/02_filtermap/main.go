package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// структура по стастике
type TasksMap struct {
	stats    map[string]int
	mu       sync.Mutex
	flagstop chan struct{}
}

func NewTasksMap() *TasksMap {
	s := &TasksMap{
		stats:    make(map[string]int),
		flagstop: make(chan struct{}),
	}

	go s.monitor()

	return s

}

func (tm *TasksMap) update(element string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.stats[element]++
}

func (tm *TasksMap) snapshot() map[string]int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	result := make(map[string]int)
	for i, v := range tm.stats {
		result[i] = v
	}

	return result
}

func (tm *TasksMap) monitor() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	select {
	case <-tm.flagstop:
		return
	case <-ticker.C:
		fmt.Println("Statistics: ", tm.snapshot())
	}
}

func (tm *TasksMap) stop() {
	close(tm.flagstop)
}

//конец статистики

// интерфейсы
type Filter interface {
	//по слайсу строк
	//FilterUnique(slice []string) ([]string, error)
	FilterUnique(s string, seen map[string]struct{}) ([]string, error)
	GetID() string
}

type SimpleFilter struct {
	ID string
}

func (sf *SimpleFilter) GetID() string {
	return sf.ID
}

//

func (sf *SimpleFilter) FilterUnique(s string, seen map[string]struct{}) ([]string, error) {
	//if len(slice) == 0 {
	//	return nil, errors.New("empty slice")
	//}

	var result []string

	if _, exists := seen[s]; !exists {
		seen[s] = struct{}{}
		result = append(result, s)
	}

	return result, nil

}

//func (sf *SimpleFilter) FilterUnique(slice []string) ([]string, error) {
//	if len(slice) == 0 {
//		return nil, errors.New("empty slice")
//	}
//
//	seen := make(map[string]struct{}) //вспомогательная структура
//	var result []string
//
//	for _, s := range slice {
//		if _, exists := seen[s]; !exists {
//			seen[s] = struct{}{}
//			result = append(result, s)
//		}
//	}
//
//	return result, nil
//
//}

//конец интерфейсы

// рабочий метод
func FilterUniqueConcurrently(ctx context.Context, input []string, workers []Filter) ([]string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		wg       sync.WaitGroup
		taskChan = make(chan string, len(workers))
		//taskChan = make(chan []string, len(workers))
		resultChan    = make(chan []string, len(workers))
		results       []string
		mapTaskString = NewTasksMap()
		seen          = make(map[string]struct{}) //вспомогательная структура
	)

	//пустим в поток
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(taskChan)
		// раберем по состовляющим массив
		for _, b := range input {
			select {
			case <-ctx.Done():
				return
			case taskChan <- b:
			}
		}

		//данный цикл разделит на слайсы массив
		//chunkSize := (len(input) + len(workers) - 1) / len(workers)
		//for i := 0; i < len(input); i += chunkSize {
		//	end := i + chunkSize
		//	if end > len(input) {
		//		end = len(input)
		//	}
		//	//go func(chunk []string) {
		//	//	taskChan <- chunk
		//	//}(input[i:end])
		//	select {
		//	case <-ctx.Done():
		//		return
		//	case taskChan <- input[i:end:end]:
		//	}
		//}
	}()

	for _, worker := range workers {
		wg.Add(1)
		go func(f Filter) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case s, ok := <-taskChan:
					if !ok {
						return
					}
					ds, err := worker.FilterUnique(s, seen)
					if err == nil {
						mapTaskString.update(worker.GetID())
						resultChan <- ds
					}
				}
			}
		}(worker)
	}

	//for _, worker := range workers {
	//	wg.Add(1)
	//	go func(f Filter) {
	//		defer wg.Done()
	//		//defer close(resultChan)
	//		for {
	//			select {
	//			case <-ctx.Done():
	//				return
	//			case chunk, ok := <-taskChan:
	//				if !ok {
	//					return
	//				}
	//				//fmt.Println(chunk)
	//				d, err := worker.FilterUnique(chunk)
	//				if err == nil {
	//					mapTaskString.update(worker.GetID())
	//					resultChan <- d
	//				} else {
	//					fmt.Printf("Worker %s failed: %v\n", worker.GetID(), err)
	//				}
	//			}
	//		}
	//	}(worker)
	//}

	//
	wg.Add(1)
	go func() {
		defer wg.Done()
		//defer close(resultChan)
		for {
			select {
			case <-ctx.Done():
				return
			case result, ok := <-resultChan:
				if !ok {
					return
				}
				results = append(results, result...)
			}
		}
	}()

	wg.Wait()
	mapTaskString.stop()

	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	fmt.Println(mapTaskString.stats)

	return results, nil

}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	input := []string{"a", "b", "a", "c", "b", "d"}
	workers := []Filter{
		&SimpleFilter{ID: "worker1"},
		&SimpleFilter{ID: "worker2"},
		&SimpleFilter{ID: "worker3"},
	}
	result, err := FilterUniqueConcurrently(ctx, input, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Result: %v\n", result)
}
