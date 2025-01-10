package main

import (
	"arrayrearrangement/alg"
	"fmt"
)

func main() {
	arrays := [][]int{{1, 2, 3}, {4, 5, 6, 7, 8}, {9, 10, 11}}
	fmt.Println(alg.Zip(arrays))
}
