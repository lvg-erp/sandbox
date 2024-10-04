package main

import (
	"binarytreevertical/alg"
	"fmt"
)

func main() {
	data := []int{3, 9, 20, 0, 0, 15, 7}
	var root *alg.TreeNode
	root = alg.InsertLevelOrder(data, root, 0)

	result := alg.VerticalOrder(root)
	fmt.Println(result)

}
