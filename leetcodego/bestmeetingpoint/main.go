package main

import (
	"bestmeetingpoint/alg"
	"fmt"
)

func main() {
	grid := [][]int{
		{1, 0, 0, 0, 1},
		{0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0},
	}

	result := alg.MinTotalDistance(grid)
	fmt.Println(result)
}
