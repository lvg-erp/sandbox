package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	const stream = 3
	arrNumber := []int{3, 4, 6, 7, 8, 9, 2}
	var wg sync.WaitGroup
	chanOut := make(chan int, len(arrNumber))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	chunkSize := (len(arrNumber) + stream - 1) / stream
	for i := 0; i < stream; i++ {
		start := chunkSize * i
		end := start + chunkSize
		if end > len(arrNumber) {
			end = len(arrNumber)
		}

		s, e := start, end
		wg.Go(func() {
			chunk := arrNumber[s:e]
			for e := 0; e < len(chunk); e++ {
				r := fnk(chunk[e])
				chanOut <- r
			}
		})
	}

	go func() {
		wg.Wait()
		close(chanOut)
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Истекло время (3 секунды). Завершаем работу.")
			for c := range chanOut {
				fmt.Println("Результат после тайм-аута: ", c)
			}
			return
		case c, ok := <-chanOut:
			time.Sleep(2 * time.Second)
			if !ok {
				fmt.Println("Завершено по завршению чтения из канала.")
				return
			}

			fmt.Println("Результаты чтения из канала: ", c)

		}
	}

}

func fnk(in int) int {
	return in * in
}
