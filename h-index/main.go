package main

import (
	"fmt"
	"h-index/alg"
)

func main() {

	cArray := []int{3, 0, 2, 5, 7, 8, 1}
	result := alg.HIndex(cArray)
	fmt.Println(result)

}
