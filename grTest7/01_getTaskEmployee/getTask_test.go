package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	_ "time"

	"github.com/stretchr/testify/assert"
)

func processEmployees(api *MockAPI, employeeIDs []int, workers int) ([]Employee, []ErrorsInStream, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskCh := make(chan int, 10)
	resultCh := make(chan Result, 10)
	resultErrorsCh := make(chan ErrorsInStream, 10)
	var results []Employee
	var resultErrors []ErrorsInStream
	var wg sync.WaitGroup
	var mu sync.Mutex
	// var once sync.Once

	for _, id := range employeeIDs {
		if id <= 0 {
			continue
		}
		taskCh <- id
	}
	close(taskCh)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for id := range taskCh {
				select {
				case <-ctx.Done():
					resultErrorsCh <- ErrorsInStream{
						WorkerID: workerID,
						Error:    fmt.Errorf("worker %d cancelled: %v", workerID, ctx.Err()),
					}
					return
				default:
					employee, err := api.GetEmployee(ctx, id)
					if err.Error != nil {
						resultErrorsCh <- ErrorsInStream{
							WorkerID: workerID,
							Error:    fmt.Errorf("error for employee %d: %v", id, err.Error),
						}
						// once.Do(cancel) // Закомментировано
						return
					}
					employee.Bonus = float64(employee.Years) * 1000
					employee.Processed = true
					resultCh <- Result{Employee: employee}
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(resultErrorsCh)
	}()

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				resultCh = nil
			} else {
				mu.Lock()
				results = append(results, result.Employee)
				mu.Unlock()
			}
		case errResult, ok := <-resultErrorsCh:
			if !ok {
				resultErrorsCh = nil
			} else {
				mu.Lock()
				resultErrors = append(resultErrors, errResult)
				mu.Unlock()
			}
		}
		if resultCh == nil && resultErrorsCh == nil {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return results, resultErrors, nil
}

func Test_ProcessEmployees(t *testing.T) {
	tests := []struct {
		name              string
		employeeIDs       []int
		workers           int
		expectedResults   []Employee
		expectedErrorsLen int
		expectedErrorMsg  string
	}{
		{
			name:        "SuccessCase",
			employeeIDs: []int{1, 2},
			workers:     2,
			expectedResults: []Employee{
				{ID: 1, Name: "Employee 1", Years: 2, Bonus: 2000, Processed: true},
				{ID: 2, Name: "Employee 2", Years: 4, Bonus: 4000, Processed: true},
			},
			expectedErrorsLen: 0,
			expectedErrorMsg:  "",
		},
		{
			name:        "ErrorCase",
			employeeIDs: []int{1, 2, 3, 4, 5},
			workers:     2,
			expectedResults: []Employee{
				{ID: 1, Name: "Employee 1", Years: 2, Bonus: 2000, Processed: true},
				{ID: 2, Name: "Employee 2", Years: 4, Bonus: 4000, Processed: true},
				{ID: 4, Name: "Employee 4", Years: 8, Bonus: 8000, Processed: true},
				{ID: 5, Name: "Employee 5", Years: 10, Bonus: 10000, Processed: true},
			},
			expectedErrorsLen: 1,
			expectedErrorMsg:  "employee not found",
		},
		{
			name:              "EmptyInputCase",
			employeeIDs:       []int{},
			workers:           2,
			expectedResults:   []Employee{},
			expectedErrorsLen: 0,
			expectedErrorMsg:  "",
		},
		{
			name:        "InvalidIDCase",
			employeeIDs: []int{-1, 0, 1, 2},
			workers:     2,
			expectedResults: []Employee{
				{ID: 1, Name: "Employee 1", Years: 2, Bonus: 2000, Processed: true},
				{ID: 2, Name: "Employee 2", Years: 4, Bonus: 4000, Processed: true},
			},
			expectedErrorsLen: 0,
			expectedErrorMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &MockAPI{}
			results, resultErrors, err := processEmployees(api, tt.employeeIDs, tt.workers)
			assert.NoError(t, err)

			assert.Equal(t, len(tt.expectedResults), len(results), "Results length mismatch")
			for i, expected := range tt.expectedResults {
				if i < len(results) {
					assert.Equal(t, expected.ID, results[i].ID, "Employee ID mismatch")
					assert.Equal(t, expected.Name, results[i].Name, "Employee Name mismatch")
					assert.Equal(t, expected.Years, results[i].Years, "Employee Years mismatch")
					assert.Equal(t, expected.Bonus, results[i].Bonus, "Employee Bonus mismatch")
					assert.True(t, results[i].Processed, "Employee Processed mismatch")
				}
			}

			for i := 1; i < len(results); i++ {
				assert.LessOrEqual(t, results[i-1].ID, results[i].ID, "Results not sorted by ID")
			}

			assert.Equal(t, tt.expectedErrorsLen, len(resultErrors), "Errors length mismatch")
			if tt.expectedErrorsLen > 0 {
				assert.Contains(t, resultErrors[0].Error.Error(), tt.expectedErrorMsg, "Error message mismatch")
			}
		})
	}
}

