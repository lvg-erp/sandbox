package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const workers = 9

func main() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ch := make(chan string, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			duration := time.Duration(r.Intn(30)+1) * time.Second
			select {
			case <-ctx.Done():
				ch <- fmt.Sprintf("Worker %d cancelled: %v", id, ctx.Err())
				return
			case <-time.After(duration):
				ch <- fmt.Sprintf("Worker %d completed in %v at %v", id, duration, time.Now().Format(time.RFC3339))
			}
		}(i)

		//go func(idx int) {
		//	defer wg.Done()
		//	r := rand.New(rand.NewSource(time.Now().UnixNano()))
		//	interval := r.Intn(10) + 1
		//	//ticker := time.NewTicker(time.Duration(interval) * time.Second)
		//	//defer ticker.Stop()
		//	//for {
		//	select {
		//	case <-ctx.Done():
		//		ch <- fmt.Sprintf("Worker %d cancelled: %v", i, ctx.Err())
		//		return
		//	case time.After(interval):
		//		//case <-ticker.C:
		//		//	select {
		//		//	case <-ctx.Done():
		//		//		//вывести не завершенную задачу
		//		//		ch <- fmt.Sprintf("Worker %d cancelled: %v", i, ctx.Err())
		//		//		return
		//		//	case ch <- itsWork():
		//		//		ch <- fmt.Sprintf("Worker %d completed ", i)
		//		//	}
		//	}
		//	//}
		//}(i)

	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	//tickerEnd := time.NewTicker(time.Second)
	//defer tickerEnd.Stop()

	//Этот цикл читает сообщения из канала ch и выводит их, но если срабатывает case <-ctx.Done() (то есть через 5 секунд после старта программы),
	//выполнение main немедленно завершается через return.
	//Проблема в том, что некоторые горутины могут отправлять сообщения в канал ch (например, сообщения об отмене) после срабатывания ctx.Done(),
	//но главный цикл уже завершился и не читает эти сообщения. Это приводит к тому, что сообщения об отмене не выводятся.
	//for {
	//	select {
	//	case <-ctx.Done():
	//		return
	//	//case <-tickerEnd.C:
	//	//	select {
	//	//	case <-ctx.Done():
	//	//		return
	//	case str := <-ch:
	//		fmt.Println(str)
	//	}
	//}

	for msg := range ch {
		fmt.Println(msg)
	}
}

func itsWork() string {
	return fmt.Sprintf("Processed events at %v", time.Now().Format(time.RFC3339))
}
