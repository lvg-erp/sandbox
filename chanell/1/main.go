package main

// перелив данных из одной функции в другую
// называется pipeline

import (
	"fmt"
	"sync"
)

func writer() <-chan int64 {
	ch := make(chan int64)
	go func() {
		for i := int64(0); i < 10; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func double(in <-chan int64) <-chan int64 {
	ch := make(chan int64)
	go func() {
		for i := range in {
			ch <- i * 2
		}
		close(ch)
	}()

	return ch

}

func reader(in <-chan int64, wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range in {
		fmt.Println(v)
	}
}

func main() {
	//работает и без waitgroup
	var wg sync.WaitGroup
	wg.Add(1)
	go reader(double(writer()), &wg)
	wg.Wait()
}

//Решение через контекст
//func reader(ctx context.Context, in <-chan int64) {
//	for v := range in {
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			fmt.Println(v)
//		}
//	}
//}
//
//func main() {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	go reader(ctx, double(writer()))
//	// Дождаться завершения (например, через time.Sleep или другой сигнал)
//}
