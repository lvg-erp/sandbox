package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func main() {

	ch := make(chan int)
	ch1 := make(chan int)
	ch2 := make(chan int)
	var wg sync.WaitGroup

	// Генерация чисел 1–10
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch)
		for i := 1; i <= 10; i++ {
			ch <- genInt()
		}
	}()

	// Умножение на 2
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch1)
		for num := range ch {
			ch1 <- multiply(num)
		}
	}()

	// Фильтрация четных чисел
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch2)
		for num := range ch1 {
			if num%2 == 0 {
				ch2 <- num
			}
		}
	}()

	for num := range ch2 {
		fmt.Printf("Result: %d\n", num)
	}

	wg.Wait()

}

func genInt() int {
	return rand.Intn(10)
}

func multiply(in int) int {
	return in * in
}

//func multiply() int64 {
//	r, err := rand.Int(rand.Reader, big.NewInt(10))
//	if err != nil {
//		log.Fatal("")
//	}
//	return r.Int64() * r.Int64()
//}

//func addition(in int64) int64 {
//	return in + in
//}
//
//func evenOnly(in int64) int64 {
//	if in%2 == 0 {
//		return in
//	}
//
//	return -1
//}
