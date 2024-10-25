package main

import (
	"fmt"
	"matrix/alg"
)

func main() {
	matrix := [][]string{
		{"1", "0", "1", "0", "0"},
		{"1", "0", "1", "1", "1"},
		{"1", "1", "1", "1", "1"},
		{"1", "0", "0", "1", "0"},
	}

	resByte := alg.StringToByte(matrix)

	res := alg.MaximalSquare(resByte)

	fmt.Println(res)

}
