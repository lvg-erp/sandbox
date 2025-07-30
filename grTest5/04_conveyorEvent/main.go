package main

import (
	"context"
	"fmt"
	"golang.org/x/exp/rand"
	"strings"
	"sync"
	"time"
)

const workers = 3

func main() {

	var wg sync.WaitGroup
	var mu sync.Mutex
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ch := make(chan string, workers)
	chF := make(chan string, workers)
	var result []string

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rnd := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
			interval := rnd.Intn(workers) + 1
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					select {
					case <-ctx.Done():
						return
					case ch <- string(generateRandomBytes()):
					}
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for str := range ch {
			if !filteringVowelLetter(str) {
				select {
				case chF <- str:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
		close(chF)
	}()

	tickerF := time.NewTicker(5 * time.Second)
	defer tickerF.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Program finished, final result: %v\n", result)
			return
		case <-tickerF.C:
			select {
			case <-ctx.Done():
				return
			case str := <-chF:
				mu.Lock()
				result = append(result, str)
				fmt.Printf("Processed events at %s: %v\n", time.Now().Format(time.RFC3339), result)
				mu.Unlock()
			}
		}
	}
}

func filteringVowelLetter(in string) bool {
	filter := []string{"a", "e", "i", "o", "u"}
	for _, f := range filter {
		if strings.Contains(in, f) {
			return true
		}
	}

	return false
}

func generateRandomBytes() []byte {
	fixedSet := []string{
		"K9xP4m",
		"tR7wQ2",
		"Z8nL5j",
		"aB1cD2",
		"Xy9zW4",
		"Pq3Rs5",
		"Lm8Nk7",
		"Jh2Tv9",
		"Fg6Yx1",
		"De4Uw8",
	}
	rand.Seed(uint64(time.Now().UnixNano()))
	return []byte(fixedSet[rand.Intn(len(fixedSet))])
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
