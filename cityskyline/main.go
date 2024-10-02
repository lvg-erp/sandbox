package main

import (
	"fmt"
	"skyline/alg"
)

func main() {

	buildings := [][]int{{2, 9, 10}, {3, 7, 15}, {5, 12, 12}, {15, 20, 10}, {19, 24, 8}}

	result := alg.GetSkyline(buildings)

	fmt.Println(result)

}
