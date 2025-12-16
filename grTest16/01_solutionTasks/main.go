package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
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
			duration: 3 * time.Second,
		},
		{
			id:       4,
			duration: 4 * time.Second,
		},
	}

	var wg sync.WaitGroup
	storeChan := make(chan int, len(tasks))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// ловим сигнал прерывания
	go func() {
		sig := <-sigs
		fmt.Println("\nПолучен сигнал:", sig)
		cancel()
	}()

	for _, task := range tasks {
		wg.Add(1)
		go func(t struct {
			id       int
			duration time.Duration
		}) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				fmt.Printf("Task canceled %d\n", t.id)
				return
			case <-time.After(t.duration):
				storeChan <- t.id
			}
		}(task)
	}

	go func() {
		wg.Wait()
		close(storeChan)
	}()

	for r := range storeChan {
		fmt.Printf("Успешно завершена задача: %d\n", r)
	}

}
