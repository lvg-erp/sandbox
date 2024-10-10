package main

import (
	"fmt"
	"search2matrix/alg"
)

func main() {
	matrix := [][]int{{1, 3, 4}, {5, 7, 9}, {10, 12, 15}, {22, 26, 28}}
	target := 1
	res := alg.SearchInMatrix(matrix, target)
	fmt.Println(res)
}
