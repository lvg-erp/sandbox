package main

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"
)

const workers = 5

type Result struct {
	Input  int
	Output int64
}

type Error struct {
	WorkerID int
	Error    error
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*100)
	defer cancel()

	in := []int{1, 2, 3, 4, 1001, 500, 8, 19}

	var wg sync.WaitGroup
	//var mu sync.Mutex

	r, e, err := processFactorials(ctx, in, workers, &wg)
	if err != nil {
		fmt.Println("Error parsing")
	}

	fmt.Println("\nИтоговые результаты:")
	fmt.Println("Успешно обработанные:", r)
	fmt.Println("Ошибки обработки:", e)

}

func processFactorials(ctx context.Context, numbers []int, workers int, wg *sync.WaitGroup) ([]Result, []Error, error) {

	var (
		results    []Result
		errResults []Error
	)

	resultCh := make(chan Result, len(numbers))
	errCh := make(chan Error, len(numbers)) //????в данном случае лучше не указывать длину канал, но если не будет буфера, на больших числах паника

	if len(numbers) == 0 {
		return results, errResults, fmt.Errorf("numbers is empty")
	}

	if workers == 0 {
		workers = 1
	}

	if workers > len(numbers) {
		workers = len(numbers)
	}

	chunkSize := (len(numbers) + workers - 1) / workers

	for i := 0; i < workers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := start + chunkSize
		if start >= len(numbers) {
			wg.Done() //
			continue
		}
		if end > len(numbers) {
			end = len(numbers)
		}
		go func(chunk []int, workerID int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				//это не верно
				//errResults = append(errResults, Error{
				//	WorkerID: workerID,
				//	Error:    fmt.Errorf("timeout for task %d", i),
				//})
				errCh <- Error{
					WorkerID: workerID,
					Error:    ctx.Err(),
				}
			default:
				for _, e := range chunk {
					if e < 0 {
						errCh <- Error{
							WorkerID: workerID,
							Error:    fmt.Errorf("negative number %d", e),
						}
						continue
					}
					if e > 1000 {
						errCh <- Error{
							WorkerID: workerID,
							Error:    fmt.Errorf("number %d too large", e),
						}
						continue
					}
					f, err := factorialInt(e, ctx)
					if err != nil {
						errCh <- Error{
							WorkerID: workerID,
							Error:    err,
						}
						continue
					}

					resultCh <- Result{
						Input:  e,
						Output: f.Int64(),
					}
				}
			}

		}(numbers[start:end], i)

	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	for rc := range resultCh {
		results = append(results, rc)
	}

	for ec := range errCh {
		errResults = append(errResults, ec)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Input < results[j].Input
	})

	sort.Slice(errResults, func(i, j int) bool {
		return errResults[i].WorkerID < errResults[j].WorkerID
	})

	return results, errResults, nil
}

func factorialInt(n int, ctx context.Context) (*big.Int, error) {
	if n < 0 {
		return nil, fmt.Errorf("negative number %d", n)
	}
	result := big.NewInt(1)
	for i := 2; i <= n; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			result.Mul(result, big.NewInt(int64(i)))
		}
	}
	if !result.IsInt64() {
		return nil, fmt.Errorf("factorial %d too large for int64", n)
	}
	return result, nil
}
