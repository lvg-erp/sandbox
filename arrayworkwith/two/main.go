package main

import (
	"fmt"
	"zero2end/alg"
)

func main() {
	nums := []int{0, 1, 0, 3, 12}

	alg.MoveZero2End(nums)

	fmt.Println(nums)

}
