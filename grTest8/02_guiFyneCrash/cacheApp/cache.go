package cacheApp

import (
	"context"
	"fmt"
	"sync"
)

type Result struct {
	Input  string
	Output int
}

type Error struct {
	Input string
	Error error
}

type cacheRequest struct {
	inputStr string
	result   chan cacheResponse
	ctx      context.Context
}

type cacheResponse struct {
	input  string
	output int
	err    error
}

type Cache struct {
	cache    map[string]int
	wg       sync.WaitGroup
	requests chan cacheRequest
	mu       sync.RWMutex
}

func NewCache(workers int) *Cache {
	if workers < 1 {
		workers = 1
	}

	cache := &Cache{
		cache:    make(map[string]int),
		requests: make(chan cacheRequest, 100),
	}

	cache.wg.Add(workers)
	for i := 0; i < workers; i++ {
		cache.worker()
	}

	return cache
}

func (c *Cache) worker() {
	defer c.wg.Done()
	for req := range c.requests {
		select {
		case <-req.ctx.Done():
			req.result <- cacheResponse{input: req.inputStr, err: req.ctx.Err()}
			continue
		default:
			c.mu.RLock()
			if value, exist := c.cache[req.inputStr]; exist {
				c.mu.RUnlock()
				req.result <- cacheResponse{input: req.inputStr, output: value}
				continue
			}
			c.mu.RUnlock()
			if req.inputStr == "" {
				req.result <- cacheResponse{
					input: req.inputStr, err: fmt.Errorf("empty string")}
			}
			value := len(req.inputStr)

			c.mu.Lock()
			c.cache[req.inputStr] = value
			c.mu.Unlock()

		}

	}
}

func (c *Cache) Shutdown() {
	close(c.requests)
	c.wg.Wait()
}

func ProcessStringsWithCache(ctx context.Context, inputs []string, workers int) ([]Result, []Error, error) {
	cache := NewCache(workers)
	defer cache.Shutdown()
	results := make([]Result, 0, len(inputs))
	errors := make([]Error, 0)
	responseCh := make(chan cacheResponse, len(inputs))
	for _, str := range inputs {
		resultChan := make(chan cacheResponse, 1)
		select {
		case cache.requests <- cacheRequest{inputStr: str, result: resultChan, ctx: ctx}:
			go func(input string) {
				select {
				case res := <-resultChan:
					responseCh <- res
				case <-ctx.Done():
					responseCh <- cacheResponse{input: input, err: ctx.Err()}
				}
			}(str)
		case <-ctx.Done():
			responseCh <- cacheResponse{input: str, err: ctx.Err()}
		}
	}

	for i := 0; i < len(inputs); i++ {
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

	return results, errors, nil
}
