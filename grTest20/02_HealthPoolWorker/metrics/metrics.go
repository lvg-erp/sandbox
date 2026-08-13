package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"workerpool/pkg"
)

type Metrics struct {
	TasksSubmitted   atomic.Int64
	TasksCompleted   atomic.Int64
	TasksFailed      atomic.Int64
	TasksPanicked    atomic.Int64
	WorkersRestarted atomic.Int64
	ActiveWorkers    atomic.Int64
	queueSize        atomic.Int64
	mu               sync.RWMutex
	startTime        time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

func (m *Metrics) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	return fmt.Sprintf(pkg.BannerForMetrics(),
		formatDuration(uptime),
		m.TasksSubmitted.Load(),
		m.TasksCompleted.Load(),
		m.TasksFailed.Load(),
		m.TasksPanicked.Load(),
		m.ActiveWorkers.Load(),
		m.WorkersRestarted.Load(),
		m.queueSize.Load(),
		m.GetSuccessRate(),
		runtime.NumGoroutine(),
	)
}

func (m *Metrics) GetSuccessRate() float64 {
	total := m.TasksCompleted.Load() + m.TasksFailed.Load() + m.TasksPanicked.Load()
	if total == 0 {
		return 100.0
	}
	return float64(m.TasksCompleted.Load()) / float64(total) * 100
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
