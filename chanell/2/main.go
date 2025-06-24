package main

import (
	"context"
	"fmt"
	"time"
)

func main() {

	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		ch1 <- 1 //Раблокировка канала
	}()

	time.Sleep(500 * time.Millisecond) // чтобы данные успели попасть в канал иначе deadlock

	timer := time.NewTimer(1 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()

	select { //блокирующий оператор
	case v := <-ch1:
		fmt.Println("v = ", v, "from ch1")
	case v := <-ch2:
		fmt.Println("v = ", v, "from ch2")
	case <-time.After(1 * time.Second): // Выход по таймауту через секунду
		fmt.Println("exited by after")
	case <-timer.C:
		fmt.Println("exited by timer") // Выход по таймеру
	case <-ctx.Done():
		fmt.Println("exited by context") // Выход по контексту
	default: // выход по дефолту
		fmt.Println("exit by default")
	}

}
