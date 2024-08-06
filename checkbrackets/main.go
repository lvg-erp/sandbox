package main

import (
	"checkbrackets/alg"
	"fmt"
)

func main() {
	testString := "{()[}"
	res := alg.IsValid(testString)
	fmt.Println(res)

}
