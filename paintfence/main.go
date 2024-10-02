package main

import (
	"fmt"
	"paintfence/alg"
)

func main() {
	n := 6
	k := 2

	result := alg.NumWays(n, k)

	fmt.Println(result)
}
