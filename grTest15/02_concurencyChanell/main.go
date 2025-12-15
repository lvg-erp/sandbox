package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {

	const stream = 10
	var wg sync.WaitGroup
	ch := make(chan int)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < stream; i++ {
			ch <- genInt()
		}
		close(ch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for c := range ch {
			fmt.Println(c)
		}
	}()

	wg.Wait()

}

func genInt() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Используем r для генерации случайных чисел
	randomNumber := r.Intn(10)
	return randomNumber
}
