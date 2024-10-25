package main

import (
	"fmt"
	"superpow/alg"
)

func main() {
	//a, b := 2, []int{3}
	a, b := 2, 3
	result := alg.GetSum(a, b)

	fmt.Println(result)

}
