package main

import (
	"fmt"
	"painthouse/alg"
)

func main() {
	costs := [][]int{
		{17, 2, 17},
		{16, 16, 5},
		{14, 3, 19},
	}

	//fmt.Println(costs[0][0])
	//fmt.Println(costs[1][2])

	result := alg.MinCost(costs)

	fmt.Println(result)
}
