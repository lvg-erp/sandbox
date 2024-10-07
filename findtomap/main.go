package main

import (
	"findtomap/alg"
	"fmt"
)

func main() {
	graph := [][]int{
		{1, 1, 0},
		{0, 1, 0},
		{1, 0, 1},
	}

	s := alg.Solution{
		N:     len(graph),
		Graph: graph,
	}
	result := s.FindCelebrity()
	fmt.Println(result) // Output will be the index of the celebrity or -1 if there is none.
}
