package main

import (
	"concurrence/6/pkg"
	"concurrence/6/pkg/waitg"
	"context"
	"fmt"
	"sync"
	"time"
)

// расширенная передача данных из одного канала в два и более каналов
// перетаскиваем данные для горутин которые будут писать в выходные каналы не дожидая когда данные заберут
// учитывая что каналы чтения работают с разной скоростью
// доработка предыдущего примера, добавляем в интерфейс функцию Add(int, int)
// инвертация зависимостей

func generateChan() chan int {
	ch := make(chan int)

	go func() {
		for i := range 5 {
			ch <- i
		}
		close(ch)
	}()

	return ch
}

func main() {
	teech := teechan.New(3, &waitg.WaitGNormal{}, &waitg.WaitGStub{}) //быстро
	//teech := teechan.New(3, &waitg.WaitGStub{}, &waitg.WaitGNormal{}) // медленно
	chans := teech.Execute(context.Background(), generateChan())

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		counter := 0
		for v := range chans[0] {
			counter++
			if counter == 1 {
				time.Sleep(time.Second)
			}
			fmt.Println("lodding...", v)
		}
	}()

	go func() {
		defer wg.Done()
		counter := 0
		for v := range chans[1] {
			counter++
			if counter == 2 {
				time.Sleep(2 * time.Second)
			}
			fmt.Println("writting metrics...", v)
		}
	}()

	go func() {
		defer wg.Done()
		counter := 0
		for v := range chans[2] {
			counter++
			if counter == 3 {
				time.Sleep(3 * time.Second)
			}
			fmt.Println("sending event to other service...", v)
		}
	}()

	wg.Wait()

}
