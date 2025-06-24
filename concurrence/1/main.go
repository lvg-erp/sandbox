package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"sync"
	"sync/atomic"
	"time"
)

// randomWait -  это вызов долго работающей функции
// нужно вызвать конкурентно эту функцию например 100 раз
// но при этом что бы функция main отработала в пределах 5 сек
// т.е. мы запускаем randomWait 100 раз и получаем время её работы от 100 до 500 сек
// но майн должна работать не более 5 сек
// мой вариант 1

var maxWaitSeconds = 5

func randomWait() int {
	workSeconds := rand.Intn(5 + 1)
	time.Sleep(time.Duration(workSeconds) * time.Second)

	return workSeconds

}

func main() {

	var wg sync.WaitGroup

	//mainSeconds := 0
	var totalWorkSeconds int64
	start := time.Now()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workSeconds := randomWait()
			atomic.AddInt64(&totalWorkSeconds, int64(workSeconds))
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Duration(maxWaitSeconds) * time.Second):
	}
	mainSeconds := int(time.Since(start).Seconds())

	fmt.Println("main", mainSeconds)
	fmt.Println("total", totalWorkSeconds)
}
