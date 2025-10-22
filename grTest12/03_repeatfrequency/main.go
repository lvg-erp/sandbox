package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// стурктура статистки воркеров
type WorkersMap struct {
	stats map[string]int
	mu    sync.RWMutex
	flag  chan struct{}
}

func NewWorkersMap() *WorkersMap {
	nwp := &WorkersMap{
		stats: make(map[string]int),
		flag:  make(chan struct{}),
	}

	go nwp.monitor()
	return nwp
}

func (wp *WorkersMap) monitor() {

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Statistics: ", wp.snapshot())
		case <-wp.flag:
			return
		}
	}

}

func (wp *WorkersMap) snapshot() map[string]int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	resultMap := make(map[string]int)

	for i, v := range wp.stats {
		resultMap[i] = v
	}

	return resultMap
}

func (wp *WorkersMap) stop() {
	close(wp.flag)
}

func (wp *WorkersMap) update(workerID string, count int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.stats[workerID] += count

}

// структура для хранения и обработки данных обработанных воркерами
type WordCounter interface {
	CountWords(slice []string) (map[string]int, error)
	GetID() string
}

type SimpleCounter struct {
	ID string
}

func (sc *SimpleCounter) GetID() string {
	return sc.ID
}

func (sc *SimpleCounter) CountWords(slice []string) (map[string]int, error) {
	if len(slice) == 0 {
		return nil, errors.New("slice is empty")
	}

	result := make(map[string]int)
	for _, s := range slice {
		result[s]++
	}

	return result, nil

}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	input := []string{"cat", "dog", "cat", "bird", "dog", "cat"}
	workers := []WordCounter{
		&SimpleCounter{ID: "worker1"},
		&SimpleCounter{ID: "worker2"},
	}
	result, err := CountWordsConcurrently(ctx, input, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Result: %v\n", result)
}

func CountWordsConcurrently(ctx context.Context, input []string, workers []WordCounter) (map[string]int, error) {

	var (
		wg            sync.WaitGroup
		taskChan      = make(chan []string) // изменил на небуферезированный
		resultChan    = make(chan map[string]int, len(workers))
		resultMap     = make(map[string]int)
		mu            sync.Mutex
		statisticsMap = NewWorkersMap()
	)

	//делим инпут на части для воркеров
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(taskChan)
		chunkSize := (len(input) + len(workers) - 1) / len(workers)
		for i := 0; i < len(input); i += chunkSize {
			end := i + chunkSize
			if end > len(input) {
				end = len(input)
			}
			select {
			case <-ctx.Done():
				return
			case taskChan <- input[i:end]:
			}
		}
	}()

	for _, worker := range workers {
		wg.Add(1)
		go func(w WordCounter) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case s, ok := <-taskChan:
				if !ok {
					return
				}
				if freq, err := w.CountWords(s); err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					statisticsMap.update(w.GetID(), len(s))
					resultChan <- freq
				}
			}
		}(worker)
	}
	// данный код не отрабаьывает - бесконечный цикл
	// не дает запустить в работу второй воркер
	//for _, worker := range workers {
	//	wg.Add(1)
	//	go func(w WordCounter) {
	//		defer wg.Done()
	//		for {
	//			select {
	//			case <-ctx.Done():
	//				return
	//			case s, ok := <-taskChan:
	//				if !ok {
	//					return
	//				}
	//				if freq, err := w.CountWords(s); err != nil {
	//					fmt.Printf("Error: %v\n", err)
	//				} else {
	//					statisticsMap.update(w.GetID(), len(s))
	//					resultChan <- freq
	//				}
	//			}
	//		}
	//	}(worker)
	//}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(resultChan)
		for {
			select {
			case <-ctx.Done():
				return
			case f, ok := <-resultChan:
				if !ok {
					return
				}
				mu.Lock()
				for w, c := range f {
					resultMap[w] += c
				}
				mu.Unlock()
			}
		}
	}()

	wg.Wait()
	statisticsMap.stop()

	if ctx.Err() != nil {
		return resultMap, ctx.Err()
	}

	if len(resultMap) == 0 {
		return nil, errors.New("no valid results")
	}

	return resultMap, nil

}