// с once.Do(cancel)
//func TestProcessEmployeesWithCancel(t *testing.T) {
//	api := &MockAPI{}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	taskCh := make(chan int, 10)
//	resultCh := make(chan Result, 10)
//	resultErrorsCh := make(chan ErrorsInStream, 10)
//	var results []Employee
//	var resultErrors []ErrorsInStream
//	var wg sync.WaitGroup
//	var mu sync.Mutex
//	var once sync.Once
//
//	employeeIDs := []int{1, 2, 3, 4, 5}
//	workers := 2
//
//	for _, id := range employeeIDs {
//		if id <= 0 {
//			continue
//		}
//		taskCh <- id
//	}
//	close(taskCh)
//
//	for i := 0; i < workers; i++ {
//		wg.Add(1)
//		go func(workerID int) {
//			defer wg.Done()
//			for id := range taskCh {
//				select {
//				case <-ctx.Done():
//					resultErrorsCh <- ErrorsInStream{
//						WorkerID: workerID,
//						Error:    fmt.Errorf("worker %d cancelled: %v", workerID, ctx.Err()),
//					}
//					return
//				default:
//					employee, err := api.GetEmployee(ctx, id)
//					if err.Error != nil {
//						resultErrorsCh <- ErrorsInStream{
//							WorkerID: workerID,
//							Error:    fmt.Errorf("error for employee %d: %v", id, err.Error),
//						}
//						once.Do(cancel) // Прерываем при первой ошибке
//						return
//					}
//					employee.Bonus = float64(employee.Years) * 1000
//					employee.Processed = true
//					resultCh <- Result{Employee: employee}
//				}
//			}
//		}(i)
//	}
//
//	go func() {
//		wg.Wait()
//		close(resultCh)
//		close(resultErrorsCh)
//	}()
//
//	for {
//		select {
//		case result, ok := <-resultCh:
//			if !ok {
//				resultCh = nil
//			} else {
//				mu.Lock()
//				results = append(results, result.Employee)
//				mu.Unlock()
//			}
//		case errResult, ok := <-resultErrorsCh:
//			if !ok {
//				resultErrorsCh = nil
//			} else {
//				mu.Lock()
//				resultErrors = append(resultErrors, errResult)
//				mu.Unlock()
//			}
//		}
//		if resultCh == nil && resultErrorsCh == nil {
//			break
//		}
//	}
//
//	// Проверка результатов
//	assert.LessOrEqual(t, len(results), 2, "Expected at most 2 results before cancellation")
//	for _, r := range results {
//		assert.Contains(t, []int{1, 2}, r.ID, "Unexpected employee ID in results")
//		assert.True(t, r.Processed, "Employee not processed")
//		assert.Equal(t, float64(r.Years)*1000, r.Bonus, "Bonus calculation incorrect")
//	}
//
//	// Проверка ошибок
//	assert.GreaterOrEqual(t, len(resultErrors), 1, "Expected at least one error")
//	assert.Contains(t, resultErrors[0].Error.Error(), "employee not found", "Expected error for employee 3")
//}
