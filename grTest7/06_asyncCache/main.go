package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

const workers = 3

type Result struct {
	Input  int
	Output int
}

type Error struct {
	Input int
	Error error
}

type cacheResponse struct {
	input  int
	output int
	err    error
}

// представляет запрос на вычисление
type cacheRequest struct {
	number int
	result chan cacheResponse
	ctx    context.Context
}

// асинхронный кэш
type AsyncCache struct {
	cache   map[int]int
	mu      sync.RWMutex
	request chan cacheRequest
	wg      sync.WaitGroup
}

func NewAsyncCache(workers int) *AsyncCache {
	cache := &AsyncCache{
		cache:   make(map[int]int),
		request: make(chan cacheRequest, 100),
	}

	cache.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go cache.worker()
	}
	return cache
}

func (c *AsyncCache) worker() {
	defer c.wg.Done()
	for req := range c.request {
		select {
		case <-req.ctx.Done():
			req.result <- cacheResponse{input: req.number, err: req.ctx.Err()}
			continue
		default:
			c.mu.RLock()
			if value, exists := c.cache[req.number]; exists {
				c.mu.RUnlock()
				req.result <- cacheResponse{input: req.number, output: value}
				continue
			}
			c.mu.RUnlock()

			if req.number < 0 {
				req.result <- cacheResponse{input: req.number, err: fmt.Errorf("negative number %d", req.number)}
				continue
			}

			value := req.number * req.number

			c.mu.Lock()
			c.cache[req.number] = value
			c.mu.Unlock()

			req.result <- cacheResponse{input: req.number, output: value}

		}
	}
}

// закрытия кэша
func (c *AsyncCache) Shutdown() {
	close(c.request)
	c.wg.Wait()
}

func main() {
	//
	ctx := context.Background()
	numbers := []int{2, -3, 2, 4, 0}
	results, errors, err := processWithCache(ctx, numbers, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("Results: ", results)
	fmt.Println("Errors: ", errors)

	//отмена контекстом
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//
	//time.Sleep(2 * time.Second)
	//numbers = []int{5, 6}
	//results, errors, err = processWithCache(ctx, numbers, workers)
	//if err != nil {
	//	fmt.Printf("Error: %v\n", err)
	//}
	//
	//fmt.Println("Results: ", results)
	//fmt.Println("Errors: ", errors)

}

func processWithCache(ctx context.Context, numbers []int, workers int) ([]Result, []Error, error) {

	if workers == 0 {
		workers = 1
	}

	cache := NewAsyncCache(workers)
	defer cache.Shutdown()
	results := make([]Result, 0, len(numbers))
	errors := make([]Error, 0)
	responseCh := make(chan cacheResponse, len(numbers))

	//уберем дубликаты
	var uniquedeNumbers []int
	uniqMap := make(map[int]struct{})
	for _, d := range numbers {
		if _, existed := uniqMap[d]; !existed {
			uniqMap[d] = struct{}{}
			uniquedeNumbers = append(uniquedeNumbers, d)
		}
	}
	//**************************
	for _, n := range uniquedeNumbers {
		resultChan := make(chan cacheResponse, 1)
		select {
		case cache.request <- cacheRequest{number: n, result: resultChan, ctx: ctx}:
			go func() {
				select {
				case res := <-resultChan:
					responseCh <- res
				case <-ctx.Done():
					responseCh <- cacheResponse{input: n, err: ctx.Err()}
				}
			}()
		case <-ctx.Done():
			responseCh <- cacheResponse{input: n, err: ctx.Err()}
		}
	}
	//for i := 0; i < len(numbers); i++ {
	for i := 0; i < len(uniquedeNumbers); i++ {
		select {
		case res := <-responseCh:
			if res.err != nil {
				errors = append(errors, Error{Input: res.input, Error: res.err})
			} else {
				results = append(results, Result{Input: res.input, Output: res.output})
			}
		case <-ctx.Done():
			return results, errors, ctx.Err()
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Input < results[j].Input
	})

	return results, errors, nil

}
