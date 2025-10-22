package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type MaxFinder interface {
	FindMax(slice []int) (int, error)
	GetID() string
}

type SimpleMaxFinder struct {
	ID string
}

func (sm SimpleMaxFinder) FindMax(slice []int) (int, error) {
	if len(slice) == 0 {
		return 0, errors.New("slice is empty")
	}
	maximum := slice[0]
	for _, elem := range slice {
		if elem > maximum {
			maximum = elem
		}
	}
	return maximum, nil
}

func (sm SimpleMaxFinder) GetID() string {
	return sm.ID
}

type ResultMap struct {
	stats map[string]int
	mu    sync.RWMutex
	stop  chan struct{}
}

func NewResultMap() *ResultMap {
	r := &ResultMap{
		stats: make(map[string]int),
		stop:  make(chan struct{}),
	}
	go r.monitor()
	return r
}

func (r *ResultMap) monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			fmt.Println("Statistics: ", r.Snapshot())
		}
	}
}

func (r *ResultMap) Snapshot() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]int)
	for i, v := range r.stats {
		result[i] = v
	}
	return result
}

func (r *ResultMap) Add(elem string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats[elem]++
}

func (r *ResultMap) Stop() {
	close(r.stop)
}

type resultCollector struct {
	results []int
	mu      sync.Mutex
}

func (rc *resultCollector) add(m int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.results = append(rc.results, m)
}

func FindMaxConcurrently(ctx context.Context, slice []int, finders []MaxFinder) (int, error) {
	if len(slice) == 0 || len(finders) == 0 {
		return 0, errors.New("slice or finders empty")
	}

	var (
		wg         sync.WaitGroup
		taskChan   = make(chan []int, len(finders))
		resultChan = make(chan int, len(finders))
		rc         = &resultCollector{}
		stats      = NewResultMap()
	)

	// Отправка подмножеств
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(taskChan)
		chunkSize := (len(slice) + len(finders) - 1) / len(finders)
		for i := 0; i < len(slice); i += chunkSize {
			end := i + chunkSize
			if end > len(slice) {
				end = len(slice)
			}
			select {
			case <-ctx.Done():
				return
			case taskChan <- slice[i:end]:
			}
		}
	}()

	// Воркеры
	for _, finder := range finders {
		wg.Add(1)
		go func(finder MaxFinder) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case chunk, ok := <-taskChan:
					if !ok {
						return
					}
					if m, err := finder.FindMax(chunk); err == nil {
						stats.Add(finder.GetID())
						resultChan <- m
					} else {
						fmt.Printf("Worker %s failed: %v\n", finder.GetID(), err)
					}
				}
			}
		}(finder)
	}

	// Закрытие resultChan после воркеров
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(resultChan)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-resultChan:
				if !ok {
					return
				}
				rc.add(m)
			}
		}
	}()

	wg.Wait()
	stats.Stop()

	// Найти максимум среди результатов
	max := math.MinInt64
	for _, m := range rc.results {
		if m > max {
			max = m
		}
	}

	if ctx.Err() != nil {
		return max, ctx.Err()
	}
	if max == math.MinInt64 {
		return 0, errors.New("no valid results")
	}
	return max, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	workerFind := []MaxFinder{
		&SimpleMaxFinder{ID: "finder_01"},
		&SimpleMaxFinder{ID: "finder_02"},
	}

	res, err := FindMaxConcurrently(ctx, numbers, workerFind)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Max: %d\n", res)
}

