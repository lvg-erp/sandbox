package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 19*time.Microsecond)
	defer cancel()

	sts := []string{"hello", "world", "change"}
	result, err, count := processString(ctx, sts, 2)
	if err != nil {
		log.Panic("Alarm")
	}
	fmt.Println(count)
	fmt.Println(result)
}

func processString(ctx context.Context, strs []string, workers int) ([]string, error, int) {
	if workers <= 0 {
		workers = 1
	}
	result := make([]string, len(strs))
	resultCh := make(chan struct {
		index   int
		value   string
		success bool
	}, len(strs))
	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	var errs []error

	chunkSize := (len(strs) + workers - 1) / workers
	for i := 0; i < workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(strs) {
			end = len(strs)
		}
		if start < len(strs) {
			wg.Add(1)
			go func(chunk []string, startIndex int) {
				defer wg.Done()
				for i, c := range chunk {
					select {
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					default:
						upStr, err := stringToUpper(c)
						resultCh <- struct {
							index   int
							value   string
							success bool
						}{index: startIndex + i, value: upStr, success: err == nil}
						if err != nil {
							errCh <- err
						}
					}
				}
			}(strs[start:end], start)
		}
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	successCount := 0
	for rc := range resultCh {
		if rc.success {
			successCount++
		}
		result[rc.index] = rc.value
	}

	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return result, errors.Join(errs...), successCount
	}
	return result, nil, successCount
}

func stringsToUpper(in []string) []string {
	var resultUpper []string
	for _, s := range in {
		resultUpper = append(resultUpper, strings.ToUpper(s))
	}

	return resultUpper

}

func stringToUpper(in string) (string, error) {

	return strings.ToUpper(in), nil
}
