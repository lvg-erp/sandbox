package main

import (
	"context"
	"fmt"
	"sync"
)

const workers = 3

type Result struct {
	Input  string
	Output int
}

type Error struct {
	Input string
	Error error
}

type cacheResponse struct {
	input  string
	output int
	err    error
}

type cacheRequest struct {
	inputStr string
	result   chan cacheResponse
	ctx      context.Context
}

type AsyncCache struct {
	cache   map[string]int
	mu      sync.RWMutex
	request chan cacheRequest
	wg      sync.WaitGroup
}

func NewAsyncCache(workers int) *AsyncCache {
	cache := &AsyncCache{
		cache:   make(map[string]int),
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
			req.result <- cacheResponse{input: req.inputStr, err: req.ctx.Err()}
			continue
		default:
			c.mu.RLock()
			if value, exists := c.cache[req.inputStr]; exists {
				c.mu.RUnlock()
				req.result <- cacheResponse{input: req.inputStr, output: value}
				continue
			}
			c.mu.RUnlock()

			if req.inputStr == "" {
				req.result <- cacheResponse{input: req.inputStr, err: fmt.Errorf("empty string")}
				continue
			}

			value := len(req.inputStr)

			c.mu.Lock()
			c.cache[req.inputStr] = value
			c.mu.Unlock()

			req.result <- cacheResponse{input: req.inputStr, output: value}
		}
	}
}

// закрытия кэша
func (c *AsyncCache) Shutdown() {
	close(c.request)
	c.wg.Wait()
}

func processStringsWithCache(ctx context.Context, inputs []string, workers int) ([]Result, []Error, error) {

	if workers <= 0 {
		workers = 1
	}
	cache := NewAsyncCache(workers)
	defer cache.Shutdown()

	results := make([]Result, 0, len(inputs))
	errors := make([]Error, 0)

	responseCh := make(chan cacheResponse, len(inputs))

	for _, s := range inputs {
		//for _, s := range ss {
		resultChan := make(chan cacheResponse, 1)
		select {
		case cache.request <- cacheRequest{inputStr: s, result: resultChan, ctx: ctx}:
			go func(input string) {
				select {
				case res := <-resultChan:
					responseCh <- res
				case <-ctx.Done():
					responseCh <- cacheResponse{
						input: input, err: ctx.Err(),
					}
				}
			}(s)
		case <-ctx.Done():
			responseCh <- cacheResponse{input: s, err: ctx.Err()}
		}
	}
	//TODO может все таки обходить по каналу?
	for i := 0; i < len(inputs); i++ {
		select {
		case res := <-responseCh:
			if res.err != nil {
				errors = append(errors, Error{
					Input: res.input,
					Error: res.err,
				})
			} else {
				results = append(results, Result{
					Input:  res.input,
					Output: res.output,
				})
			}
		case <-ctx.Done():
			return results, errors, ctx.Err()
		}
	}

	// TODO: sort???
	return results, errors, nil

}

func main() {
	ctx := context.Background()
	numbers := []string{"hello world", "", "test", "hello world"}
	results, errors, err := processStringsWithCache(ctx, numbers, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("Results: ", results)
	fmt.Println("Errors: ", errors)
}
