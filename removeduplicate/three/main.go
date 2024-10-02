package main

import (
	"fmt"
	"three/alg"
)

func main() {
	nms := []int{1, 2, 3, 1, 2, 3}
	k := 2
	result := alg.ContainsNearbyDuplicate(nms, k)

	fmt.Println(result)
}
