package main

import (
	"binarytreethree/alg"
	"fmt"
)

func main() {

	data := []int{3, 9, 20, 18, 33, 15, 7}
	var root *alg.TreeNode
	root = alg.CreateTreeFormArray(data, root, 0)
	fmt.Println(root)
	sum, path := alg.MaxPathSum(root)

	fmt.Printf("Max summa: %d\n", sum)
	fmt.Printf("Path: %v\n", path)

}
