package main

import (
	"fmt"
	"numberislands/alg"
)

func main() {
	matrix := [][]byte{
		{'1', '1', '0', '0'},
		{'1', '1', '0', '0'},
	}

	res := alg.SearchEarth(matrix)
	fmt.Println(res)

}
