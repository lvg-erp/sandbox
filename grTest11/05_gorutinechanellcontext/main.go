package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 15, 18, 20, 29, 30}
	result, err := Pipeline(ctx, input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Sums: %v\n", result)
}

func Pipeline(ctx context.Context, in []int) ([]int, error) {

	var (
		wg         sync.WaitGroup
		chanChet   = make(chan int, 10)
		chanDouble = make(chan int, 10)
		chanSums   = make(chan int, 10)
		sums       []int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ChetChanInt(in, chanChet, ctx)
		close(chanChet)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		DoubleIntToChan(chanChet, chanDouble, ctx)
		close(chanDouble)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chanSums)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		sum := 0
		for {
			select {
			case <-ctx.Done():
				if sum > 0 {
					chanSums <- sum
				}
				return
			case n, ok := <-chanDouble:
				if !ok {
					if sum > 0 {
						chanSums <- sum
					}
					return
				}
				sum += n
			case <-ticker.C:
				if sum > 0 {
					chanSums <- sum
					sum = 0
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case s, ok := <-chanSums:
				if !ok {
					return
				}
				sums = append(sums, s)
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	return sums, ctx.Err()

}

func ChetChanInt(input []int, chet chan<- int, ctx context.Context) {

	for _, e := range input {
		select {
		case <-ctx.Done():
			return
		default:
			if e%2 == 0 {
				chet <- e
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

}

func DoubleIntToChan(chet <-chan int, resDouble chan<- int, ctx context.Context) {

	for c := range chet {
		select {
		case <-ctx.Done():
			return
		default:
			resDouble <- c * 2
		}
	}
}

//func DoubleToSqrtChan(double <-chan int, resSqrt chan<- int) {
//	for c := range double {
//		resSqrt <- c * c
//	}
//}
