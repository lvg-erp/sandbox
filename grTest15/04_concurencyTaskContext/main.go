package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tasks := []struct {
		id       int
		duration time.Duration
	}{
		{
			id:       1,
			duration: 1 * time.Second,
		},
		{
			id:       2,
			duration: 2 * time.Second,
		},
		{
			id:       3,
			duration: 4 * time.Second,
		},
		{
			id:       4,
			duration: 3 * time.Second,
		},
	}

	resultCh := make(chan int, len(tasks))
	stop := make(chan struct{})
	for _, task := range tasks {
		wg.Add(1)
		go func(tid int, tTime time.Duration) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				fmt.Printf("Task canceled %d", tid)
				return
			case <-time.After(tTime):
				select {
				case resultCh <- tid:
				default:
				}
			}
		}(task.id, task.duration)
	}

	go func() {
		wg.Wait()
		close(stop)
	}()

	select {
	case tid := <-resultCh:
		fmt.Printf("Задача %d завершилась первой, останавливаем остальные\n", tid)
		cancel()
	case <-stop:
		fmt.Println("Все задачи завершены успешно")
	}

}
