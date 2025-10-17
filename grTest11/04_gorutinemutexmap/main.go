package main

import (
	"fmt"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type ConcurrentStats struct {
	cs map[string]int
	mu sync.Mutex
}

func NewConcurrentStats() *ConcurrentStats {
	cs := make(map[string]int)
	return &ConcurrentStats{
		cs: cs,
	}
}

func main() {
	var wg sync.WaitGroup
	stats := NewConcurrentStats()
	context, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 1; j <= 100; j++ {
				select {
				case <-context.Done():
					return
				default:
					stats.Record(fmt.Sprintf("button%d_click", idx))
					//stats.Record(fmt.Sprintf("button%d_click", idx%2))
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-context.Done():
				return
			case <-ticker.C:
				fmt.Println("Snapshot: ", stats.SnapShot())
			}
		}
	}()
	time.Sleep(5 * time.Second)
	cancel()
	wg.Wait()
	stats.Reset()
	fmt.Println("Final Snapshot: ", stats.SnapShot())
}

func (s *ConcurrentStats) Record(event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cs[event]++
}

func (s *ConcurrentStats) SnapShot() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	rMap := make(map[string]int)
	for key, value := range s.cs {
		rMap[key] = value
	}
	return rMap

}

func (s *ConcurrentStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	//for i, _ := range s.cs {
	//	delete(s.cs, i)
	//}
	s.cs = make(map[string]int)
}
