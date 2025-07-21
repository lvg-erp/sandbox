package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	rand_m "math/rand"
	"sync"
	"time"
)

func main() {
	ch := make(chan int)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- genRandP()
				time.Sleep(time.Duration(rand_m.Intn(2000)) * time.Millisecond)
			}
		}

	}()

	for {
		select {
		case num := <-ch:
			fmt.Println(num)
		case <-time.After(1 * time.Second):
			fmt.Println("Time out")
		case <-ctx.Done():
			fmt.Println("Program finished after 10 sec")
			wg.Wait()
			return
		}
	}

}

func genRandP() int {
	randomIndex, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		log.Fatalf("Failed to generate random number: %v", err)
	}
	return int(randomIndex.Int64())
}
