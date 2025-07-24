package main

import (
	"fmt"
	"math"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	ch := make(chan float64)
	inF := []float64{256.0, 384.0, 757.0, 64.0, 99.0, 893.0}
	var out []float64
	//обрабатываем каждое значение в отдельной горутине
	for i := 0; i < len(inF); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ch <- sqrtTest(inF[idx])
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for f := range ch {
		out = append(out, f)
	}

	fmt.Println(out)

}

func sqrtTest(in float64) float64 {
	return math.Sqrt(in)
}
