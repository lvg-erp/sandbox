package main

import (
	"fmt"
	"listone/alg"
)

func main() {
	nestedList := [][][]int{
		{
			{1, 2, 3},
		},
		{},
		{},
		{},
		{},
		{},
	}

	// Convert to linked list
	linkedList := alg.NestedListToLinkedList(nestedList)
	solution := alg.NewSolution(linkedList)
	res := solution.GetRandom()
	fmt.Println(res)
}
