package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

func main() {
	err := generateReadNumber(context.Background(), 3)
	if err != nil {
		log.Panic("Errors from stream")
	}
}

func generateReadNumber(ctx context.Context, consumers int) error {

	numbers := make(chan int, 10000)
	var errs []error
	var wg sync.WaitGroup
	errChan := make(chan error, consumers)

	go func() {
		for i := 0; i < 10000; i++ {
			res := i * 2
			numbers <- res
		}
		close(numbers)
	}()

	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for num := range numbers {
				select {
				case <-ctx.Done():
					return
				default:
					fmt.Printf("Consumer %d processed %d\n", idx, num)
				}
			}
		}(i + 1)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil

}
