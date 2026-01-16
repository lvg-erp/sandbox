package main

import (
	"fmt"
	"golang.org/x/net/context"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	arrClient = []int{20, 40, 55, 66, 88, 76, 54}
	wg        sync.WaitGroup
	mu        sync.Mutex
)

type ResultClient struct {
	ClientID   int
	Result     string
	TimeResult int
}

func NewResultClient() *ResultClient {
	return &ResultClient{}
}

func (rc *ResultClient) Record(id int, status string, timeR int) {
	rc.ClientID = id
	rc.Result = status
	rc.TimeResult = timeR
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("\nПолучен сигнал прерывания", sig)
		cancel()
	}()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chanRes := make(chan *ResultClient, len(arrClient))
	for _, cID := range arrClient {
		sleepSecond := rnd.Intn(5) + 1
		start := time.Now()
		rct := NewResultClient()

		wg.Go(func() {
			time.Sleep(time.Duration(sleepSecond) * time.Second)

			elapsed := time.Since(start).Seconds()
			select {
			case <-ctx.Done():
				mu.Lock()
				defer mu.Unlock()
				rct.Record(cID, "process canceled", 0)
				chanRes <- rct
				return
			default:
				mu.Lock()
				defer mu.Unlock()
				rct.Record(cID, "process success worked", int(elapsed))
				chanRes <- rct
			}
		})
	}

	wg.Wait()
	close(chanRes)

	for c := range chanRes {
		fmt.Printf("Task %d: Status: %s, time: %d seconds\n", c.ClientID, c.Result, c.TimeResult)
	}

}
