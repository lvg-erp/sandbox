package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tik := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Canceled by context on timeout")
			return
		default:
			tik += 1
			fmt.Println("Worker ", tik)
			time.Sleep(100 * time.Millisecond)
		}
	}

}
