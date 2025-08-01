package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	a       = 100
	workers = 5
)

func main() {

	var wg sync.WaitGroup
	ch := make(chan string, workers)
	chErr := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var once sync.Once // Для однократного вызова cancel
	var hasError bool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			b := r.Intn(10)

			select {
			case <-ctx.Done():
				//ch <- fmt.Sprintf("Proccess %d cancelled: %v", idx, ctx.Err())
				return
			default:
				result, err := divide(a, b)
				if err != nil {
					chErr <- formatResult(idx, a, b, result, err)
					hasError = true
					once.Do(cancel)
					//cancel()
					return
				}
				ch <- formatResult(idx, a, b, result, err)
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(chErr)
		close(ch)
	}()

	for {
		if hasError {
			// Если была ошибка, читаем только из chErr
			select {
			case msg, ok := <-chErr:
				if !ok {
					return
				}
				fmt.Println("Ошибка:", msg)
			}
		} else {
			select {
			case msg, ok := <-chErr:
				if !ok {
					return
				}
				fmt.Println("Ошибка:", msg)
				hasError = true
			case msg, ok := <-ch:
				if !ok {
					if len(chErr) == 0 {
						return
					}
					continue
				}
				fmt.Println("Результат:", msg)
			}
		}
	}

	//for ch != nil || chErr != nil {
	//	select {
	//	case msg, ok := <-ch:
	//		if !ok {
	//			ch = nil // Канал закрыт
	//			continue
	//		}
	//		fmt.Println(msg)
	//	case msg, ok := <-chErr:
	//		if !ok {
	//			chErr = nil // Канал закрыт
	//			continue
	//		}
	//		fmt.Println(msg)
	//	}
	//}

	//if len(chErr) > 0 {
	//	for e := range chErr {
	//		fmt.Println(e)
	//	}
	//} else {
	//	for r := range ch {
	//		fmt.Println(r)
	//	}
	//}

}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func formatResult(idx, a, b, result int, err error) string {
	if err != nil {
		return fmt.Sprintf("Worker %d: Ошибка для %d/%d: %v", idx, a, b, err)
	}
	return fmt.Sprintf("Worker %d: %d/%d = %d", idx, a, b, result)
}
