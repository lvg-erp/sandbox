package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	tasks = []int{1, 2, 3, 4, 5, 6}
	wg    sync.WaitGroup
)

type TaskResult struct {
	IdTask    int
	Status    string
	TimeReady int
}

const timeout = 2

func NewTaskRec() *TaskResult {
	return &TaskResult{}

}

func (ts *TaskResult) Record(id int, status string, timeR int) {
	ts.IdTask = id
	ts.Status = status
	ts.TimeReady = timeR
}

//func (ts *TaskResult) tsMonitor(idT int) {
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	defer cancel()
//	select {
//	case <-ctx.Done():
//		ts.Record(idT, "false", 1)
//	default:
//		ts.Record(idT, "success", 1)
//	}
//}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	resultsChan := make(chan *TaskResult, len(tasks))

	for _, taskID := range tasks {
		wg.Go(func() {
			tr := NewTaskRec()
			ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
			defer cancel()

			sleepSeconds := r.Intn(5) + 1
			start := time.Now()
			select {
			case <-time.After(time.Duration(sleepSeconds) * time.Second):
				elapsed := time.Since(start).Seconds()
				if elapsed > timeout {
					tr.Record(taskID, "cancel by timeout", int(elapsed))
				} else {
					tr.Record(taskID, "success", int(elapsed))
				}
				resultsChan <- tr
			case <-ctx.Done():
				elapsed := time.Since(start).Seconds()
				tr.Record(taskID, "timeout", int(elapsed))
			}

		})
	}

	wg.Wait()
	close(resultsChan)

	for res := range resultsChan {
		fmt.Printf("Task %d: Status: %s, time: %d seconds\n", res.IdTask, res.Status, res.TimeReady)
	}
}
