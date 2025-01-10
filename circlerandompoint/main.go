package main

import (
	"circlerandompoint/alg"
	"fmt"
)

func main() {
	s := alg.NewSolution(1.0, 0.0, 0.0)
	fmt.Println(s.RandPoint())
}
