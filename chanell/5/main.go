package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// реализовать паралельное выполнение воркеров
// прокинуть контекст

type outVal struct {
	val int
	err error
}

var errTimeOut = errors.New("time out")

func processData(ctx context.Context, v int) chan outVal {
	ch := make(chan struct{})
	out := make(chan outVal)
	go func() {
		time.Sleep(time.Duration(rand.IntN(10)) * time.Second)
		close(ch)
	}()

	go func() {
		select {
		case <-ch:
			out <- outVal{
				val: v * 2,
				err: nil,
			}
		case <-ctx.Done():
			out <- outVal{
				val: 0,
				err: errTimeOut,
			}
		}
	}()

	return out

}

// вынесем воркер в отдельную функцию
func worker(ctx context.Context, in <-chan int, out chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	// пишем в горутине в канал

	//for v := range in {
	//	select {
	//	case out <- processData(ctx, v):
	//	case <-ctx.Done():
	//		return
	//	}
	//}

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-in:
			if !ok {
				return
			}
			select {
			case ov := <-processData(ctx, v):
				if ov.err != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- ov.val:
				}

			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// подсветим что один канал на чтение другой на запись
func processParallel(ctx context.Context, in <-chan int, out chan<- int, numWorkers int) {
	// синхронизация по вайт группам
	wg := &sync.WaitGroup{}
	// цикл по воркерам
	for range numWorkers {
		wg.Add(1)
		go worker(ctx, in, out, wg)
	}

	// ждем заверешения всех воркеров и закрываем канал
	go func() {
		wg.Wait()
		close(out)
	}()

}

func main() {

	in := make(chan int)
	out := make(chan int)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	go func() {
		defer close(in)
		for i := range 10 {
			select {
			case in <- i + 1:
			case <-ctx.Done():
				return
			}

		}
		//close(in)
	}()

	start := time.Now()
	processParallel(ctx, in, out, 5)

	for v := range out {
		fmt.Println("v = ", v)
	}

	fmt.Println("main duration: ", time.Since(start))

}
