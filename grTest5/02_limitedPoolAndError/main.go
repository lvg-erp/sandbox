package main

import (
	"context"
	rand_m "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

const workers = 5

type Task struct {
	ID       int64
	SourceID int64
	Time     time.Time
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	ch := make(chan Task, workers)
	chSuccess := make(chan Task, workers)
	chError := make(chan error, workers)
	var totalCount int

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			randomIndex, err := rand_m.Int(rand_m.Reader, big.NewInt(1000))
			if err != nil {
				log.Fatalf("Failed to generate random number: %v", err)
			}
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					select {
					case <-ctx.Done():
						return
					case ch <- Task{
						ID:       randomIndex.Int64(),
						SourceID: int64(i),
						Time:     time.Now(),
					}:
					}
				}
			}
		}(i)
	}

	for y := 0; y < 2; y++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				select {
				case <-ctx.Done():
					return
				case task := <-ch:
					if r.Float64() < 0.3 {
						err := fmt.Errorf("task %d, gID %v, failed at %s", task.ID, task.SourceID, task.Time.Format(time.RFC3339))
						select {
						case <-ctx.Done():
							return
						case chError <- err:
						}
					} else {
						select {
						case <-ctx.Done():
							return
						case chSuccess <- task:
							mu.Lock()
							totalCount++
							mu.Unlock()
						}
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
		close(chSuccess)
		close(chError)
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Total success tasks: %v\n", totalCount)
			return
		case t := <-chSuccess:
			fmt.Printf("Success task: Source: %v, gID: %v, time: %v\n", t.ID, t.SourceID, t.Time)
		case err := <-chError:
			fmt.Printf("Error task: Error: %v\n", err)
		}
	}
}
