package main

import (
	"fmt"
	"stringworksix/alg"
)

func main() {

	words := []string{"abcw", "baz", "foo", "bar", "xtfn", "abcdef"}
	resultInt, resultString := alg.MaxProduct(words)

	fmt.Println(resultInt)
	fmt.Println(resultString)

}
