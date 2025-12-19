package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		longLongOperation(ctx)
	}()

	wg.Wait()

}

func longLongOperation(ctx context.Context) {
	fmt.Println("Starting long operation...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("operation canceled by timeout")
			return
		default:
			time.Sleep(2 * time.Second)
			fmt.Println("operation completed successfully")
		}
	}

}
