package main

import (
	"fmt"
	"sync"
)

//Напишите программу, которая принимает слайс чисел ([]int) и вычисляет их сумму,
//разделяя работу между несколькими горутинами.
//Используйте каналы для передачи частичных сумм и sync.WaitGroup для синхронизации.
//Верните общую сумму.

func parallelSum(in []int, count int) int {
	wg := sync.WaitGroup{}
	ch := make(chan int)

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			chunkSize := len(in) / count
			start := i * chunkSize
			fmt.Println("this start", start, "-", i)
			end := start + chunkSize
			fmt.Println("this end", end, "-", i)
			if i == count-1 {
				end = len(in)
			}
			sum := 0
			for j := start; j < end; j++ {
				sum += in[j]
			}
			fmt.Println("sum ", sum)
			ch <- sum
			wg.Done()
		}(i)
	}

	//go func() {
	//	s1 := 0
	//	for i := 0; i < len(in); i++ {
	//		if i%2 == 0 {
	//			s1 = s1 + in[i]
	//		}
	//	}
	//	ch <- s1
	//	wg.Done()
	//}()
	//
	//go func() {
	//	s2 := 0
	//	for i := 0; i < len(in); i++ {
	//		if i%2 == 1 {
	//			s2 = s2 + in[i]
	//		}
	//	}
	//	ch <- s2
	//	wg.Done()
	//}()

	result := 0
	for i := 0; i < 2; i++ {
		result += <-ch
	}

	wg.Wait()

	close(ch)

	return result
}

func main() {

	input := []int{1, 2, 3, 4, 5}
	result := parallelSum(input, 2) // 2 горутины
	fmt.Println(result)             // Должно вывести: 15

}
