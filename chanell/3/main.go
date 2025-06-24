package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()

	ch := make(chan int)

	go func() {
		for i := range 1000 {
			select {
			case ch <- i:
			case <-ctx.Done(): // предотварщение утечки горутин
			}

		}
		close(ch)
	}()

	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return
			}
			fmt.Println(v)
		case <-ctx.Done():
			return
		}
	}

}
