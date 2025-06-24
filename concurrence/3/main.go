package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"time"
)

// randomWait -  это вызов долго работающей функции
// нужно вызвать конкурентно эту функцию например 100 раз
// но при этом что бы функция main отработала в пределах 5 сек
// т.е. мы запускаем randomWait 100 раз и получаем время её работы от 100 до 500 сек
// но майн должна работать не более 5 сек
// НЕ мой вариант 1 используем каналы

var maxWaitSeconds = 5

func randomWait() int {
	workSeconds := rand.Intn(5 + 1)
	time.Sleep(time.Duration(workSeconds) * time.Second)

	return workSeconds

}

func main() {
	totalWorkSeconds := 0
	ch := make(chan int)
	start := time.Now()

	for range 100 {
		go func() {
			seconds := randomWait()
			ch <- seconds
		}()
	}

	for range 100 {
		totalWorkSeconds += <-ch
	}

	mainSeconds := time.Since(start)

	fmt.Println("main", mainSeconds)
	fmt.Println("total", totalWorkSeconds)
}
