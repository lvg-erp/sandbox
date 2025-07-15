package main

import (
	"fmt"
	"strings"
	"sync"
)

func main() {
	ins := []string{"hello", "world", "golang", "is", "awesome"}
	result := processString(ins, 3)
	fmt.Println(result)
}

func processString(ins []string, worker int) []string {
	if worker <= 0 {
		worker = 1
	}
	var result []string
	// расчитаем по сколько размерность данных для кажой горутины
	chunkSize := (len(ins) + worker - 1) / worker
	ch := make(chan string, len(ins))
	// бежим по количеству воркеров создаем горутины
	var wg sync.WaitGroup
	for i := 0; i < worker; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(ins) {
			end = len(ins)
		}
		if start < len(ins) {
			wg.Add(1)
			go func(chunk []string) {
				defer wg.Done()
				for _, c := range chunk {
					//result = append(result, toUpper(c))
					ch <- toUpper(c)
				}
			}(ins[start:end])
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		result = append(result, r)
	}

	return result

}

func toUpper(in string) string {
	return strings.ToUpper(in)
}