//type MaxFinder interface {
//	FindMax(slice []int) (int, error)
//	GetID() string
//}
//
//// max
//type SimpleMaxFinder struct {
//	ID string
//}
//
//func (sm SimpleMaxFinder) FindMax(slice []int) (int, error) {
//	if len(slice) == 0 {
//		return 0, errors.New("slice is empty")
//	}
//
//	maximum := slice[0]
//	for _, elem := range slice {
//		if elem > maximum {
//			maximum = elem
//		}
//	}
//
//	return maximum, nil
//}
//
//func (sm SimpleMaxFinder) GetID() string {
//	return sm.ID
//}
//
//type ResultMap struct {
//	stats map[string]int
//	mu    sync.RWMutex
//	stop  chan struct{}
//}
//
//func NewResultMap() *ResultMap {
//
//	r := &ResultMap{
//		stats: make(map[string]int),
//		stop:  make(chan struct{}),
//	}
//
//	go r.monitor()
//
//	return r
//}
//
//func (r *ResultMap) monitor() {
//
//	ticker := time.NewTicker(1 * time.Second)
//	defer ticker.Stop()
//	for {
//		select {
//		case <-r.stop:
//			return
//		case <-ticker.C:
//			fmt.Println("Statistics: ", r.Snapshot())
//		}
//	}
//
//	//ticker := time.NewTicker(1 * time.Second)
//	//defer ticker.Stop()
//	//for {
//	//	select {
//	//	case <-r.stop:
//	//		return
//	//	default:
//	//		fmt.Println("Statistics: ", r.Snapshot())
//	//	}
//	//}
//
//}
//
//func (r *ResultMap) Snapshot() map[string]int {
//	r.mu.RLock()
//	defer r.mu.RUnlock()
//	result := make(map[string]int)
//	for i, v := range r.stats {
//		result[i] = v
//	}
//
//	return result
//
//}
//
//func (r *ResultMap) Add(elem string) {
//	r.mu.Lock()
//	defer r.mu.Unlock()
//	r.stats[elem]++
//}
//
//func (r *ResultMap) Stop() {
//	close(r.stop)
//}
//
//func RunFinderMax(ctx context.Context, slice []int, finders []MaxFinder) (int, error) {
//	if len(slice) == 0 || len(finders) == 0 {
//		return 0, errors.New("slice or finders empty")
//	}
//
//	var (
//		wg         sync.WaitGroup
//		taskChan   = make(chan []int, len(finders))
//		resultChan = make(chan int, len(finders))
//		rc         = &resultCollector{max: math.MinInt64}
//		stats      = NewResultMap()
//	)
//
//	// Отправка подмножеств
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		defer close(taskChan)
//		chunkSize := (len(slice) + len(finders) - 1) / len(finders)
//		for i := 0; i < len(slice); i += chunkSize {
//			end := i + chunkSize
//			if end > len(slice) {
//				end = len(slice)
//			}
//			select {
//			case <-ctx.Done():
//				return
//			case taskChan <- slice[i:end]:
//			}
//		}
//	}()
//
//	// Воркеры
//	for _, finder := range finders {
//		wg.Add(1)
//		go func(finder MaxFinder) {
//			defer wg.Done()
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				case chunk, ok := <-taskChan:
//					if !ok {
//						return
//					}
//					if m, err := finder.FindMax(chunk); err == nil {
//						stats.Add(finder.GetID())
//						resultChan <- m
//					} else {
//						fmt.Printf("Worker %s failed: %v\n", finder.GetID(), err)
//					}
//				}
//			}
//		}(finder)
//	}
//
//	// Сбор результатов
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		defer close(resultChan)
//		for m := range resultChan {
//			rc.update(m)
//		}
//	}()
//
//	wg.Wait()
//	stats.Stop()
//
//	if ctx.Err() != nil {
//		return rc.max, ctx.Err()
//	}
//	if rc.max == math.MinInt64 {
//		return 0, errors.New("no valid results")
//	}
//	return rc.max, nil
//}
//
//type resultCollector struct {
//	max int
//	mu  sync.Mutex
//}
//
//func (rc *resultCollector) update(m int) {
//	rc.mu.Lock()
//	defer rc.mu.Unlock()
//	if m > rc.max {
//		rc.max = m
//	}
//}
//
//func main() {
//
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	defer cancel()
//	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
//	workerFind := []MaxFinder{
//		&SimpleMaxFinder{ID: "finder_01"},
//		&SimpleMaxFinder{ID: "finder_02"},
//	}
//
//	res, err := RunFinderMax(ctx, numbers, workerFind)
//	if err != nil {
//		fmt.Println("Error")
//	}
//
//	fmt.Println(res)
//}

//func RunFinderMax(ctx context.Context, slice []int, finders []MaxFinder) (int, error) {
//	if len(slice) == 0 || len(finders) == 0 {
//		return 0, errors.New("slice to search is empty")
//	}
//	var (
//		wg         sync.WaitGroup
//		taskChan   = make(chan []int, len(finders))
//		resultChan = make(chan int, len(finders))
//		result     int
//		statsMap   = NewResultMap()
//	)
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		defer close(taskChan)
//		chunkSize := (len(slice) + len(finders) - 1) / len(finders)
//		for i := 0; i < len(slice); i += chunkSize {
//			end := i + chunkSize
//			if end > len(slice) {
//				end = len(slice)
//			}
//			select {
//			case <-ctx.Done():
//				return
//			case taskChan <- slice[i:end]:
//			}
//		}
//	}()
//
//	for _, f := range finders {
//		wg.Add(1)
//		go func(mf MaxFinder) {
//			defer wg.Done()
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				case chunk, ok := <-taskChan:
//					if !ok {
//						return
//					}
//					if m, err := f.FindMax(chunk); err == nil {
//						statsMap.Add(f.GetID())
//						resultChan <- m
//					} else {
//						fmt.Printf("Worker %s failed: %v\n", f.GetID(), err)
//					}
//				}
//			}
//		}(f)
//	}
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		defer close(resultChan)
//		for rc := range resultChan {
//			if rc > result {
//				result = rc
//			}
//		}
//	}()
//
//	wg.Wait()
//	statsMap.Stop()
//
//	if ctx.Err() != nil {
//		return result, ctx.Err()
//	}
//
//	return result, nil
//
//}

////filter
//
//type Filter interface {
//	FilterUnique(sliceFilter []string) ([]string, error)
//	GetID() string
//}
//
//type SimpleFilter struct {
//	ID string
//}
//
//func (sf *SimpleFilter) FilterUnique(sliceFilter []string) ([]string, error) {
//
//	if len(sliceFilter) == 0 {
//		return nil, errors.New("slice filter is empty")
//	}
//	seen := make(map[string]struct{})
//
//	var result []string
//
//	for _, elem := range sliceFilter {
//		if _, ok := seen[elem]; !ok {
//			seen[elem] = struct{}{}
//		}
//	}
//
//	return result, nil
//
//}
//
//func (sf *SimpleFilter) GetID() string {
//	return sf.ID
//}
