package main

import (
	"fmt"
	"searchinarray/alg"
)

func main() {
	nums := []int{4, 5, 6, 6, 7, 0, 1, 2, 4, 4}
	res := alg.Search(nums, 2)
	fmt.Println(res)
}
