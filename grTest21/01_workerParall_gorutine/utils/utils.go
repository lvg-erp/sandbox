package utils

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func processorOrder(ctx context.Context, id int) error {
	// По условию случайное время от 1 - 3 сек
	duration := time.Duration(1+rand.Intn(3)) * time.Second
	fmt.Printf("Заказ %d принят в обработку (займет %v)\n", id, duration)
	select {
	case <-time.After(duration):
		fmt.Printf("Заказ %d обработан за %v\n", id, duration)
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func Worker(ctx context.Context, id int, orders <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for orderID := range orders {
		if err := processorOrder(ctx, orderID); err != nil {
			fmt.Printf("Заказ %d не обработан\n", id)
			return
		}
	}
}
