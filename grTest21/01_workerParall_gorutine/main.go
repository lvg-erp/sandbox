package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	"workerparallel/utils"
)

const (
	workers     = 5
	totalOrders = 10
	timeOut     = 5 * time.Second
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	orders := make(chan int)
	var wg sync.WaitGroup

	// воркеры
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go utils.Worker(ctx, i, orders, &wg)
	}

	// генерируем заказаы
	go func() {
		defer close(orders)
		for i := 0; i < totalOrders; i++ {
			select {
			case orders <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	// ожидаем завершения
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("Все заказы обработаны успешно!")
	case <-ctx.Done():
		fmt.Println("timeout")
	}

}
