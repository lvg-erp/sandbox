package main

import (
	"binarytree/alg"
	"fmt"
	tn "github.com/egregors/TreeNode"
	"log"
)

func main() {
	//Input: root = [1,2,3,null,5]
	//"[1,2,3,null,null,4,5]"
	//"[1,2,3,4,6,5,7]"
	//Output: ["1->2->5","1->3"]

	data := "[1,2,3,null,null,4,5]"
	//из пакета гита
	//чтобы использовать сиой пакет меняем путь
	root, err := tn.NewTreeNode(data)
	if err != nil {
		log.Fatal(err)
	}
	res := alg.BinaryTreePaths(root)
	fmt.Println(res)

}
