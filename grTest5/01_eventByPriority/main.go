package main

import (
	"context"
	rand_m "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const (
	low     = "03"
	medium  = "02"
	high    = "01"
	workers = 3
)

type Source struct {
	ID       int64
	Priority string
	Time     time.Time
}

type ByPriority []Source

func (a ByPriority) Len() int      { return len(a) }
func (a ByPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool {
	priorityOrder := map[string]int{high: 3, medium: 2, low: 1}
	return priorityOrder[a[i].Priority] > priorityOrder[a[j].Priority]
}

func main() {

	var wg sync.WaitGroup
	var mu sync.Mutex
	var result []Source
	ch := make(chan Source)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	priorities := []string{low, medium, high}

	fmt.Printf("Programm started by: %v\n", time.Now())

	for i := 0; i < workers; i++ {
		wg.Add(1)
		interval := time.Duration(i+1) * time.Second
		go func(interval time.Duration) {
			defer wg.Done()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for {
				randomIndex, err := rand_m.Int(rand_m.Reader, big.NewInt(1000))
				if err != nil {
					log.Fatalf("Failed to generate random number: %v", err)
				}
				select {
				case <-ctx.Done():
					return
				case t := <-ticker.C:
					select {
					case ch <- Source{
						ID:       randomIndex.Int64(),
						Priority: priorities[r.Intn(len(priorities))],
						Time:     t,
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}(interval)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	//вывод с задержской каждые 5 сек
	outputTicker := time.NewTicker(5 * time.Second)
	defer outputTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(result) > 0 {
				fmt.Printf("Finished at %s:\n", time.Now().Format(time.RFC3339))
				mu.Lock()
				sort.Sort(ByPriority(result))
				for _, r := range result {
					fmt.Printf("This struct by priority Source %d, Priority: %s, Time: %s\n", r.ID, r.Priority, r.Time)
				}
			}
			mu.Unlock()
			fmt.Printf("Programm finished by: %v\n", time.Now())
			return
		case res, ok := <-ch:
			if !ok {
				return
			}
			mu.Lock()
			result = append(result, res)
			mu.Unlock()
		case <-outputTicker.C: // выводим каждые 5 минут результаты
			if len(result) > 0 {
				fmt.Printf("Batch at %s:\n", time.Now().Format(time.RFC3339))
				mu.Lock()
				sort.Sort(ByPriority(result))
				for _, r := range result {
					fmt.Printf("This struct by priority Source %d, Priority: %s, Time: %s\n", r.ID, r.Priority, r.Time)
				}
				mu.Unlock()
			}
		}
	}

}
