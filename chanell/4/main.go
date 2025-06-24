package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// Имеется/ функция которая работает неопределенно долго
func randomTimeWork() {
	time.Sleep(time.Duration(rand.IntN(100)) * time.Second)
}

// написать функцию которая прерывет выполнение randomTimeWork через 3 сек

func predictableTimeWork() error {

	done := make(chan struct{})

	go func() {
		defer close(done)
		randomTimeWork()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("randomTimeWork timed out after 3 seconds")
	}
}

func main() {
	err := predictableTimeWork()
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("randomTimeWork завершилась успешно")
}
