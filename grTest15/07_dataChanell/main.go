package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	numChan := 5
	chanRes := make([]chan int, numChan)
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < numChan; i++ {
		ch := make(chan int)
		chanRes[i] = ch
		wg.Add(1)
		go func(c chan int, id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					close(c)
					return
				default:
					val := rand.Intn(100)
					c <- val
					time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)
				}
			}
		}(ch, i)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Время истекло. Завершение.")
			wg.Wait() // Ждем завершения генераторов
			return
		default:
			// Постараемся сразу проверить все каналы
			for i, ch := range chanRes {
				select {
				case v, ok := <-ch:
					if ok {
						fmt.Printf("Источник %d: %d\n", i+1, v)
					}
				default:

				}
			}
			// Немного подождем, чтобы не долбить CPU
			time.Sleep(50 * time.Millisecond)
		}
	}
}
