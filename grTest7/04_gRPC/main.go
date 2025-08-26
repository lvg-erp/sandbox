package main

import (
	"context"
	"fmt"
	"github.com/lvg-erp/sandbox/grTest7/04_gRPC/client"
	"github.com/lvg-erp/sandbox/grTest7/04_gRPC/server"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan string, 10)
	//Старт сервера
	wg.Add(1)
	go server.RunServer(ctx, &wg)
	// даем серверу запуститься
	time.Sleep(1 * time.Second)
	//клиент
	wg.Add(1)
	go client.RunClient(ctx, &wg, results)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}

	time.Sleep(5 * time.Second)
	cancel()

}
