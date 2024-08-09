package main

import (
	"fmt"
	"sequence/alg"
)

func main() {
	input := []int{100, 4, 200, 1, 2, 3}
	res := alg.LongestConsecutive(input)

	fmt.Println(res)
}
