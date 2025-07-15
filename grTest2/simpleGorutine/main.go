package main

import (
	"fmt"
)

func main() {
	nums := []int{1, 2, 3, 4, 5}
	result := getSum(nums)

	fmt.Println(result)
}

func getSum(nums []int) int {

	//var wg sync.WaitGroup
	ch := make(chan int)
	//var result int
	sum := 0
	//wg.Add(1)
	go func() {
		//defer wg.Done()
		for _, num := range nums {
			sum += num
		}
		ch <- sum
	}()

	//go func() {
	//	wg.Wait()
	//	close(ch)
	//}()

	return <-ch

}
