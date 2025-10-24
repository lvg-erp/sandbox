package main

import (
	"context"
	"fmt"
	"time"
)

type FibResult struct {
	Index int
	Value int
}

func fib(n int) int {
	if n < 2 {
		return n
	}

	return fib(n-1) + fib(n-2)
}

func generateFib(ctx context.Context, ch chan<- FibResult) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			close(ch)
			return
		case <-ticker.C:
			if i > 100 {
				close(ch)
				return
			}
			ch <- FibResult{Index: i, Value: fib(i)}
		}
	}

}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		ch = make(chan FibResult)
	)

	go generateFib(ctx, ch)

	for c := range ch {
		fmt.Printf("Fib(%d) -- %d\n", c.Index, c.Value)
	}

}
