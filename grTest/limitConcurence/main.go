package main

import (
	"fmt"
	"sync"
)

func main() {
	const workers = 2
	tasks := []string{"abs", "drywq", "tratray"}

	result := processTask(tasks, workers)

	fmt.Println(result)

}

func processTask(tasks []string, workers int) []int {
	results := make([]int, 0, len(tasks))
	taskCh := make(chan struct {
		index int
		task  string
	}, len(tasks))
	resultCh := make(chan struct{ index, result int }, len(tasks))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				resultCh <- struct{ index, result int }{index: i, result: len(task.task)}
			}
		}()
	}
	for i, task := range tasks {
		taskCh <- struct {
			index int
			task  string
		}{index: i, task: task}
	}
	close(taskCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for ch := range resultCh {
		results = append(results, ch.result)
	}

	return results
}
