package main

import (
	"fmt"
	"matrixtwo/alg"
)

func main() {
	matrix := [][]int{{9, 9, 4}, {6, 6, 8}, {2, 1, 1}}

	result := alg.LongestInscreasingPath(matrix)

	fmt.Print(result)

}
