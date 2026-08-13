package task

import (
	"fmt"
	"time"
)

// ===== ТИПЫ ЗАДАЧ =====

type SimpleTask struct {
	ID   int
	Name string
}

func (t *SimpleTask) Process() error {
	time.Sleep(time.Duration(100+t.ID%400) * time.Millisecond)

	if t.ID%20 == 0 {
		return fmt.Errorf("task %d failed: network error", t.ID)
	}
	if t.ID%33 == 0 {
		panic(fmt.Sprintf("task %d panic: unexpected data", t.ID))
	}
	return nil
}

type APITask struct {
	ID   int
	URL  string
	Body string
}

func (t *APITask) Process() error {
	time.Sleep(time.Duration(200+t.ID%300) * time.Millisecond)

	switch t.ID % 7 {
	case 0:
		return fmt.Errorf("API timeout for %s", t.URL)
	case 3:
		panic("connection reset by peer")
	}
	return nil
}

type CPUTask struct {
	ID   int
	Size int
}

func (t *CPUTask) Process() error {
	sum := 0
	for i := 0; i < t.Size*1000; i++ {
		sum += i * i
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}
