package main

import (
	"context"
	"fmt"
	"time"

	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// сигнал прерывания
	go func() {
		sig := <-sigs
		fmt.Println("\nПолучен сигнал остановки:", sig)
		cancel()
	}()

	i := 0
	wg.Go(func() {

		for {
			i++
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				fmt.Printf("Счетчик %d\n", i)
			}
		}
	})

	wg.Wait()
}
