package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const workers = 3

type Result struct {
	Input      int
	Output     int
	IsPriority bool
}

type Error struct {
	WorkerID int
	Error    error
}

func main() {
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	//defer cancel()
	//priorityTasks := []int{300, 320, 444, 557}
	//normalTasks := []int{1, 3, 7, 5}
	//r, errors, err := processTasksWithPriority(ctx, priorityTasks, normalTasks, workers)
	//
	//if err != nil {
	//	fmt.Printf("Error: %v\n", err)
	//}
	//
	//fmt.Println("Priority tasks: ", r)
	//fmt.Println("Normal tasks: ", n)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	priorityTasks := []int{7, 8}
	normalTasks := []int{9, 10}
	r, errors, err := processTasksWithPriority(ctx, priorityTasks, normalTasks, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println("Results:", r)
	fmt.Println("Errors:", errors)

	//ctx = context.Background()
	//priorityTasks = []int{1, 2}
	//normalTasks = []int{}
	//r, errors, err = processTasksWithPriority(ctx, priorityTasks, normalTasks, 2)
	//if err != nil {
	//	fmt.Printf("Error: %v\n", err)
	//}
	//fmt.Println("Results:", r)
	//fmt.Println("Errors:", errors)

}

func processTasksWithPriority(ctx context.Context, priorityTasks, normalTasks []int, workers int) ([]Result, []Error, error) {

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]Result, 0, len(priorityTasks)+len(normalTasks))
	errors := make([]Error, 0)

	if len(priorityTasks) == 0 || len(normalTasks) == 0 {
		return nil, nil, fmt.Errorf("input data is empty")
	}

	if workers == 0 {
		workers = 1
	}
	priorityChan := make(chan int, len(priorityTasks))
	normalChan := make(chan int, len(normalTasks))

	for _, pt := range priorityTasks {
		priorityChan <- pt
	}
	close(priorityChan)
	for _, nt := range normalTasks {
		normalChan <- nt
	}
	close(normalChan)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					errors = append(errors, Error{WorkerID: id, Error: ctx.Err()})
					mu.Unlock()
					return
				case pr, ok := <-priorityChan:
					if !ok {
						select {
						case <-ctx.Done():
							mu.Lock()
							errors = append(errors, Error{WorkerID: id, Error: ctx.Err()})
							mu.Unlock()
							return
						case nrm, ok := <-normalChan:
							if !ok {
								// Оба канала пусты, завершаем
								return
							}
							mu.Lock()
							results = append(results, Result{
								Input:      nrm,
								Output:     nrm * 2,
								IsPriority: false,
							})
							mu.Unlock()
						}
					} else {
						mu.Lock()
						results = append(results, Result{
							Input:      pr,
							Output:     pr * 2,
							IsPriority: true,
						})
						mu.Unlock()
					}
				}
			}
		}(i)

	}

	wg.Wait()

	return results, errors, nil

}
