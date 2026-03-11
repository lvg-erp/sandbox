package main

import (
	"github.com/davecgh/go-spew/spew"
	"math/big"
	"math/rand"
	"sync"
)

const stream = 8

type Results []Result

type Result struct {
	Original  int
	Square    int64
	Factorial uint64 // только если |n| ≤ 12
	IsPrime   bool
}

// генерация массива
func gen() []int {
	arr := make([]int, 1000)
	for i := range arr {
		arr[i] = rand.Intn(1001) - 500
	}

	return arr
}

// факториал числа
func factorialInt(n int) *uint64 {

	var result uint64

	if n < 0 {
		n = -n
	}

	if n < 12 {
		result := big.NewInt(1)
		for i := 2; i <= n; i++ {
			result.Mul(result, big.NewInt(int64(i)))
		}
	}

	return &result
}

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	//for i := 5; i*i <= n; i += 6 {
	//	if n%i == 0 || n%(i+2) == 0 {
	//		return false
	//	}
	//}
	return true
}

func main() {

	arrNum := gen()

	rr := proccess(arrNum)

	spew.Dump(rr)
}

func proccess(in []int) Results {
	var (
		wg sync.WaitGroup
		rs Results
	)

	// разобьем массив на части для паралельной работы
	chunkSize := (len(in) + stream - 1) / stream
	//канал результатов
	ch := make(chan Result, len(in))

	for i := 0; i < stream; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(in) {
			end = len(in)
		}
		s, e := start, end
		if start < len(in) {
			wg.Go(func() {
				chunk := in[s:e]
				for e := 0; e < len(chunk); e++ {

					r := Result{
						Original:  chunk[e],
						Factorial: *factorialInt(chunk[e]),
						IsPrime:   isPrime(chunk[e]),
						Square:    int64(chunk[e] * chunk[e]),
					}
					ch <- r
				}
			})
		}

	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for f := range ch {
		rs = append(rs, f)
	}

	return rs
}
