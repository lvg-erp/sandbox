package main

import (
	"fmt"
	"stringworkfour/alg"
)

func main() {

	s := "adfgjkrt"
	a := []int{1, 3, 5, 8}

	res := alg.Capitalize(s, a)
	fmt.Println(res)
}
