package main

import (
	"fmt"
	"removeduplicate/alg"
)

func main() {

	nums := []int{1, 1, 1, 2, 2, 2, 3}
	res, arr := alg.RemoveDuplicates(nums)
	fmt.Println(res)
	fmt.Println(arr)
}
