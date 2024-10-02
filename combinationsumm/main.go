package main

import (
	"combinationsumm/alg"
	"fmt"
)

func main() {
	k := 3
	n := 9

	//arr := [][]int{{1,2,6},{1,3,5},{2,3,4}}

	result := alg.CombinationSum(k, n)

	fmt.Println(result)
}
