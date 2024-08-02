package main

import (
	"arrayworkwith/alg"
	"fmt"
)

func main() {
	arr1 := []string{"banana", "kiwi", "apple", "oranges"}
	arr2 := []string{"peer", "ananas", "apple", "oranges"}

	res := alg.MergeArrayWithoutDuplicate(arr1, arr2)

	fmt.Println(res)
}
