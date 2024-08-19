package main

import (
	"calculator/alg"
	"fmt"
)

func main() {
	res := alg.Operation()

	fmt.Printf("Result operation: %.2f\n", res)

}
