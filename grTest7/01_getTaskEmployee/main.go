package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Employee struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Years     int     `json:"years"`     // стаж
	Bonus     float64 `json:"bonus"`     // премия
	Processed bool    `json:"processed"` // флаг обработки
}

type ErrorsInStream struct {
	WorkerID int
	Error    error
}

// MockAPI имитирует API для получения данных сотрудников
type MockAPI struct{}

func (api *MockAPI) GetEmployee(ctx context.Context, id int) (Employee, ErrorsInStream) {
	select {
	case <-ctx.Done():
		return Employee{}, ErrorsInStream{Error: ctx.Err()}
	default:
		time.Sleep(100 * time.Millisecond)
		// для примера смоделируем ошибку на ИД 3
		if id == 3 {
			return Employee{}, ErrorsInStream{
				WorkerID: id,
				Error:    errors.New("employee not found"),
			}
		}
		return Employee{
			ID:    id,
			Name:  fmt.Sprintf("Employee %d", id),
			Years: id * 2,
		}, ErrorsInStream{}
	}
}

// Сохраняем результат
type Result struct {
	Employee Employee `json:"employee"`
}

const workers = 3 // число потоков обработки

func main() {
	api := &MockAPI{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskCh := make(chan int, 10)                    // канал для задач
	resultCh := make(chan Result, 10)               // канал для результатов
	resultErrorsCh := make(chan ErrorsInStream, 10) // канал для ошибок
	var results []Employee
	var resultErrors []ErrorsInStream
	var wg sync.WaitGroup
	var mu sync.Mutex
	// как триггер отмены пред обработанных результатов
	//var once sync.Once

	employeesIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, id := range employeesIDs {
		if id <= 0 {
			fmt.Printf("Пропущен некорректный ID сотрудника: %d\n", id)
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
						//once.Do(cancel) // Анулируем успешные обработки задач
						return
					}
					// Расчет бонусов
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

	// Читаем из каналов результатов и ошибок
	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				resultCh = nil
			} else {
				mu.Lock()
				results = append(results, result.Employee)
				mu.Unlock()
				fmt.Printf("Результат: обработан сотрудник %s, бонус: %.2f\n", result.Employee.Name, result.Employee.Bonus)
			}
		case errResult, ok := <-resultErrorsCh:
			if !ok {
				resultErrorsCh = nil
			} else {
				mu.Lock()
				resultErrors = append(resultErrors, errResult)
				mu.Unlock()
				fmt.Printf("Ошибка: %v\n", errResult.Error)
			}
		}
		if resultCh == nil && resultErrorsCh == nil {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	// Выводим результаты и ошибки вместе
	fmt.Println("\nИтоговые результаты:")
	fmt.Println("Успешно обработанные сотрудники:", results)
	fmt.Println("Ошибки обработки:", resultErrors)
}
