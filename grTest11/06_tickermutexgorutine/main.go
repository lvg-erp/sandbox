package main

import (
	"fmt"
	"sync"
	"time"
)

const workers = 5

type Monitor struct {
	jobs map[string]int
	mu   sync.Mutex
	stop chan struct{}
}

func NewMonitor() *Monitor {
	j := make(map[string]int)
	m := &Monitor{
		jobs: j,
		stop: make(chan struct{}),
	}
	go m.monitor()
	return m
}

func main() {

	var (
		wg sync.WaitGroup
	)

	monitor := NewMonitor()

	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 1; j < 100; j++ {
				select {
				case <-monitor.stop:
					return
				default:
					monitor.Record(fmt.Sprintf("worker_record%d", idx))
					time.Sleep(500 * time.Millisecond)
				}
			}
		}(i)
	}

	time.Sleep(10 * time.Second)
	monitor.Stop()

	wg.Wait()

	fmt.Println("Prefinal Snapshot:", monitor.Snapshot())
	monitor.Reset()
	fmt.Println("Final Snapshot:", monitor.Snapshot())

}

func (m *Monitor) Record(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job]++
}

func (m *Monitor) Snapshot() map[string]int {

	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]int)
	for i, v := range m.jobs {
		result[i] = v
	}

	return result

}

func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = make(map[string]int)
}

func (m *Monitor) Stop() {
	//m.mu.Lock()
	//defer m.mu.Unlock()
	close(m.stop)
}

func (m *Monitor) monitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			fmt.Println("Stats:", m.Snapshot())
		}
	}
}
