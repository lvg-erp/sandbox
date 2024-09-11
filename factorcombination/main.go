package main

import (
	"faactorcombination/alg"
	"fmt"
)

func main() {

	numericInput := 9
	result := alg.GetFactors(numericInput)

	fmt.Println(result)

}
