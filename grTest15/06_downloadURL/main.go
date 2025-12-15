package main

import (
	"context"
	"fmt"
	"golang.org/x/exp/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	rand.Seed(uint64(time.Now().UnixNano()))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urls := []string{
		"http://example.com/file1",
		"http://example.com/file2",
		"http://example.com/file3",
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// ловим сигнал прерывания
	go func() {
		sig := <-sigs
		fmt.Println("\nПолучен сигнал:", sig)
		cancel()
	}()

	rs := make(chan string, len(urls))

	for _, url := range urls {
		wg.Add(1)

		go func(u string) {
			defer wg.Done()
			downloadingFile(ctx, u, rs)
		}(url)
	}

	go func() {
		wg.Wait()
		close(rs)
	}()

	// Читаем результаты по мере поступления
	for res := range rs {
		fmt.Println(res)
	}

	fmt.Println("Программа завершена.")
}

func downloadingFile(ctx context.Context, url string, resultChan chan<- string) {
	// Генерируем случайную длительность скачивания
	duration := time.Duration(rand.Intn(3)+1) * time.Second

	select {
	case <-time.After(duration):
		size := rand.Intn(4000) + 1000
		resultChan <- fmt.Sprintf("URL: %s, Size: %d байт", url, size)
	case <-ctx.Done():
		resultChan <- fmt.Sprintf("URL: %s, Загрузка отменена", url)
	}
}
